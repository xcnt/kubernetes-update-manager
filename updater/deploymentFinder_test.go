package updater

import (
	. "gopkg.in/check.v1"
)

type DeploymentFinderSuite struct {
	deploymentFinder *DeploymentFinder
	config           *Config
	kubernetesAPI    KubernetesAPI
	imageName        string
	updateClassifier string
}

var _ = Suite(&DeploymentFinderSuite{})

func (suite *DeploymentFinderSuite) SetUpTest(c *C) {
	suite.imageName = "xcnt/test:1.0.0"
	suite.updateClassifier = "stable"
	suite.kubernetesAPI = NewFakeKubernetesAPI()
	suite.kubernetesAPI.NewNamespace("default")
	suite.config = NewConfig(suite.kubernetesAPI.Client, NewImage(suite.imageName), suite.updateClassifier)
	suite.config.SetNamespaces([]string{"default"})
	suite.deploymentFinder = NewDeploymentFinder(suite.config)
}

func (suite *DeploymentFinderSuite) TestDeploymentList(c *C) {
	deployment := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: suite.updateClassifier,
		},
		suite.imageName,
	)
	deployment2 := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: "a" + suite.updateClassifier,
		},
		suite.imageName,
	)
	deployment3 := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: "b" + suite.updateClassifier,
		},
		suite.imageName,
	)
	deployment4 := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: suite.updateClassifier,
		},
		"a"+suite.imageName,
	)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment2)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment4)
	suite.kubernetesAPI.NewNamespace("other")
	suite.kubernetesAPI.NewDeploymentIn("other", deployment3)
	deployments, err := suite.deploymentFinder.List()
	c.Assert(err, IsNil)
	c.Assert(len(deployments), Equals, 1)
	c.Assert(deployments[0].Spec.Template.Spec.Containers[0].Image, Equals, suite.imageName)
}

func (suite *DeploymentFinderSuite) TestDeploymentListFor(c *C) {
	deployment := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: suite.updateClassifier,
		},
		suite.imageName,
	)
	deployment2 := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: "a" + suite.updateClassifier,
		},
		suite.imageName,
	)
	deployment3 := GetDeploymentWith(
		map[string]string{
			UpdateClassifier: "b" + suite.updateClassifier,
		},
		suite.imageName,
	)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment2)
	suite.kubernetesAPI.NewNamespace("other")
	suite.kubernetesAPI.NewDeploymentIn("other", deployment3)
	deployments, err := suite.deploymentFinder.ListFor("other")
	c.Assert(err, IsNil)
	c.Assert(len(deployments), Equals, 0)
}
