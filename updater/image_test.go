package updater

import (
	"strings"

	. "gopkg.in/check.v1"
)

type ImageSuite struct {
	imageName string
	image     *Image
}

var _ = Suite(&ImageSuite{})

func (s *ImageSuite) SetUpTest(c *C) {
	s.imageName = "eu.gcr.io/xcnt-infrastructure/jenkins:2018-02-18-v1"
	s.image = NewImage(s.imageName)
}

func (s *ImageSuite) TestString(c *C) {
	c.Assert(s.image.String(), Equals, s.imageName)
}

func (s *ImageSuite) TestEquals(c *C) {
	c.Assert(s.image.EqualsName(s.imageName), Equals, true)
}

func (s *ImageSuite) TestGetImage(c *C) {
	c.Assert(s.image.GetImage(), Equals, "eu.gcr.io/xcnt-infrastructure/jenkins")
}

func (s *ImageSuite) TestGetImageWithoutTag(c *C) {
	image := NewImage(s.image.GetImage())
	c.Assert(image.GetImage(), Equals, "eu.gcr.io/xcnt-infrastructure/jenkins")
}

func (s *ImageSuite) TestGetTag(c *C) {
	c.Assert(s.image.GetTag(), Equals, "2018-02-18-v1")
}

func (s *ImageSuite) TestGetTagWithoutTag(c *C) {
	image := NewImage(s.image.GetImage())
	c.Assert(image.GetTag(), Equals, "")
}

func (s *ImageSuite) TestHasTag(c *C) {
	c.Assert(s.image.HasTag(), Equals, true)
}

func (s *ImageSuite) TestHasTagWithoutTag(c *C) {
	image := NewImage(s.image.GetImage())
	c.Assert(image.HasTag(), Equals, false)
}

func (s *ImageSuite) TestEqualsName(c *C) {
	c.Assert(s.image.EqualsName(s.imageName), Equals, true)
}

func (s *ImageSuite) TestEqualsImage(c *C) {
	c.Assert(s.image.EqualsImage(s.imageName), Equals, true)
}

func (s *ImageSuite) TestEqualsImageWithOtherTag(c *C) {
	name := strings.Join([]string{s.image.GetImage(), "some-other-tag"}, ":")
	c.Assert(s.image.EqualsImage(name), Equals, true)
}

func (s *ImageSuite) TestEqualsImageOtherImageSameTag(c *C) {
	name := strings.Join([]string{"eu.gcr.io/xcnt-infrastructure/jenkins2", s.image.GetTag()}, ":")
	c.Assert(s.image.EqualsImage(name), Equals, false)
}
