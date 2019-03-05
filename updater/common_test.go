package updater

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	. "gopkg.in/check.v1"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var lastRevision int

func NextRevision() int {
	lastRevision++
	return lastRevision
}

func NewName() string {
	return fmt.Sprintf("name-%d", rand.Int())
}

func GetJobDefaultAnnotation(imageNames ...string) batchv1.Job {
	return GetJobWith(map[string]string{UpdateClassifier: "stable"}, imageNames...)
}

func GetJobWith(annotations map[string]string, imageNames ...string) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        NewName(),
			Namespace:   "default",
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: GetPodSpecWith(imageNames...),
			},
		},
	}
}

func GetDeploymentDefaultAnnotation(imageNames ...string) v1.Deployment {
	return GetDeploymentWith(map[string]string{UpdateClassifier: "stable"}, imageNames...)
}

func GetDeploymentWith(annotations map[string]string, imageNames ...string) v1.Deployment {
	replicas := int32(1)
	return v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        NewName(),
			Namespace:   "default",
			Annotations: annotations,
		},
		Status: v1.DeploymentStatus{
			Replicas:      replicas,
			ReadyReplicas: 0,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				Spec: GetPodSpecWith(imageNames...),
			},
		},
	}
}

func GetPodSpecWith(imageNames ...string) apiv1.PodSpec {
	containers := make([]apiv1.Container, 0)
	for _, imageName := range imageNames {
		containers = append(containers, GetContainerWith(imageName))
	}

	return apiv1.PodSpec{
		Containers: containers,
	}
}

func GetContainerWith(imageName string) apiv1.Container {
	return apiv1.Container{
		Name:  NewName(),
		Image: imageName,
	}
}

func GetReplicaSetFor(deployment *v1.Deployment) v1.ReplicaSet {
	imageNames := make([]string, 0)
	for _, container := range deployment.Spec.Template.Spec.Containers {
		imageNames = append(imageNames, container.Image)
	}
	replicaSet := GetReplicaSetWith(imageNames...)
	handle := true
	replicaSet.OwnerReferences = []metav1.OwnerReference{
		metav1.OwnerReference{
			Kind:       "Deployment",
			Controller: &handle,
			Name:       deployment.Name,
		},
	}
	replicaSet.Annotations = map[string]string{
		"deployment.kubernetes.io/revision": strconv.Itoa(NextRevision()),
	}
	return replicaSet
}

func GetReplicaSetWith(imageNames ...string) v1.ReplicaSet {
	return v1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NewName(),
			Namespace: "default",
		},
		Spec: v1.ReplicaSetSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: GetPodSpecWith(imageNames...),
			},
		},
	}
}
