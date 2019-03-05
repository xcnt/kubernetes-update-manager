package updater

import (
	. "gopkg.in/check.v1"
)

type JobFinderSuite struct {
	jobFinder        *JobFinder
	config           *Config
	kubernetesAPI    KubernetesAPI
	imageName        string
	updateClassifier string
}

var _ = Suite(&JobFinderSuite{})

func (suite *JobFinderSuite) SetUpTest(c *C) {
	suite.imageName = "xcnt/test:1.0.0"
	suite.updateClassifier = "stable"
	suite.kubernetesAPI = NewFakeKubernetesAPI()
	suite.kubernetesAPI.NewNamespace("default")
	suite.config = NewConfig(suite.kubernetesAPI.Client, NewImage(suite.imageName), suite.updateClassifier)
	suite.config.SetNamespaces([]string{"default"})
	suite.jobFinder = NewJobFinder(suite.config)
}

func (suite *JobFinderSuite) TestJobConfiguration(c *C) {
	job1 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, suite.imageName)
	job2 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier + "_"}, suite.imageName)
	job3 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, "1"+suite.imageName)
	job4 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, suite.imageName)

	suite.kubernetesAPI.NewNamespace("default")
	suite.kubernetesAPI.NewJobIn("default", job1)
	suite.kubernetesAPI.NewJobIn("default", job2)
	suite.kubernetesAPI.NewNamespace("other")
	suite.kubernetesAPI.NewJobIn("other", job3)
	suite.kubernetesAPI.NewJobIn("other", job4)

	jobList, err := suite.jobFinder.List()
	c.Assert(err, IsNil)
	c.Assert(len(jobList), Equals, 1)
}

func (suite *JobFinderSuite) TestJobConfigurationWithMultipleNamespaces(c *C) {
	job1 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, suite.imageName)
	job2 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier + "_"}, suite.imageName)
	job3 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, "1"+suite.imageName)
	job4 := GetJobWith(map[string]string{UpdateClassifier: suite.updateClassifier}, suite.imageName)
	suite.config.SetNamespaces([]string{"default", "other"})

	suite.kubernetesAPI.NewNamespace("default")
	suite.kubernetesAPI.NewJobIn("default", job1)
	suite.kubernetesAPI.NewJobIn("default", job2)
	suite.kubernetesAPI.NewNamespace("other")
	suite.kubernetesAPI.NewJobIn("other", job3)
	suite.kubernetesAPI.NewJobIn("other", job4)

	jobList, err := suite.jobFinder.List()
	c.Assert(err, IsNil)
	c.Assert(len(jobList), Equals, 1)
	c.Assert(jobList[0].Name, Equals, job1.Name)
}
