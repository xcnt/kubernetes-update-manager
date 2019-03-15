package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"kubernetes-update-manager/updater"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	. "gopkg.in/check.v1"
)

type UpdaterTestSuite struct {
	GenericWebTestSuite
}

var _ = Suite(&UpdaterTestSuite{})

func (suite *UpdaterTestSuite) TestGetUnauthorized(c *C) {
	searchUUID := uuid.New().String()

	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("GET", fmt.Sprintf("/updates/%s", searchUUID), nil)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusUnauthorized)
}

func (suite *UpdaterTestSuite) TestGetNotFound(c *C) {
	searchUUID := uuid.New().String()

	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("GET", fmt.Sprintf("/updates/%s", searchUUID), nil)
	suite.Authenticate(req)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusNotFound)
}

func (suite *UpdaterTestSuite) TestGetInvalidUUID(c *C) {
	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("GET", "/updates/abc", nil)
	suite.Authenticate(req)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusBadRequest)
}

func (suite *UpdaterTestSuite) PostRequestWith(data url.Values) *http.Request {
	req, _ := http.NewRequest("POST", "/updates", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	suite.Authenticate(req)
	return req
}

func (suite *UpdaterTestSuite) PostRequestComplete() *http.Request {
	data := url.Values{}
	data.Set(ImageParam, "xcnt/test:1.0.0")
	data.Set(UpdateClassifierParam, "stable")
	return suite.PostRequestWith(data)
}

func (suite *UpdaterTestSuite) TestPostUnauthorized(c *C) {
	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("POST", "/updates", nil)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusUnauthorized)
}

func (suite *UpdaterTestSuite) TestPostNoImage(c *C) {
	w := suite.recorder
	router := suite.router
	data := url.Values{}
	data.Set(UpdateClassifierParam, "stable")
	req := suite.PostRequestWith(data)

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusBadRequest)
}

func (suite *UpdaterTestSuite) TestPostNoUpdateClassifier(c *C) {
	w := suite.recorder
	router := suite.router
	data := url.Values{}
	data.Set(ImageParam, "xcnt/test:1.0.0")
	req := suite.PostRequestWith(data)

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusBadRequest)
}

func (suite *UpdaterTestSuite) TestPost(c *C) {
	w := suite.recorder
	router := suite.router
	req := suite.PostRequestComplete()

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusOK)
	buffer := bytes.Buffer{}
	buffer.ReadFrom(w.Body)
	response := &UpdateProgressSerialized{}
	err := json.Unmarshal(buffer.Bytes(), response)
	c.Assert(err, IsNil)
	c.Assert(response.UUID, Not(Equals), uuid.Nil)
}

func (suite *UpdaterTestSuite) TestPostNoAutloadNamespaces(c *C) {
	w := suite.recorder
	router := suite.router
	suite.config.AutoloadNamespaces = false
	suite.config.Namespaces = []string{"default"}
	req := suite.PostRequestComplete()

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusOK)
	buffer := bytes.Buffer{}
	buffer.ReadFrom(w.Body)
	response := &UpdateProgressSerialized{}
	err := json.Unmarshal(buffer.Bytes(), response)
	c.Assert(err, IsNil)
	c.Assert(response.UUID, Not(Equals), uuid.Nil)
}

func (suite *UpdaterTestSuite) TestPostWithError(c *C) {
	w := suite.recorder
	req := suite.PostRequestComplete()
	router, mgr := getWeb(suite.config)
	mgr.Plan = func(_ *updater.Config) (updater.UpdatePlan, error) {
		return nil, errors.New("this is an error")
	}

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusInternalServerError)
}

func (suite *UpdaterTestSuite) TestGetWithUUID(c *C) {
	w := suite.recorder
	router := suite.router
	req := suite.PostRequestComplete()

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusOK)
	buffer := bytes.Buffer{}
	buffer.ReadFrom(w.Body)
	response := &UpdateProgressSerialized{}
	err := json.Unmarshal(buffer.Bytes(), response)
	c.Assert(err, IsNil)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/updates/%s", response.UUID), nil)
	suite.Authenticate(req)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusOK)
	getResponse := &UpdateProgressSerialized{}
	err = json.Unmarshal(buffer.Bytes(), getResponse)
	c.Assert(err, IsNil)

	c.Assert(getResponse.UUID, Equals, response.UUID)
}

func (suite *UpdaterTestSuite) TestDeleteUnauthorized(c *C) {
	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/updates/%s", uuid.New().String()), nil)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusUnauthorized)
}

func (suite *UpdaterTestSuite) TestDelete(c *C) {
	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/updates/%s", uuid.New().String()), nil)
	suite.Authenticate(req)

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusNoContent)
}

func (suite *UpdaterTestSuite) TestDeleteWithExisting(c *C) {
	w := suite.recorder
	router := suite.router
	req := suite.PostRequestComplete()

	router.ServeHTTP(w, req)
	c.Assert(w.Code, Equals, http.StatusOK)
	buffer := bytes.Buffer{}
	buffer.ReadFrom(w.Body)
	response := &UpdateProgressSerialized{}
	err := json.Unmarshal(buffer.Bytes(), response)
	c.Assert(err, IsNil)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/updates/%s", response.UUID), nil)
	suite.Authenticate(req)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusNoContent)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/updates/%s", response.UUID), nil)
	suite.Authenticate(req)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusNotFound)
}
