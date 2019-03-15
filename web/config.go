package web

import "k8s.io/client-go/kubernetes"

// Config holds the necessary data for running the web interface of the update precense.
type Config struct {
	// Clientset holds the kubernetes configuration which should be used to access to the cluster.
	Clientset kubernetes.Interface
	// Namespaces ist he list of namespaces which should be searched when searching for update candidates. This is ignored if autoload
	// namespaces is set to true
	Namespaces []string
	// AutoloadNamespaces is a toggle scanning the cluster for all namespaces when applying the update configuration and sets all namespaces
	// as the update candidate.
	AutoloadNamespaces bool
	// APIKey is a pre shared key which is used to authenticate requests against the update endpoints.
	APIKey string
}
