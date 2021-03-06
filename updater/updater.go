package updater

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// ErrNoReplicaSet is returned, if a deployment on a rollback call doesn't have any replica sets attached
	ErrNoReplicaSet = errors.New("The deployment does not have replica sets attached")
	// ErrPreviousReplicaSetNotFound describes the error that on a rollback the previous replicaset hasn't been identified
	ErrPreviousReplicaSetNotFound = errors.New("The replicaset before the current one wasn't found")
)

// updateProgressConfiguration includes checking the status.
type updateProgressConfiguration struct {
	jobs        []*batchv1.Job
	deployments []*v1.Deployment
	failed      bool
	finishTime  *time.Time
}

// GetJobs returns a list of jobs which are included in the update progress
func (up *updateProgressConfiguration) GetJobs() []*batchv1.Job {
	return up.jobs
}

// GetDeployments returns the list of deployments which needs to be updated
func (up *updateProgressConfiguration) GetDeployments() []*v1.Deployment {
	return up.deployments
}

// Failed returns whether or not the update has failed
func (up *updateProgressConfiguration) Failed() bool {
	return up.failed
}

// Successful returns true if the complete update progress has run through.
func (up *updateProgressConfiguration) Successful() bool {
	return len(up.GetJobs()) == up.FinishedJobsCount() && len(up.GetDeployments()) == up.UpdatedDeploymentsCount()
}

// Finished returns if the update progress has run through succesfully or unsuccessfully
func (up *updateProgressConfiguration) Finished() bool {
	if up.Failed() || up.Successful() {
		up.setFinishTimeIfNecessary()
		return true
	}
	return false
}

func (up *updateProgressConfiguration) setFinishTimeIfNecessary() {
	if up.finishTime == nil {
		finishTime := time.Now()
		up.finishTime = &finishTime
	}
}

// Abort cancels the run of this specific udpater.
func (up *updateProgressConfiguration) Abort() {
	up.failed = true
}

// FinishedJobsCount returns how many jobs have been finished
func (up *updateProgressConfiguration) FinishedJobsCount() int {
	count := 0
	for _, job := range up.GetJobs() {
		if isJobFinished(job) {
			count++
		}
	}
	return count
}

// UpdatedDeploymentsCount returns the amount of deployments which update has been finished
func (up *updateProgressConfiguration) UpdatedDeploymentsCount() int {
	count := 0
	for _, deployment := range up.GetDeployments() {
		if isDeploymentFinished(deployment) {
			count++
		}
	}
	return count
}

// FinishTime returns when the progress was finished. If the update hasn't finished yet, this will return nil.
func (up *updateProgressConfiguration) FinishTime() *time.Time {
	return up.finishTime
}

// Update executes the passed update plan against the given kubernetes wrapper asynchronously
func Update(updatePlan UpdatePlan, kubernetesWrapper KubernetesWrapper) UpdateProgress {
	up := &updater{
		updatePlan:        updatePlan,
		kubernetesWrapper: kubernetesWrapper,
	}
	return up.Update()
}

type updater struct {
	updatePlan        UpdatePlan
	updateProgress    *updateProgressConfiguration
	kubernetesWrapper KubernetesWrapper
}

// Update runs the update in a new go routing and returns the update progress
func (up *updater) Update() UpdateProgress {
	updatePlan := up.updatePlan
	toCreateJobs := updatePlan.GetToCreateJobs()
	jobs := make([]*batchv1.Job, len(toCreateJobs))
	for index, job := range toCreateJobs {
		jobs[index] = &job
	}
	toApplyDeployments := updatePlan.GetToApplyDeployments()
	deployments := make([]*v1.Deployment, len(toApplyDeployments))
	for index, deployment := range toApplyDeployments {
		deployments[index] = &deployment
	}

	updateProgress := &updateProgressConfiguration{
		jobs:        jobs,
		deployments: deployments,
		failed:      false,
	}
	up.updateProgress = updateProgress
	go up.runUpdate()
	return updateProgress
}

