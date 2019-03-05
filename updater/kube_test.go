package updater

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func NewFakeKubernetesAPI() KubernetesAPI {
	return KubernetesAPI{
		Client: testclient.NewSimpleClientset(),
	}
}

type KubernetesAPI struct {
	Client *testclient.Clientset
}

// NewNamespaceWithPostfix creates a new namespace with a stable postfix
func (k KubernetesAPI) NewNamespace(namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err := k.Client.CoreV1().Namespaces().Create(ns)
	return err
}

// NewDeploymentIn creates the specified deployment configuration
func (k KubernetesAPI) NewDeploymentIn(namespace string, deployment appsv1.Deployment) error {
	_, err := k.Client.AppsV1().Deployments(namespace).Create(&deployment)
	return err
}

// UpdateDeploymentIn updates the specified deployment in the provided namespace
func (k KubernetesAPI) UpdateDeploymentIn(namespace string, deployment *appsv1.Deployment) error {
	_, err := k.Client.AppsV1().Deployments(namespace).Update(deployment)
	return err
}

// NewJobIn creates the job in the provided namespace
func (k KubernetesAPI) NewJobIn(namespace string, job batchv1.Job) error {
	_, err := k.Client.BatchV1().Jobs(namespace).Create(&job)
	return err
}

// UpdateJobIn updates the job in the provided namespace
func (k KubernetesAPI) UpdateJobIn(namespace string, job *batchv1.Job) error {
	_, err := k.Client.BatchV1().Jobs(namespace).Update(job)
	return err
}

// NewReplicaSetIn creates a new replica set in the provided namespace
func (k KubernetesAPI) NewReplicaSetIn(namespace string, replicaSet appsv1.ReplicaSet) error {
	_, err := k.Client.AppsV1().ReplicaSets(namespace).Create(&replicaSet)
	return err
}
