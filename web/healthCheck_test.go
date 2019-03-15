package web

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type HealthCheckSuite struct {
	GenericWebTestSuite
}

var _ = Suite(&HealthCheckSuite{})

func (suite *HealthCheckSuite) TestHealthCall(c *C) {
	w := suite.recorder
	router := suite.router
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	c.Assert(w.Code, Equals, http.StatusNoContent)
}
