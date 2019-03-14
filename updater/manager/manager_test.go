package manager

import (
	"errors"
	"kubernetes-update-manager/updater"
	"os"
	"time"

	. "github.com/cbrand/gocheck_matchers"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "gopkg.in/check.v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

type ManagerSuite struct {
	clientset        *testclient.Clientset
	controller       *gomock.Controller
	manager          *Manager
	image            *updater.Image
	config           *updater.Config
	updateClassifier string
	planCalled       bool
	updateCalled     bool
	finishTime       *time.Time
}

var _ = Suite(&ManagerSuite{})

func (managerSuite *ManagerSuite) SetUpTest(c *C) {
	managerSuite.controller = gomock.NewController(c)
	managerSuite.clientset = testclient.NewSimpleClientset()
	manager := NewManager(managerSuite.clientset)
	managerSuite.manager = manager
	managerSuite.image = updater.NewImage("xcnt/test:1.0.0")
	managerSuite.updateClassifier = "stable"
	managerSuite.planCalled = false
	managerSuite.updateCalled = false
	managerSuite.config = updater.NewConfig(managerSuite.clientset, managerSuite.image, managerSuite.updateClassifier)
	manager.Plan = func(config *updater.Config) (updater.UpdatePlan, error) {
		managerSuite.planCalled = true
		return NewMockUpdatePlan(managerSuite.controller), nil
	}
	manager.Update = func(updatePlan updater.UpdatePlan, wrapper updater.KubernetesWrapper) updater.UpdateProgress {
		managerSuite.updateCalled = true
		mockUpdateProgress := NewMockUpdateProgress(managerSuite.controller)
		mockUpdateProgress.EXPECT().Finished().DoAndReturn(func() bool { return managerSuite.finishTime != nil }).MinTimes(0)
		mockUpdateProgress.EXPECT().FinishTime().DoAndReturn(func() *time.Time { return managerSuite.finishTime }).MinTimes(0)

		return mockUpdateProgress
	}
}

func (managerSuite *ManagerSuite) TearDownTest(c *C) {
	managerSuite.controller.Finish()
}

func (managerSuite *ManagerSuite) TestManagerCreate(c *C) {
	manager := managerSuite.manager
	config := updater.NewConfig(managerSuite.clientset, managerSuite.image, managerSuite.updateClassifier)
	updateProgress, err := manager.Create(config)
	c.Assert(updateProgress, NotNil)
	c.Assert(updateProgress.UUID().String(), Not(Equals), uuid.Nil.String())
	c.Assert(err, IsNil)
	c.Assert(managerSuite.planCalled, IsTrue)
	c.Assert(managerSuite.updateCalled, IsTrue)
}

func (managerSuite *ManagerSuite) TestManagerCreateWithErrorInPlan(c *C) {
	manager := managerSuite.manager
	expectedErr := errors.New("test")
	manager.Plan = func(config *updater.Config) (updater.UpdatePlan, error) {
		managerSuite.planCalled = true
		return NewMockUpdatePlan(managerSuite.controller), expectedErr
	}
	_, err := manager.Create(managerSuite.config)
	c.Assert(err, Equals, expectedErr)
	c.Assert(managerSuite.planCalled, IsTrue)
	c.Assert(managerSuite.updateCalled, IsFalse)
}

func (managerSuite *ManagerSuite) TestCleanupUnfinished(c *C) {
	manager := managerSuite.manager
	updateProgress, err := manager.Create(managerSuite.config)
	manager.Cleanup()
	retrievedProgress, err := manager.GetByString(updateProgress.UUID().String())
	c.Assert(err, IsNil)
	c.Assert(updateProgress.UUID(), Equals, retrievedProgress.UUID())
}

func (managerSuite *ManagerSuite) TestCleanupRecentlyFinished(c *C) {
	manager := managerSuite.manager
	updateProgress, err := manager.Create(managerSuite.config)
	now := time.Now()
	managerSuite.finishTime = &now
	manager.Cleanup()
	retrievedProgress, err := manager.GetByString(updateProgress.UUID().String())
	c.Assert(err, IsNil)
	c.Assert(updateProgress.UUID(), Equals, retrievedProgress.UUID())
}

func (managerSuite *ManagerSuite) TestCleanupStillInProgress(c *C) {
	manager := managerSuite.manager
	updateProgress, err := manager.Create(managerSuite.config)
	later := time.Now().Add(-10 * time.Minute).Add(+1 * time.Second)
	managerSuite.finishTime = &later
	manager.Cleanup()
	retrievedProgress, err := manager.GetByString(updateProgress.UUID().String())
	c.Assert(err, IsNil)
	c.Assert(updateProgress.UUID(), Equals, retrievedProgress.UUID())
}

func (managerSuite *ManagerSuite) TestCleanupFinished10Minute(c *C) {
	manager := managerSuite.manager
	updateProgress, err := manager.Create(managerSuite.config)
	later := time.Now().Add(-10 * time.Minute)
	managerSuite.finishTime = &later
	manager.Cleanup()
	retrievedProgress, err := manager.GetByString(updateProgress.UUID().String())
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), IsTrue)
	c.Assert(retrievedProgress, IsNil)
}

func (managerSuite *ManagerSuite) TestGetByStringWithNonUUID(c *C) {
	manager := managerSuite.manager
	item, err := manager.GetByString("not-an-uuid")
	c.Assert(item, IsNil)
	c.Assert(os.IsNotExist(err), IsFalse)
}

func (managerSuite *ManagerSuite) TestDeleteByString(c *C) {
	manager := managerSuite.manager
	updateProgress, _ := manager.Create(managerSuite.config)
	manager.DeleteByString(updateProgress.UUID().String())
	_, err := manager.Get(updateProgress.UUID())
	c.Assert(os.IsNotExist(err), IsTrue)
}

func (managerSuite *ManagerSuite) TestDeleteByStringNonUUID(c *C) {
	manager := managerSuite.manager
	manager.DeleteByString("abc")
}

func (managerSuite *ManagerSuite) TestDeleteNotExisting(c *C) {
	manager := managerSuite.manager
	manager.Delete(uuid.New())
}
