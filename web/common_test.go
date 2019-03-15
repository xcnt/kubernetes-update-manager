package web

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "gopkg.in/check.v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type GenericWebTestSuite struct {
	config    *Config
	clientset *testclient.Clientset
	router    *gin.Engine
	recorder  *httptest.ResponseRecorder
}

func (suite *GenericWebTestSuite) SetUpTest(c *C) {
	gin.SetMode(gin.ReleaseMode)
	suite.recorder = httptest.NewRecorder()
	suite.clientset = testclient.NewSimpleClientset()
	suite.config = &Config{
		Clientset:          suite.clientset,
		AutoloadNamespaces: true,
	}
	suite.router = GetWeb(suite.config)
}
