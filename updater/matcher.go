package updater

import (
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	// UpdateClassifier specifies which updater should be included
	UpdateClassifier = "xcnt.io/update-classifier"
)

// MatchesDeployment returns if the specified deployment includes the matching configuration.
func MatchesDeployment(matchConfig MatchConfig, deployment v1.Deployment) bool {
	if !matchesDeploymentSpec(matchConfig, deployment.Spec) {
		return false
	}
	meta := deployment.GetObjectMeta()
	return MatchesAnnotation(matchConfig, meta.GetAnnotations())
}

// matchesDeploymentSpec returns if the specified deployment specification includes the image
// which should be updated.
func matchesDeploymentSpec(matchConfig MatchConfig, deployment v1.DeploymentSpec) bool {
	return MatchesPodSpec(matchConfig, deployment.Template.Spec)
}

// MatchesJob checks if the specified job fits to the match configuration.
func MatchesJob(matchConfig MatchConfig, job batchv1.Job) bool {
	if !matchesJobSpec(matchConfig, job.Spec) {
		return false
	}
	meta := job.GetObjectMeta()
	return MatchesAnnotation(matchConfig, meta.GetAnnotations())
}

func matchesJobSpec(matchConfig MatchConfig, jobSpec batchv1.JobSpec) bool {
	return MatchesPodSpec(matchConfig, jobSpec.Template.Spec)
}

// MatchesPodSpec checks if the pod specification includes the specified image
// which returns data.
func MatchesPodSpec(matchConfig MatchConfig, podSpec apiv1.PodSpec) bool {
	return matchesContainers(matchConfig, podSpec.Containers) || matchesContainers(matchConfig, podSpec.InitContainers)
}

func matchesContainers(matchConfig MatchConfig, containers []apiv1.Container) bool {
	for _, container := range containers {
		if matchesContainer(matchConfig, container) {
			return true
		}
	}
	return false
}

func matchesContainer(matchConfig MatchConfig, container apiv1.Container) bool {
	return matchConfig.GetImage().EqualsImage(container.Image)
}

// MatchesAnnotation returns if the annotation includes the
// specified update classifier
func MatchesAnnotation(matchConfig MatchConfig, annotations map[string]string) bool {
	item, ok := annotations[UpdateClassifier]
	if !ok {
		return ok
	}

	return item == matchConfig.GetUpdateClassifier()
}
