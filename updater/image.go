package updater

import (
	"strings"
)

// NewImage returns the image configuration for the specified name
func NewImage(name string) *Image {
	name = strings.TrimSpace(name)
	return &Image{name: name}
}

// Image handles teh deconstruction of a name for a docker image
type Image struct {
	name string
}

// String returns the string represengtation of the image
func (image *Image) String() string {
	return image.GetName()
}

// GetName returns the complete image and tag which has been passed to the configuration
func (image *Image) GetName() string {
	return image.name
}

// GetImage returns the name of the docker image which gets configured
func (image *Image) GetImage() string {
	splits := image.getSplitConfig()
	return splits[0]
}

// GetTag returns the tag of the specified image. If no tag has been provided, an empty string is returned
func (image *Image) GetTag() string {
	splits := image.getSplitConfig()
	if len(splits) == 1 {
		return ""
	}
	return splits[1]
}

// HasTag returns if the image string has the specified tag configured
func (image *Image) HasTag() bool {
	return len(image.GetTag()) > 0
}

// Equals returns if the passed image is the same.
func (image *Image) Equals(other *Image) bool {
	return image.GetName() == other.GetName()
}

// EqualsName returns if the passed name is the same as the image configuration is
func (image *Image) EqualsName(name string) bool {
	other := NewImage(name)
	return image.Equals(other)
}

// EqualsImage checks if the specified name has the same image as the current image.
func (image *Image) EqualsImage(name string) bool {
	other := NewImage(name)
	return image.GetImage() == other.GetImage()
}

func (image *Image) getSplitConfig() []string {
	return strings.SplitN(image.GetName(), ":", 2)
}
