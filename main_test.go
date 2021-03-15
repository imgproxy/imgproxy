package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite

	oldConf config
}

func TestMain(m *testing.M) {
	initialize()
	os.Exit(m.Run())
}

func (s *MainTestSuite) SetupTest() {
	s.oldConf = conf
	// Reset presets
	conf.Presets = make(presets)
}

func (s *MainTestSuite) TearDownTest() {
	conf = s.oldConf
}
