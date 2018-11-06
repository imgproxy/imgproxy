package main

import (
	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite

	oldConf config
}

func (s *MainTestSuite) SetupTest() {
	s.oldConf = conf
}

func (s *MainTestSuite) TearDownTest() {
	conf = s.oldConf
}
