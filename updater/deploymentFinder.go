package updater

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeploymentFinder returns an interface enabling the search for deployment candidates which need to be updated
func NewDeploymentFinder(config *Config) *DeploymentFinder {
	return &DeploymentFinder{
		config: config,
	}
}

// DeploymentFinder holds the configuration for searching for deployments which should be updated
type DeploymentFinder struct {
	config *Config
}

// List returns all deployments for all configured namespaces fitting the specified update configuration
func (deploymentFinder *DeploymentFinder) List() ([]v1.Deployment, error) {
	namespaces := deploymentFinder.config.GetNamespaces()
	deployments := make([]v1.Deployment, 0)
	for _, namespace := range namespaces {
		namespaceDeployments, err := deploymentFinder.ListFor(namespace)
		if err != nil {
			return deployments, err
		}
		deployments = append(deployments, namespaceDeployments...)
	}
	return deployments, nil
}

// ListFor lists all deployments for the specified namespace and returns the configurations
func (deploymentFinder *DeploymentFinder) ListFor(namespace string) ([]v1.Deployment, error) {
	deploymentAPI := deploymentFinder.config.GetDeploymentAPIFor(namespace)
	response, err := deploymentAPI.List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return make([]v1.Deployment, 0), err
	}
	deployments := make([]v1.Deployment, 0)
	for _, deployment := range response.Items {
		if deploymentFinder.matches(deployment) {
			deployments = append(deployments, deployment)
		}
	}
	return deployments, nil
}

func (deploymentFinder *DeploymentFinder) matches(deployment v1.Deployment) bool {
	return MatchesDeployment(deploymentFinder.config, deployment)
}
