package updater

import apiv1 "k8s.io/api/core/v1"

// GetImagesOf returns all string images of the specified podspec.
func GetImagesOf(podSpec apiv1.PodSpec) []string {
	images := map[string]bool{}
	fillStringMap(images, getContainerImagesOf(podSpec.Containers))
	fillStringMap(images, getContainerImagesOf(podSpec.InitContainers))
	return stringMapToSlice(images)
}

func fillStringMap(mapConfig map[string]bool, stringSlice []string) []string {
	for container := range mapConfig {
		stringSlice = append(stringSlice, container)
	}
	return stringSlice
}

func getContainerImagesOf(containers []apiv1.Container) []string {
	images := map[string]bool{}

	for _, container := range containers {
		images[container.Image] = true
	}

	return stringMapToSlice(images)
}

func stringMapToSlice(stringMap map[string]bool) []string {
	imageSlice := make([]string, len(stringMap))
	i := 0
	for image := range stringMap {
		imageSlice[i] = image
		i++
	}
	return imageSlice
}
