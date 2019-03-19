package updater

import (
	. "github.com/cbrand/gocheck_matchers"
	. "gopkg.in/check.v1"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type UpdatePlanerSuite struct {
	config       *Config
	updatePlaner *UpdatePlaner
	deployments  []v1.Deployment
	jobs         []batchv1.Job
}

var _ = Suite(&UpdatePlanerSuite{})

func (suite *UpdatePlanerSuite) SetUpSuite(c *C) {
	fakeAPI := NewFakeKubernetesAPI()
	suite.config = NewConfig(fakeAPI.Client, NewImage("xcnt/test:1.0.0"), "stable")
	deployment := GetDeploymentDefaultAnnotation("xcnt/test2:latest", "xcnt/test:0.9.9", "xcnt/tmp:1.0.0")
	deployment.Spec.Template.Spec.InitContainers = []apiv1.Container{GetContainerWith("xcnt/test:0.9.9")}
	suite.deployments = []v1.Deployment{deployment}
	job := GetJobDefaultAnnotation("xcnt/test2:latest", "xcnt/test:0.9.9", "xcnt/tmp:1.0.0")
	job.UID = "something"
	job.SelfLink = "something"
	job.Spec.Template.Spec.InitContainers = []apiv1.Container{GetContainerWith("xcnt/test:0.9.9")}
	suite.jobs = []batchv1.Job{job}
	suite.updatePlaner = &UpdatePlaner{
		JobLister:        func() []batchv1.Job { return suite.jobs },
		DeploymentLister: func() []v1.Deployment { return suite.deployments },
	}
	suite.deployments = append(suite.deployments)
}

func (suite *UpdatePlanerSuite) GetVerifiedDeployment(c *C) v1.Deployment {
	updatePlan := suite.updatePlaner.Plan(suite.config)
	deployments := updatePlan.GetToApplyDeployments()
	c.Assert(len(deployments), Equals, 1)
	return deployments[0]
}

func (suite *UpdatePlanerSuite) TestPlanDeploymentContainers(c *C) {
	deployment := suite.GetVerifiedDeployment(c)
	// Check that the update classifier hasn't been touched
	_, ok := deployment.Annotations[UpdateClassifier]
	c.Assert(ok, Equals, true)
	containers := deployment.Spec.Template.Spec.Containers
	suite.verifyContainers(c, containers)
}

func (suite *UpdatePlanerSuite) verifyContainers(c *C, containers []apiv1.Container) {
	c.Assert(len(containers), Equals, 3)
	c.Assert(containers[0].Image, Equals, "xcnt/test2:latest")
	c.Assert(containers[1].Image, Equals, "xcnt/test:1.0.0")
	c.Assert(containers[2].Image, Equals, "xcnt/tmp:1.0.0")
}

func (suite *UpdatePlanerSuite) TestPlanInitContainers(c *C) {
	deployment := suite.GetVerifiedDeployment(c)
	initContainers := deployment.Spec.Template.Spec.InitContainers
	c.Assert(len(initContainers), Equals, 1)
	c.Assert(initContainers[0].Image, Equals, "xcnt/test:1.0.0")
}

func (suite *UpdatePlanerSuite) verfiyPlanedJob(c *C) batchv1.Job {
	updatePlan := suite.updatePlaner.Plan(suite.config)
	jobs := updatePlan.GetToCreateJobs()
	c.Assert(len(jobs), Equals, 1)
	return jobs[0]
}

func (suite *UpdatePlanerSuite) TestPlanJob(c *C) {
	job := suite.verfiyPlanedJob(c)
	c.Assert(job.UID, Equals, types.UID(""))
	c.Assert(job.SelfLink, Equals, "")
	_, ok := job.Annotations[UpdateClassifier]
	c.Assert(ok, Equals, false)
	suite.verifyContainers(c, job.Spec.Template.Spec.Containers)
}

func (suite *UpdatePlanerSuite) TestPlanJobInitContainer(c *C) {
	job := suite.verfiyPlanedJob(c)
	initContainers := job.Spec.Template.Spec.InitContainers
	c.Assert(len(initContainers), Equals, 1)
	c.Assert(initContainers[0].Image, Equals, "xcnt/test:1.0.0")
}

func (suite *UpdatePlanerSuite) TestPlanJobWithPreExistingLabels(c *C) {
	job := suite.jobs[0]
	job.Spec.Template.ObjectMeta.SetLabels(map[string]string{
		controllerUUIDLabel: "abc",
	})
	suite.jobs[0] = job
	job = suite.verfiyPlanedJob(c)
	_, ok := job.Spec.Template.ObjectMeta.GetLabels()[controllerUUIDLabel]
	c.Assert(ok, IsFalse)
}

func (suite *UpdatePlanerSuite) TestResourceVersionRemoval(c *C) {
	job := suite.jobs[0]
	job.ResourceVersion = "test"
	suite.jobs[0] = job
	job = suite.verfiyPlanedJob(c)
	c.Assert(job.ResourceVersion, Equals, "")
}

func (suite *UpdatePlanerSuite) TestPlanFunction(c *C) {
	plan, err := Plan(suite.config)
	c.Assert(err, IsNil)
	c.Assert(len(plan.GetToApplyDeployments()), Equals, 0)
	c.Assert(len(plan.GetToCreateJobs()), Equals, 0)
}
