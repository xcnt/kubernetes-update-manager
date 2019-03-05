package updater

import (
	. "gopkg.in/check.v1"
)

type ReplicaSetFinderSuite struct {
	replicaSetFinder *ReplicaSetFinder
	config           *Config
	kubernetesAPI    KubernetesAPI
	imageName        string
}

var _ = Suite(&ReplicaSetFinderSuite{})

func (suite *ReplicaSetFinderSuite) SetUpTest(c *C) {
	suite.imageName = "xcnt/test:1.0.0"
	suite.kubernetesAPI = NewFakeKubernetesAPI()
	suite.kubernetesAPI.NewNamespace("default")
	suite.config = NewConfig(suite.kubernetesAPI.Client, NewImage(suite.imageName), "default")
	suite.config.SetNamespaces([]string{"default"})
	suite.replicaSetFinder = NewReplicaSetFinder(suite.config)
}

func (suite *ReplicaSetFinderSuite) TestFindReplicaForDeployment(c *C) {
	otherDeployment := GetDeploymentDefaultAnnotation(suite.imageName)
	deployment := GetDeploymentDefaultAnnotation(suite.imageName)
	suite.kubernetesAPI.NewDeploymentIn("default", deployment)
	suite.kubernetesAPI.NewDeploymentIn("default", otherDeployment)
	suite.kubernetesAPI.NewDeploymentIn("default", GetDeploymentDefaultAnnotation(suite.imageName))
	rsTarget1 := GetReplicaSetFor(&deployment)
	rsTarget2 := GetReplicaSetFor(&deployment)
	suite.kubernetesAPI.NewReplicaSetIn("default", rsTarget1)
	suite.kubernetesAPI.NewReplicaSetIn("default", rsTarget2)
	suite.kubernetesAPI.NewReplicaSetIn("default", GetReplicaSetFor(&otherDeployment))
	replicaSetsOfDeployment, err := suite.replicaSetFinder.GetSetsFor(&deployment)
	c.Assert(err, IsNil)
	c.Assert(len(replicaSetsOfDeployment), Equals, 2)
	c.Assert(replicaSetsOfDeployment[0].Name, Equals, rsTarget1.Name)
	c.Assert(replicaSetsOfDeployment[1].Name, Equals, rsTarget2.Name)
}
