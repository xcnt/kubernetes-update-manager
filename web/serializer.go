package web

import (
	"kubernetes-update-manager/updater/manager"
	"time"
)

// ProgressCountSerialized shows the progress information of the current
// status of the update.
type ProgressCountSerialized struct {
	// Total amount of items which needs to be processed
	Total int `json:"total"`
	// Updated amount of items which have already been processed
	Updated int `json:"updated"`
}

// CountSerialized specifies the counts of the current job.
type CountSerialized struct {
	// Jobs count for this update step to process
	Jobs ProgressCountSerialized `json:"jobs"`
	// Deployments count of this update step to process
	Deployments ProgressCountSerialized `json:"deployments"`
}

// StatusSerialized returns information about the current status
// of the update job progress.
type StatusSerialized struct {
	// FinishTime is nil, when the job hasn't run through yet
	// and returns the time when the update has completed either
	// successfully or unsuccessfully.
	FinishTime *time.Time `json:"finish_time"`
	// Finished returns if the update has been completed in a
	// successful or failed way.
	Finished bool `json:"finished"`
	// Failed returns if the update has failed
	Failed bool `json:"failed"`
	// Successful returns if the update has been succesful
	Successful bool `json:"successful"`
}

// UpdateProgressSerialized represents a serialized upgrade step
// which is used in the web interface to update information about
// the current update progress.
type UpdateProgressSerialized struct {
	// UUID returns the unique identifier of this update configuration.
	UUID string `json:"uuid"`
	// Counts returns the amount of jobs and deployments when it has been progressed
	Counts CountSerialized `json:"counts"`
	// Status returns the current status of the update progress.
	Status StatusSerialized `json:"status"`
}

func serializeUpdateProgress(progress manager.UpdateProgress) *UpdateProgressSerialized {
	return &UpdateProgressSerialized{
		UUID: progress.UUID().String(),
		Counts: CountSerialized{
			Deployments: ProgressCountSerialized{
				Total:   len(progress.GetDeployments()),
				Updated: progress.UpdatedDeploymentsCount(),
			},
			Jobs: ProgressCountSerialized{
				Total:   len(progress.GetJobs()),
				Updated: progress.FinishedJobsCount(),
			},
		},
		Status: StatusSerialized{
			FinishTime: progress.FinishTime(),
			Finished:   progress.Finished(),
			Failed:     progress.Failed(),
			Successful: progress.Successful(),
		},
	}
}
