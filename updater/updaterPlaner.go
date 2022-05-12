package updater

import (
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	controllerUUIDLabel = "controller-uid"
	jobNameLabel        = "job-name"
	nameLabel           = "name"
)

func generateJobName(jobName string) string {
	dateFormat := "2006-01-02-15-04-05"
	maxJobNameSize := 63 - len(dateFormat)
	if len(jobName) > maxJobNameSize {
		jobName = jobName[:maxJobNameSize]
	}
	jobName = fmt.Sprintf("%s-%s", jobName, time.Now().Format("2006-01-02-15-04-05"))
	return jobName
}

// Plan returns an UpdatePlan for the specified configuration
func Plan(config *Config) (UpdatePlan, error) {
	deployments, err := NewDeploymentFinder(config).List()
	if err != nil {
		return nil, err
	}

	jobs, err := NewJobFinder(config).List()
	if err != nil {
		return nil, err
	}

	updatePlaner := &UpdatePlaner{
		JobLister:        func() []batchv1.Job { return jobs },
		DeploymentLister: func() []v1.Deployment { return deployments },
	}
	return updatePlaner.Plan(config), nil
}

type updatePlan struct {
	deployments []v1.Deployment
	jobs        []batchv1.Job
}

// GetToCreateJobs returns a slice of jobs which should be created for the deployments to run.
func (updatePlan *updatePlan) GetToCreateJobs() []batchv1.Job {
	return updatePlan.jobs
}

// GetToApplyDeployments returns a slice of deployments which are the deployment configurations needed to be applied to the cluster for the update
// to run through
func (updatePlan *updatePlan) GetToApplyDeployments() []v1.Deployment {
	return updatePlan.deployments
}

// UpdatePlaner provides a configuration struct to generate planed upgrades for specific deployments and jobs.
type UpdatePlaner struct {
	// JobLister is a function which returns all jobs which should be used for update migrations
	JobLister func() []batchv1.Job
	// DeploymentLister is a function which returns all deployments which should be adjusted for the update to run through.
	DeploymentLister func() []v1.Deployment
	config           *Config
}

// Plan returns the update plan which needs to be applied for the configuration to work
func (updatePlaner *UpdatePlaner) Plan(config *Config) UpdatePlan {
	updatePlaner.config = config
	deployments := updatePlaner.updatedDeployments()
	jobs := updatePlaner.migrationJobs()
	return &updatePlan{
		deployments: deployments,
		jobs:        jobs,
	}
}

func (updatePlaner *UpdatePlaner) updatedDeployments() []v1.Deployment {
	deployments := updatePlaner.DeploymentLister()
	updatedDeployments := make([]v1.Deployment, len(deployments))
	for index, deployment := range deployments {
		newDeployment := *deployment.DeepCopy()
		newDeployment.Spec.Template.Spec = updatePlaner.updatePodSpec(newDeployment.Spec.Template.Spec)
		updatedDeployments[index] = newDeployment
	}
	return updatedDeployments
}

func (updatePlaner *UpdatePlaner) migrationJobs() []batchv1.Job {
	jobs := updatePlaner.JobLister()
	updatedJobs := make([]batchv1.Job, len(jobs))
	for index, job := range jobs {
		updatedJobs[index] = updatePlaner.createMigrationJob(job)
	}
	return updatedJobs
}

func (updatePlaner *UpdatePlaner) createMigrationJob(job batchv1.Job) batchv1.Job {
	clonedJob := *job.DeepCopy()
	clonedJob.SetUID("")
	clonedJob.SelfLink = ""
	clonedJob.Name = generateJobName(clonedJob.Name)
	clonedJob.ResourceVersion = ""
	delete(clonedJob.Annotations, UpdateClassifier)
	labels := clonedJob.Spec.Template.ObjectMeta.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	for _, key := range []string{controllerUUIDLabel, nameLabel} {
		delete(labels, key)
	}
	labels[jobNameLabel] = clonedJob.Name
	clonedJob.Spec.Template.ObjectMeta.SetLabels(labels)
	clonedJob.Spec.Selector = nil
	clonedJob.Spec.Template.Spec = updatePlaner.updatePodSpec(clonedJob.Spec.Template.Spec)
	return clonedJob
}

func (updatePlaner *UpdatePlaner) updatePodSpec(podSpec apiv1.PodSpec) apiv1.PodSpec {
	podSpec.Containers = updatePlaner.updateContainers(podSpec.Containers)
	podSpec.InitContainers = updatePlaner.updateContainers(podSpec.InitContainers)
	return podSpec
}

func (updatePlaner *UpdatePlaner) updateContainers(toUpdateContainers []apiv1.Container) []apiv1.Container {
	image := updatePlaner.config.GetImage()
	if toUpdateContainers == nil {
		return make([]apiv1.Container, 0)
	}
	containers := make([]apiv1.Container, len(toUpdateContainers))
	for index, container := range toUpdateContainers {
		if image.EqualsImage(container.Image) {
			container.Image = image.String()
		}
		containers[index] = container
	}
	return containers
}
