package manager

import (
	"time"

	. "github.com/cbrand/gocheck_matchers"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "gopkg.in/check.v1"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
)

type UpdateProgressSuite struct {
	controller            *gomock.Controller
	wrappedUpdateProgress *MockUpdateProgress
	updateProgress        UpdateProgress
}

var _ = Suite(&UpdateProgressSuite{})

func (suite *UpdateProgressSuite) SetUpTest(c *C) {
	suite.controller = gomock.NewController(c)
	suite.wrappedUpdateProgress = NewMockUpdateProgress(suite.controller)
	suite.updateProgress = WrapUpdateProgress(suite.wrappedUpdateProgress)
}

func (suite *UpdateProgressSuite) TearDownTest(c *C) {
	suite.controller.Finish()
}

func (suite *UpdateProgressSuite) TestUUID(c *C) {
	c.Assert(suite.updateProgress.UUID(), Not(Equals), uuid.Nil)
}

func (suite *UpdateProgressSuite) TestGetJobs(c *C) {
	suite.wrappedUpdateProgress.EXPECT().GetJobs().Return([]*batchv1.Job{}).MinTimes(1).MaxTimes(1)
	suite.updateProgress.GetJobs()
}

func (suite *UpdateProgressSuite) TestGetDeployments(c *C) {
	suite.wrappedUpdateProgress.EXPECT().GetDeployments().Return([]*v1.Deployment{}).MinTimes(1).MaxTimes(1)
	suite.updateProgress.GetDeployments()
}

func (suite *UpdateProgressSuite) TestFinishedJobsCount(c *C) {
	suite.wrappedUpdateProgress.EXPECT().FinishedJobsCount().Return(111).MinTimes(1).MaxTimes(1)
	c.Assert(suite.updateProgress.FinishedJobsCount(), Equals, 111)
}

func (suite *UpdateProgressSuite) TestUpdatedDeploymentsCount(c *C) {
	suite.wrappedUpdateProgress.EXPECT().UpdatedDeploymentsCount().Return(884).MinTimes(1).MaxTimes(1)
	c.Assert(suite.updateProgress.UpdatedDeploymentsCount(), Equals, 884)
}

func (suite *UpdateProgressSuite) TestFinishTime(c *C) {
	t := time.Now()
	suite.wrappedUpdateProgress.EXPECT().FinishTime().Return(&t).MinTimes(1).MaxTimes(1)
	c.Assert(*suite.updateProgress.FinishTime(), Equals, t)
}

func (suite *UpdateProgressSuite) TestFinished(c *C) {
	suite.wrappedUpdateProgress.EXPECT().Finished().Return(true).MinTimes(1).MaxTimes(1)
	c.Assert(suite.updateProgress.Finished(), IsTrue)
}

func (suite *UpdateProgressSuite) TestSuccessful(c *C) {
	suite.wrappedUpdateProgress.EXPECT().Successful().Return(false).MinTimes(1).MaxTimes(1)
	c.Assert(suite.updateProgress.Successful(), IsFalse)
}

func (suite *UpdateProgressSuite) TestFailed(c *C) {
	suite.wrappedUpdateProgress.EXPECT().Failed().Return(false).MinTimes(1).MaxTimes(1)
	c.Assert(suite.updateProgress.Failed(), IsFalse)
}

func (suite *UpdateProgressSuite) TestAbort(c *C) {
	suite.wrappedUpdateProgress.EXPECT().Abort().MinTimes(1).MaxTimes(1)
	suite.updateProgress.Abort()
}
