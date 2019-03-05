package updater

import (
	"sort"
	"strconv"

	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSetRevisionAnnotation is the annotation used to specify the revision of a specific replicaset for a deployment.
const ReplicaSetRevisionAnnotation = "deployment.kubernetes.io/revision"

// NewReplicaSetFinder returns a struct which operates on a kubernetes cluster to find replicasets
func NewReplicaSetFinder(config KubernetesWrapper) *ReplicaSetFinder {
	return &ReplicaSetFinder{wrapper: config}
}

// ReplicaSetFinder wraps a kubernetes API and provides functionality to retrieve and filter
// data from replicasets
type ReplicaSetFinder struct {
	wrapper KubernetesWrapper
}

// GetSetsFor returns all replicaSets which belong to the specified deployment sorted by revision.
func (rsFinder *ReplicaSetFinder) GetSetsFor(deployment *v1.Deployment) ([]v1.ReplicaSet, error) {
	replicaSets, err := rsFinder.GetSetsForNamespace(deployment.Namespace)
	if err != nil {
		return nil, err
	}
	filteredSets := make([]v1.ReplicaSet, 0)
	for _, replicaSet := range replicaSets {
		if replicaSetMatchesForDeployment(replicaSet, deployment) {
			filteredSets = append(filteredSets, replicaSet)
		}
	}

	sort.Slice(
		filteredSets,
		func(left int, right int) bool {
			return rsRevisionFromAnnotation(filteredSets[left]) < rsRevisionFromAnnotation(filteredSets[right])
		},
	)
	return filteredSets, nil
}

// GetSetsForNamespace enumerates all replica sets inside of the provided namespace. The result is sorted by
// the revisions in ascending order.
func (rsFinder *ReplicaSetFinder) GetSetsForNamespace(name string) ([]v1.ReplicaSet, error) {
	replicaSetAPI := rsFinder.wrapper.GetReplicaSetAPIFor(name)
	replicaSets, err := replicaSetAPI.List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return replicaSets.Items, err
}

func rsRevisionFromAnnotation(replicaSet v1.ReplicaSet) int {
	revisionString, ok := replicaSet.Annotations[ReplicaSetRevisionAnnotation]
	if !ok {
		return -1
	}
	revision, err := strconv.Atoi(revisionString)
	if err != nil {
		return -1
	}
	return revision
}

func replicaSetMatchesForDeployment(replicaSet v1.ReplicaSet, deployment *v1.Deployment) bool {
	for _, ownerReference := range replicaSet.ObjectMeta.OwnerReferences {
		if ownerReference.Kind == "Deployment" && ownerReference.Name == deployment.Name {
			return true
		}
	}
	return false
}
