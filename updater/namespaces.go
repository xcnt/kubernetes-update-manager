package updater

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NamespaceAPIGetter provides an interface which is required to query for namespaces
type NamespaceAPIGetter interface {
	// GetNamespacesAPI returns the clientset's namespace interface
	GetNamespacesAPI() corev1.NamespaceInterface
}

// ListNamespaces returns a list of all namespaces which are in the cluster
func ListNamespaces(config NamespaceAPIGetter) ([]string, error) {
	namespacesAPI := config.GetNamespacesAPI()
	namespaces, err := namespacesAPI.List(metaV1.ListOptions{})
	if err != nil {
		return make([]string, 0), err
	}
	names := make([]string, len(namespaces.Items))
	for index, namespace := range namespaces.Items {
		names[index] = namespace.GetName()
	}
	return names, nil
}
