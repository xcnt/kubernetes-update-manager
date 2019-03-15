package client

import (
	"github.com/google/uuid"
	. "gopkg.in/check.v1"
)

type UpdateExecutionSuite struct{}

var _ = Suite(&UpdateExecutionSuite{})

func (updateSuite *UpdateExecutionSuite) TestUUID(c *C) {
	uuidObject := uuid.New()
	resp := UpdateExecution{
		updateProgressUUID: uuidObject.String(),
	}
	c.Assert(resp.UUID(), Equals, uuidObject)
}

func (updateSuite *UpdateExecutionSuite) TestEmptyUUID(c *C) {
	resp := UpdateExecution{}
	c.Assert(resp.UUID(), Equals, uuid.Nil)
}
