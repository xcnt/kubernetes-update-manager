package web

import (
	. "gopkg.in/check.v1"
)

type RouterTestSuite struct {
	GenericWebTestSuite
}

var _ = Suite(&RouterTestSuite{})

func (suite *RouterTestSuite) TestGetWebWithMiddlewares(c *C) {
	router := GetWeb(suite.config)
	c.Assert(router, NotNil)
}
