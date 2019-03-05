package updater

import (
	. "gopkg.in/check.v1"
)

type NamespacesSuite struct {
	kubernetesAPI KubernetesAPI
	config        *Config
}

var _ = Suite(&NamespacesSuite{})

func (suite *NamespacesSuite) SetUpSuite(c *C) {
	suite.kubernetesAPI = NewFakeKubernetesAPI()
	suite.config = NewConfig(suite.kubernetesAPI.Client, NewImage(""), "")
}

func (suite *NamespacesSuite) TestList(c *C) {
	suite.kubernetesAPI.NewNamespace("default")
	suite.kubernetesAPI.NewNamespace("other")
	ns, err := ListNamespaces(suite.config)
	c.Assert(err, IsNil)
	c.Assert(len(ns), Equals, 2)
	c.Assert(ns[0], Equals, "default")
	c.Assert(ns[1], Equals, "other")
}