func (up *updater) runUpdate() error {
	updatePlan := up.updatePlan
	kubernetesWrapper := up.kubernetesWrapper
	updateProgressConfiguration := up.updateProgress
	jobs := updatePlan.GetToCreateJobs()
	deployments := updatePlan.GetToApplyDeployments()
	log.WithFields(log.Fields{
		"numJobs":        len(jobs),
		"numDeployments": len(deployments),
	}).Debug("Running update")
	for index, job := range jobs {
		jobLogger := log.WithFields(log.Fields{
			"name":      job.Name,
			"namespace": job.Namespace,
			"images":    strings.Join(GetImagesOf(job.Spec.Template.Spec), ", "),
		})
		jobLogger.Debug("Creating job")
		createdJob, err := kubernetesWrapper.GetJobAPIFor(job.Namespace).Create(&job)
		updateProgressConfiguration.jobs[index] = createdJob
		if err != nil {
			updateProgressConfiguration.Abort()
			jobLogger.WithError(err).Error("Error while creating job")
			raven.CaptureError(err, nil)
			return err
		}
	}

	for index, deployment := range deployments {
		deploymentLogger := log.WithFields(log.Fields{
			"name":      deployment.Name,
			"namespace": deployment.Namespace,
			"images":    strings.Join(GetImagesOf(deployment.Spec.Template.Spec), ", "),
		})
		deploymentLogger.Debug("Updating deployment")
		updatedDeployment, err := kubernetesWrapper.GetDeploymentAPIFor(deployment.Namespace).Update(&deployment)
		if err != nil {
			up.rollback()
			deploymentLogger.WithError(err).Error("Error while updating a deployment")
			raven.CaptureError(err, nil)
			return err
		}
		updateProgressConfiguration.deployments[index] = updatedDeployment
	}

	return up.monitorChangesLoop()
}

func (up *updater) monitorChangesLoop() error {
	var err error
	for ; ; err = up.monitorChanges() {
		if err != nil {
			return err
		}
		if up.updateProgress.Finished() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (up *updater) monitorChanges() error {
	err := up.monitorJobs()
	if err != nil {
		return err
	}
	err = up.monitorDeployments()
	if err != nil {
		return err
	}
	return nil
}

func (up *updater) monitorDeployments() error {
	kubernetesAPI := up.kubernetesWrapper
	for _, deployment := range up.updateProgress.GetDeployments() {
		currentDeployment, err := kubernetesAPI.GetDeploymentAPIFor(deployment.Namespace).Get(deployment.Name, metaV1.GetOptions{})
		if err != nil {
			continue
		}
		currentDeployment.DeepCopyInto(deployment)
	}
	return nil
}

func (up *updater) monitorJobs() error {
	kubernetesAPI := up.kubernetesWrapper
	status := up.updateProgress
	for _, job := range up.updateProgress.GetJobs() {
		currentJob, err := kubernetesAPI.GetJobAPIFor(job.Namespace).Get(job.Name, metaV1.GetOptions{})
		if err != nil {
			continue
		}
		if currentJob.Status.Failed > 0 {
			status.Abort()
			return up.rollback()
		}
		currentJob.DeepCopyInto(job)
	}
	return nil
}

func (up *updater) rollback() error {
	for _, deployment := range up.updateProgress.GetDeployments() {
		err := up.rollbackDeployment(deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (up *updater) rollbackDeployment(deployment *v1.Deployment) error {
	log.WithFields(log.Fields{
		"namespace": deployment.Namespace,
		"type":      "deployment",
		"name":      deployment.Name,
	}).Debug("Rolling back deployment")
	kubernetes := up.kubernetesWrapper
	replicaSetFinder := NewReplicaSetFinder(kubernetes)
	replicaSets, err := replicaSetFinder.GetSetsFor(deployment)
	if err != nil {
		return err
	}
	if len(replicaSets) == 0 {
		return ErrNoReplicaSet
	}

	replicaSetByRevision := map[string]v1.ReplicaSet{}
	for _, replicaSet := range replicaSets {
		replicaSetByRevision[replicaSet.Annotations[ReplicaSetRevisionAnnotation]] = replicaSet
	}

	generation := deployment.GetObjectMeta().GetGeneration()
	targetGeneration := int(generation) - 1
	toRollbackReplicaSet, ok := replicaSetByRevision[strconv.Itoa(targetGeneration)]

	if !ok {
		return ErrPreviousReplicaSetNotFound
	}

	deployment.Spec.Template = toRollbackReplicaSet.Spec.Template

	deploymentAPI := up.kubernetesWrapper.GetDeploymentAPIFor(deployment.Namespace)
	_, err = deploymentAPI.Update(deployment)
	return err
}

func isJobFinished(job *batchv1.Job) bool {
	return job.Status.Succeeded > 0
}

func isDeploymentFinished(deployment *v1.Deployment) bool {
	return deployment.Generation == deployment.Status.ObservedGeneration && isDeploymentStatusFinished(deployment.Status)
}

func isDeploymentStatusFinished(deploymentStatus v1.DeploymentStatus) bool {
	return deploymentStatus.Replicas == deploymentStatus.ReadyReplicas
}
