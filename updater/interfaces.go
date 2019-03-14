package updater

import (
	"time"

	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	appsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	batchV1Interface "k8s.io/client-go/kubernetes/typed/batch/v1"
)

// MatchConfig interface includes functions needed to be provided to
// match the cofigurations
type MatchConfig interface {
	// GetImage returns the image configuration for the update
	GetImage() *Image
	// GetUpdateClassifier returns the update classifier which is
	// applied for the configuration
	GetUpdateClassifier() string
}

// UpdatePlan implements configuration which can be applied to handle an update of a docker image.
type UpdatePlan interface {
	// GetToCreateJobs returns a slice of jobs which should be created for the deployments to run.
	GetToCreateJobs() []batchv1.Job
	// GetToApplyDeployments returns a slice of deployments which are the deployment configurations needed to be applied to the cluster for the update
	// to run through
	GetToApplyDeployments() []v1.Deployment
}

// KubernetesWrapper includes functionality which needs to be implemented for returning the job interface.
type KubernetesWrapper interface {
	// GetJobAPIFor returns the clientset's job interface
	GetJobAPIFor(namespace string) batchV1Interface.JobInterface
	// GetDeploymentAPIFor returns the clientset's specified deployment api for the configuration
	GetDeploymentAPIFor(namespace string) appsV1.DeploymentInterface
	// GetReplicaSetAPIFor returns the clientset's specified replicaset api for the configuration
	GetReplicaSetAPIFor(namespace string) appsV1.ReplicaSetInterface
}

// UpdateProgress interface can be used to query status of current upgrade processes.
type UpdateProgress interface {
	// GetJobs returns a list of jobs which are included in the update progress
	GetJobs() []*batchv1.Job
	// GetDeployments returns the list of deployments which needs to be updated
	GetDeployments() []*v1.Deployment
	// FinishedJobsCount returns how many jobs have been finished
	FinishedJobsCount() int
	// UpdatedDeploymentsCount returns the amount of deployments which update has been finished
	UpdatedDeploymentsCount() int
	// FinishTime returns when the progress was finished. If the update hasn't finished yet, this will return nil
	FinishTime() *time.Time
	// Finished returns if the update progress has run through succesfully or unsuccessfully
	Finished() bool
	// Failed returns if the update is marked as failed
	Failed() bool
	// Successful returns true if the complete update progress has run through
	Successful() bool
	// Abort cancels the run of this specific udpater.
	Abort()
}
