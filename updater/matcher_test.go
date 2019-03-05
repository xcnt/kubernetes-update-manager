package updater

import (
	gomock "github.com/golang/mock/gomock"
	. "gopkg.in/check.v1"
	apiv1 "k8s.io/api/core/v1"
)

type MatcherSuite struct {
	controller    *gomock.Controller
	matcherConfig *MockMatchConfig
}

var _ = Suite(&MatcherSuite{})

func (suite *MatcherSuite) SetUpTest(c *C) {
	ctrl := gomock.NewController(c)
	suite.controller = ctrl
	suite.matcherConfig = NewMockMatchConfig(ctrl)
}

func (suite *MatcherSuite) MockImage(image string) {
	suite.matcherConfig.
		EXPECT().
		GetImage().
		Return(NewImage(image)).
		AnyTimes()
}

func (suite *MatcherSuite) MockUpdateClassifier(updateClassifier string) {
	suite.matcherConfig.
		EXPECT().
		GetUpdateClassifier().
		Return(updateClassifier).
		AnyTimes()
}

func (suite *MatcherSuite) TearDownTest(c *C) {
	suite.controller.Finish()
}

func (suite *MatcherSuite) TestMatchesAnnotation(c *C) {
	suite.MockUpdateClassifier("stable")
	annotations := map[string]string{
		UpdateClassifier: "stable",
	}
	c.Assert(MatchesAnnotation(suite.matcherConfig, annotations), Equals, true)
}

func (suite *MatcherSuite) TestMatchesAnnotationNotMatched(c *C) {
	suite.MockUpdateClassifier("stäble")
	annotations := map[string]string{
		UpdateClassifier: "stable",
	}
	c.Assert(MatchesAnnotation(suite.matcherConfig, annotations), Equals, false)
}

func (suite *MatcherSuite) TestMachesAnnotationNotSet(c *C) {
	suite.MockUpdateClassifier("stable")
	annotations := map[string]string{
		"other": "stable",
	}
	c.Assert(MatchesAnnotation(suite.matcherConfig, annotations), Equals, false)
}

func (suite *MatcherSuite) TestMatchesPodSpec(c *C) {
	suite.MockImage("xcnt/test:1.0.0")
	check := MatchesPodSpec(suite.matcherConfig, GetPodSpecWith("xcnt/test:0.9.9"))
	c.Assert(check, Equals, true)
}

func (suite *MatcherSuite) TestMatchesWithNonSpecifiedPodSpec(c *C) {
	suite.MockImage("xcnt/test:1.0.0")
	check := MatchesPodSpec(suite.matcherConfig, GetPodSpecWith("xcnt/test2:1.0.0"))
	c.Assert(check, Equals, false)
}

func (suite *MatcherSuite) TestMatchesWithMultipleContainers(c *C) {
	podSpec := GetPodSpecWith("xcnt/test2:1.0.0", "xcnt/test:0.1.5", "xcnt/test3:0.1.5")
	check := MatchesPodSpec(suite.matcherConfig, podSpec)
	c.Assert(check, Equals, true)
}

func (suite *MatcherSuite) TestMatchesDeployment(c *C) {
	deployment := GetDeploymentWith(map[string]string{UpdateClassifier: "stable"}, "xcnt/test:0.1.5")
	check := MatchesDeployment(suite.matcherConfig, deployment)
	c.Assert(check, Equals, true)
}

func (suite *MatcherSuite) TestMatchesDeploymentWithWrongClassifier(c *C) {
	deployment := GetDeploymentWith(map[string]string{UpdateClassifier: "latest"}, "xcnt/test:0.1.5")
	check := MatchesDeployment(suite.matcherConfig, deployment)
	c.Assert(check, Equals, false)
}

func (suite *MatcherSuite) TestMatchesDeploymentWithWrongImage(c *C) {
	deployment := GetDeploymentWith(map[string]string{UpdateClassifier: "stable"}, "xcnt/test1:0.1.5")
	check := MatchesDeployment(suite.matcherConfig, deployment)
	c.Assert(check, Equals, false)
}

func (suite *MatcherSuite) TestMatchesDeploymentWithInitContainer(c *C) {
	deployment := GetDeploymentWith(map[string]string{UpdateClassifier: "stable"}, "xcnt/test1:0.1.5")
	deployment.Spec.Template.Spec.InitContainers = []apiv1.Container{GetContainerWith("xcnt/test:0.9.9")}
	check := MatchesDeployment(suite.matcherConfig, deployment)
	c.Assert(check, Equals, true)
}
