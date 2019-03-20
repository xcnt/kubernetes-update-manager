package updater

import (
	"strconv"
	"time"

	. "gopkg.in/check.v1"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdaterSuite struct {
	config           *Config
	kubernetesAPI    KubernetesAPI
	imageName        string
	updateDeployment *v1.Deployment
	updateJob        *batchv1.Job
	updatePlan       UpdatePlan
}

var _ = Suite(&UpdaterSuite{})

func (suite *UpdaterSuite) SetUpTest(c *C) {
	suite.imageName = "xcnt/test:1.0.0"
	suite.kubernetesAPI = NewFakeKubernetesAPI()
	suite.kubernetesAPI.NewNamespace("default")
	suite.config = NewConfig(suite.kubernetesAPI.Client, NewImage(suite.imageName), "default")
	suite.config.SetNamespaces([]string{"default"})

	deployment := GetDeploymentDefaultAnnotation("xcnt/test:0.9.9")
	suite.kubernetesAPI.NewDeploymentIn("default", deployment)
	updateDeployment := v1.Deployment{}
	deployment.DeepCopyInto(&updateDeployment)
	updateDeployment.Spec.Template.Spec.Containers[0].Image = suite.imageName
	suite.updateDeployment = &updateDeployment

	job := GetJobDefaultAnnotation(suite.imageName)
	suite.updateJob = &job

	suite.updatePlan = &updatePlan{
		deployments: []v1.Deployment{updateDeployment},
		jobs:        []batchv1.Job{job},
	}
}

func (suite *UpdaterSuite) TestUpdaterFailure(c *C) {
	progress := Update(suite.updatePlan, suite.config)
	c.Assert(progress.Finished(), Equals, false)
	updateJob := suite.updateJob
	job := &batchv1.Job{}
	updateJob.DeepCopyInto(job)
	job.Status.Failed++
	c.Assert(int(job.Status.Failed), Equals, 1)
	time.Sleep(100 * time.Millisecond)
	suite.kubernetesAPI.UpdateJobIn("default", job)
	suite.waitForFinish(progress)
	c.Assert(progress.FinishTime(), NotNil)
	c.Assert(progress.Finished(), Equals, true)
	c.Assert(progress.Successful(), Equals, false)
	c.Assert(progress.Failed(), Equals, true)
}

func (suite *UpdaterSuite) TestUpdateSuccess(c *C) {
	progress := Update(suite.updatePlan, suite.config)
	c.Assert(progress.Finished(), Equals, false)
	updateJob := suite.updateJob
	job := &batchv1.Job{}
	updateJob.DeepCopyInto(job)
	job.Status.Succeeded++
	time.Sleep(100 * time.Millisecond)
	c.Assert(int(job.Status.Succeeded), Equals, 1)
	suite.kubernetesAPI.UpdateJobIn("default", job)
	suite.waitForJobToFinish(job)
	suite.waitForJobCountToBe(1, progress)
	c.Assert(progress.Finished(), Equals, false)
	c.Assert(progress.Successful(), Equals, false)
	c.Assert(progress.Failed(), Equals, false)
	c.Assert(progress.FinishedJobsCount(), Equals, 1)
	c.Assert(progress.UpdatedDeploymentsCount(), Equals, 0)

	deployment := &v1.Deployment{}
	suite.updateDeployment.DeepCopyInto(deployment)
	deployment.Status.ReadyReplicas = 1
	deployment.Status.UpdatedReplicas = 1
	suite.kubernetesAPI.UpdateDeploymentIn("default", deployment)
	suite.waitForDeploymentToFinish(deployment)
	suite.waitForDeploymentCountToBe(1, progress)
	c.Assert(progress.Finished(), Equals, true)
	c.Assert(progress.Successful(), Equals, true)
	c.Assert(progress.Failed(), Equals, false)
	c.Assert(progress.FinishedJobsCount(), Equals, 1)
	c.Assert(progress.UpdatedDeploymentsCount(), Equals, 1)
}

func (suite *UpdaterSuite) TestFailureWithReplicaSets(c *C) {
	deployment := suite.updateDeployment
	previousRS := GetReplicaSetFor(deployment)
	suite.kubernetesAPI.NewReplicaSetIn("default", previousRS)
	lastRS := GetReplicaSetFor(deployment)
	generation, ok := lastRS.Annotations[ReplicaSetRevisionAnnotation]
	c.Assert(ok, Equals, true)
	data, err := strconv.Atoi(generation)
	c.Assert(err, IsNil)
	deployment.Generation = int64(data)
	suite.kubernetesAPI.UpdateDeploymentIn("default", deployment)
	suite.kubernetesAPI.NewReplicaSetIn("default", lastRS)

	suite.updatePlan.GetToApplyDeployments()[0] = *deployment
	progress := Update(suite.updatePlan, suite.config)
	c.Assert(progress.Finished(), Equals, false)
	updateJob := suite.updateJob
	job := &batchv1.Job{}
	updateJob.DeepCopyInto(job)
	job.Status.Failed++
	c.Assert(int(job.Status.Failed), Equals, 1)
	time.Sleep(300 * time.Millisecond)
	suite.kubernetesAPI.UpdateJobIn("default", job)
	suite.waitForFinish(progress)

	time.Sleep(300 * time.Millisecond)
	retrievedDeployment, err := suite.config.GetDeploymentAPIFor("default").Get(deployment.Name, metaV1.GetOptions{})
	c.Assert(err, IsNil)
	c.Assert(retrievedDeployment.Spec.Template.Spec.Containers[0].Name, Equals, previousRS.Spec.Template.Spec.Containers[0].Name)
}

func (suite *UpdaterSuite) TestUpdateAbort(c *C) {
	progress := Update(suite.updatePlan, suite.config)
	c.Assert(progress.Finished(), Equals, false)
	updateJob := suite.updateJob
	job := &batchv1.Job{}
	updateJob.DeepCopyInto(job)
	job.Status.Succeeded++
	time.Sleep(100 * time.Millisecond)
	c.Assert(int(job.Status.Succeeded), Equals, 1)
	suite.kubernetesAPI.UpdateJobIn("default", job)
	suite.waitForJobToFinish(job)
	suite.waitForJobCountToBe(1, progress)
	c.Assert(progress.Finished(), Equals, false)
	c.Assert(progress.Successful(), Equals, false)
	c.Assert(progress.Failed(), Equals, false)
	c.Assert(progress.FinishedJobsCount(), Equals, 1)
	c.Assert(progress.UpdatedDeploymentsCount(), Equals, 0)

	progress.Abort()
	time.Sleep(200 * time.Millisecond)
	deployment := &v1.Deployment{}
	suite.updateDeployment.DeepCopyInto(deployment)
	deployment.Status.ReadyReplicas = 1
	deployment.Status.UpdatedReplicas = 1
	suite.kubernetesAPI.UpdateDeploymentIn("default", deployment)
	suite.waitForDeploymentToFinish(deployment)
	suite.waitForDeploymentCountToBe(1, progress)
	c.Assert(progress.Finished(), Equals, true)
	c.Assert(progress.Failed(), Equals, true)
	c.Assert(progress.FinishedJobsCount(), Equals, 1)
	c.Assert(progress.UpdatedDeploymentsCount(), Equals, 0)
}

func (suite *UpdaterSuite) waitForJobToFinish(job *batchv1.Job) {
	job, _ = suite.config.GetJobAPIFor(job.Namespace).Get(job.Name, metaV1.GetOptions{})
	for i := 0; i < 20 && job.Status.Succeeded == 0; i++ {
		time.Sleep(100 * time.Millisecond)
		job, _ = suite.config.GetJobAPIFor(job.Namespace).Get(job.Name, metaV1.GetOptions{})
	}
}

func (suite *UpdaterSuite) waitForDeploymentToFinish(deployment *v1.Deployment) {
	deployment, _ = suite.config.GetDeploymentAPIFor(deployment.Namespace).Get(deployment.Name, metaV1.GetOptions{})
	for i := 0; i < 20 && !isDeploymentFinished(deployment); i++ {
		time.Sleep(100 * time.Millisecond)
		deployment, _ = suite.config.GetDeploymentAPIFor(deployment.Namespace).Get(deployment.Name, metaV1.GetOptions{})
	}
}

func (suite *UpdaterSuite) waitForDeploymentCountToBe(target int, status UpdateProgress) {
	for i := 0; i < 50 && status.UpdatedDeploymentsCount() < target; i++ {
		time.Sleep(100 * time.Millisecond)
	}
}

func (suite *UpdaterSuite) waitForJobCountToBe(target int, status UpdateProgress) {
	for i := 0; i < 50 && status.FinishedJobsCount() < target; i++ {
		time.Sleep(100 * time.Millisecond)
	}
}

func (suite *UpdaterSuite) waitForFinish(status UpdateProgress) {
	for i := 0; i < 50 && !status.Finished(); i++ {
		time.Sleep(100 * time.Millisecond)
	}
}
