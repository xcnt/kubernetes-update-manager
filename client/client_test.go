package client

import (
	"errors"
	"kubernetes-update-manager/web"
	"net/http"
	"net/url"
	"os"
	"path"

	. "github.com/cbrand/gocheck_matchers"
	"github.com/google/uuid"
	. "gopkg.in/check.v1"
	"gopkg.in/jarcoal/httpmock.v1"
)

type ClientSuite struct {
	updateCommand *UpdateCommand
}

var _ = Suite(&ClientSuite{})

func (suite *ClientSuite) SetUpTest(c *C) {
	httpmock.Activate()
	suite.updateCommand = &UpdateCommand{
		TargetEndpoint:   "https://localhost/updates/",
		Image:            "xcnt/test:1.0.0",
		UpdateClassifier: "stable",
		APIKey:           "this-is-a-test-api-key",
	}
}

func (suite *ClientSuite) TearDownTest(c *C) {
	httpmock.DeactivateAndReset()
}

func (suite *ClientSuite) mockCreate() {
	httpmock.RegisterResponder("POST", "https://localhost/updates/", func(req *http.Request) (*http.Response, error) {
		webSerialized := &web.UpdateProgressSerialized{
			UUID: uuid.New().String(),
		}
		return httpmock.NewJsonResponse(http.StatusCreated, webSerialized)
	})
}

func (suite *ClientSuite) getExecutionStatus(c *C) *UpdateExecution {
	suite.mockCreate()
	result, err := suite.updateCommand.Run()
	c.Assert(err, IsNil)
	return result.(*UpdateExecution)
}

func (suite *ClientSuite) TestRun(c *C) {
	suite.mockCreate()
	result, err := suite.updateCommand.Run()
	c.Assert(result, NotNil)
	c.Assert(err, IsNil)
}

func (suite *ClientSuite) TestRunError(c *C) {
	httpmock.RegisterResponder("POST", "https://localhost/updates/", func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("test")
	})
	result, err := suite.updateCommand.Run()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
}

func (suite *ClientSuite) TestRunErrorUnauthorized(c *C) {
	httpmock.RegisterResponder("POST", "https://localhost/updates/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusUnauthorized, ""), nil
	})
	result, err := suite.updateCommand.Run()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrUnauthorized)
}

func (suite *ClientSuite) TestRunErrorInternalServerError(c *C) {
	httpmock.RegisterResponder("POST", "https://localhost/updates/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
	})
	result, err := suite.updateCommand.Run()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
}

func (suite *ClientSuite) TestRunErrorNotDeserializable(c *C) {
	httpmock.RegisterResponder("POST", "https://localhost/updates/", func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusCreated, ""), nil
	})
	result, err := suite.updateCommand.Run()
	c.Assert(result, IsNil)
	c.Assert(err, NotNil)
}

func (suite *ClientSuite) mockGet(c *C) *UpdateExecution {
	status := suite.getExecutionStatus(c)
	parsedURL, _ := url.Parse("https://localhost/updates")
	parsedURL.Path = path.Join(parsedURL.Path, status.UUID().String())
	httpmock.RegisterResponder("GET", parsedURL.String(), func(req *http.Request) (*http.Response, error) {
		webSerialized := &web.UpdateProgressSerialized{
			UUID: status.UUID().String(),
			Status: web.StatusSerialized{
				Finished: true,
			},
		}
		return httpmock.NewJsonResponse(http.StatusOK, webSerialized)
	})
	return status
}

func (suite *ClientSuite) TestGet(c *C) {
	status := suite.mockGet(c)

	response, err := status.Get()
	c.Assert(err, IsNil)
	c.Assert(response.UUID, Equals, status.UUID().String())
	c.Assert(response.Status.Finished, IsTrue)
}

func (suite *ClientSuite) TestGetWithWrongBaseURL(c *C) {
	status := suite.getExecutionStatus(c)
	suite.updateCommand.TargetEndpoint = "sasasa://invalid$URI§Here"
	response, err := status.Get()
	c.Assert(err, NotNil)
	c.Assert(response, IsNil)
}

func (suite *ClientSuite) TestGetUnauthorized(c *C) {
	status := suite.getExecutionStatus(c)
	httpmock.RegisterResponder("GET", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusUnauthorized, ""), nil
	})
	response, err := status.Get()
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrUnauthorized)
	c.Assert(response, IsNil)
}

func (suite *ClientSuite) TestGetNotFound(c *C) {
	status := suite.getExecutionStatus(c)
	httpmock.RegisterResponder("GET", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusNotFound, ""), nil
	})
	response, err := status.Get()
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), IsTrue)
	c.Assert(response, IsNil)
}

func (suite *ClientSuite) TestGetNotDeserializable(c *C) {
	status := suite.getExecutionStatus(c)
	httpmock.RegisterResponder("GET", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, ""), nil
	})
	response, err := status.Get()
	c.Assert(err, NotNil)
	c.Assert(response, IsNil)
}

func (suite *ClientSuite) TestFinish(c *C) {
	status := suite.mockGet(c)
	httpmock.RegisterResponder("DELETE", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusNoContent, ""), nil
	})
	err := status.Finish()
	c.Assert(err, IsNil)
}

func (suite *ClientSuite) TestFinishError(c *C) {
	status := suite.mockGet(c)
	httpmock.RegisterResponder("DELETE", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusNoContent, ""), errors.New("something")
	})
	err := status.Finish()
	c.Assert(err, NotNil)
}

func (suite *ClientSuite) TestFinishUnauthorized(c *C) {
	status := suite.mockGet(c)
	httpmock.RegisterResponder("DELETE", status.objectURL().String(), func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusUnauthorized, ""), nil
	})
	err := status.Finish()
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrUnauthorized)
}
