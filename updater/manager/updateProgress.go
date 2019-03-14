package manager

import (
	"kubernetes-update-manager/updater"
	"time"

	uuidGenerator "github.com/google/uuid"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

// WrapUpdateProgress returns a newly generated update progress configuration with a newly created unique uuid.
func WrapUpdateProgress(toWrapProgress updater.UpdateProgress) *UpdateProgressImpl {
	return &UpdateProgressImpl{
		uuid:     uuidGenerator.New(),
		progress: toWrapProgress,
	}
}

// UpdateProgressImpl is the implementation of the UpdateProgress interface
type UpdateProgressImpl struct {
	uuid     uuidGenerator.UUID
	progress updater.UpdateProgress
}

// UUID returns the unique identifier for the specified update progress.
func (updaterProgress *UpdateProgressImpl) UUID() uuidGenerator.UUID {
	return updaterProgress.uuid
}

// GetJobs returns a list of jobs which are included in the update progress.
func (updaterProgress *UpdateProgressImpl) GetJobs() []*batchv1.Job {
	return updaterProgress.progress.GetJobs()
}

// GetDeployments returns the list of deployments which needs to be updated.
func (updaterProgress *UpdateProgressImpl) GetDeployments() []*v1.Deployment {
	return updaterProgress.progress.GetDeployments()
}

// FinishedJobsCount returns how many jobs have been finished.
func (updaterProgress *UpdateProgressImpl) FinishedJobsCount() int {
	return updaterProgress.progress.FinishedJobsCount()
}

// UpdatedDeploymentsCount returns the amount of deployments which update has been finished.
func (updaterProgress *UpdateProgressImpl) UpdatedDeploymentsCount() int {
	return updaterProgress.progress.UpdatedDeploymentsCount()
}

// FinishTime returns when the progress was finished. If the update hasn't finished yet, this will return nil.
func (updaterProgress *UpdateProgressImpl) FinishTime() *time.Time {
	return updaterProgress.progress.FinishTime()
}

// Finished returns if the update progress has run through succesfully or unsuccessfully
func (updaterProgress *UpdateProgressImpl) Finished() bool {
	return updaterProgress.progress.Finished()
}

// Failed returns if the update is marked as failed
func (updaterProgress *UpdateProgressImpl) Failed() bool {
	return updaterProgress.progress.Failed()
}

// Successful returns true if the complete update progress has run through
func (updaterProgress *UpdateProgressImpl) Successful() bool {
	return updaterProgress.progress.Successful()
}

// Abort cancels the run of this specific udpater.
func (updaterProgress *UpdateProgressImpl) Abort() {
	updaterProgress.progress.Abort()
}
