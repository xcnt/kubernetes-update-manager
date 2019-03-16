package web

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "gopkg.in/check.v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

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
		APIKey:             RandStringRunes(25),
	}
	suite.router, _ = getWeb(suite.config, false)
}

func (suite *GenericWebTestSuite) Authenticate(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("APIKey %s", suite.config.APIKey))
	return req
}
