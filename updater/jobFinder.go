package updater

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewJobFinder returns an interface enabling the search for jobs which are used for migrations
func NewJobFinder(config *Config) *JobFinder {
	return &JobFinder{
		config,
	}
}

// JobFinder includes functionality to search for jobs which should be used when upgrading the jobs
type JobFinder struct {
	*Config
}

// List returns all jobs which are configured for the provided image.
func (jobFinder *JobFinder) List() ([]batchv1.Job, error) {
	jobs := make([]batchv1.Job, 0)
	for _, namespace := range jobFinder.GetNamespaces() {
		namespaceJobs, err := jobFinder.ListFor(namespace)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, namespaceJobs...)
	}
	return jobs, nil
}

// ListFor returns the migration jobs in the specified namespace
func (jobFinder *JobFinder) ListFor(namespace string) ([]batchv1.Job, error) {
	jobAPI := jobFinder.GetJobAPIFor(namespace)
	response, err := jobAPI.List(context.TODO(), metaV1.ListOptions{})
	jobs := make([]batchv1.Job, 0)
	if err != nil {
		return jobs, err
	}

	for _, job := range response.Items {
		if jobFinder.matches(job) {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

func (jobFinder *JobFinder) matches(job batchv1.Job) bool {
	return MatchesJob(jobFinder.Config, job)
}
