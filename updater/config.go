package updater

import (
	"k8s.io/client-go/kubernetes"
	appsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	batchV1Interface "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewClientsetWrapper returns a wrapper for the clientset
func NewClientsetWrapper(clientset kubernetes.Interface) *ClientsetWrapper {
	return &ClientsetWrapper{
		clientset: clientset,
	}
}

// NewConfig returns a configuration object which can be used to pass state between the different
// configuration structs.
func NewConfig(clientset kubernetes.Interface, image *Image, updateClassifier string) *Config {
	return &Config{
		ClientsetWrapper: *NewClientsetWrapper(clientset),
		image:            image,
		updateClassifier: updateClassifier,
	}
}

// ClientsetWrapper wraps a kubernetes clientset and returns the interfaces necessary for running the update progress.
type ClientsetWrapper struct {
	clientset kubernetes.Interface
}

// GetClientset returns the specified client set configuration.
func (config *ClientsetWrapper) GetClientset() kubernetes.Interface {
	return config.clientset
}

// GetJobAPIFor returns the clientset's job interface
func (config *ClientsetWrapper) GetJobAPIFor(namespace string) batchV1Interface.JobInterface {
	return config.GetClientset().BatchV1().Jobs(namespace)
}

// GetNamespacesAPI returns the clientset's namespace interface
func (config *ClientsetWrapper) GetNamespacesAPI() corev1.NamespaceInterface {
	coreV1 := config.getCoreV1()
	return coreV1.Namespaces()
}

// GetDeploymentAPIFor returns the clientset's specified deployment api for the configured API
func (config *ClientsetWrapper) GetDeploymentAPIFor(namespace string) appsV1.DeploymentInterface {
	appsV1 := config.GetClientset().AppsV1()
	return appsV1.Deployments(namespace)
}

// GetReplicaSetAPIFor returns the API to interact with replicaset for the passed namespace
func (config *ClientsetWrapper) GetReplicaSetAPIFor(namespace string) appsV1.ReplicaSetInterface {
	appsV1 := config.GetClientset().AppsV1()
	return appsV1.ReplicaSets(namespace)
}

func (config *ClientsetWrapper) getCoreV1() corev1.CoreV1Interface {
	return config.GetClientset().CoreV1()
}

// Config includes a configuration for a planned update. It is used to pass between the different classes shared state.
type Config struct {
	ClientsetWrapper
	image            *Image
	updateClassifier string
	namespaces       []string
}

// GetNamespaces returns an array of all namespaces which should be used.
func (config *Config) GetNamespaces() []string {
	if config.namespaces == nil {
		config.SetNamespaces(make([]string, 0))
	}
	return config.namespaces
}

// SetNamespaces allows the configuration for which namespaces should be handled inside of the config
func (config *Config) SetNamespaces(namespaces []string) {
	config.namespaces = namespaces
}

// GetImage returns the image which should be updated.
func (config *Config) GetImage() *Image {
	return config.image
}

// GetUpdateClassifier returns the update classifier passed to this update configuration.
func (config *Config) GetUpdateClassifier() string {
	return config.updateClassifier
}
