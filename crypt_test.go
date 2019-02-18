package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CryptTestSuite struct{ MainTestSuite }

func (s *CryptTestSuite) SetupTest() {
	s.MainTestSuite.SetupTest()

	conf.Keys = []securityKey{securityKey("test-key")}
	conf.Salts = []securityKey{securityKey("test-salt")}
}

func (s *CryptTestSuite) TestValidatePath() {
	err := validatePath("dtLwhdnPPiu_epMl1LrzheLpvHas-4mwvY6L3Z8WwlY", "asd")
	assert.Nil(s.T(), err)
}

func (s *CryptTestSuite) TestValidatePathTruncated() {
	conf.SignatureSize = 8

	err := validatePath("dtLwhdnPPis", "asd")
	assert.Nil(s.T(), err)
}

func (s *CryptTestSuite) TestValidatePathInvalid() {
	err := validatePath("dtLwhdnPPis", "asd")
	assert.Error(s.T(), err)
}

func (s *CryptTestSuite) TestValidatePathMultiplePairs() {
	conf.Keys = append(conf.Keys, securityKey("test-key2"))
	conf.Salts = append(conf.Salts, securityKey("test-salt2"))

	err := validatePath("dtLwhdnPPiu_epMl1LrzheLpvHas-4mwvY6L3Z8WwlY", "asd")
	assert.Nil(s.T(), err)

	err = validatePath("jbDffNPt1-XBgDccsaE-XJB9lx8JIJqdeYIZKgOqZpg", "asd")
	assert.Nil(s.T(), err)

	err = validatePath("dtLwhdnPPis", "asd")
	assert.Error(s.T(), err)
}

func TestCrypt(t *testing.T) {
	suite.Run(t, new(CryptTestSuite))
}
