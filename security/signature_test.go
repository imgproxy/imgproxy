package security

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type SignatureTestSuite struct {
	suite.Suite
}

func (s *SignatureTestSuite) SetupTest() {
	config.Reset()

	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}
}

func (s *SignatureTestSuite) TestVerifySignature() {
	err := VerifySignature("oWaL7QoW5TsgbuiS9-5-DI8S3Ibbo1gdB2SteJh3a20", "asd")
	s.Require().NoError(err)
}

func (s *SignatureTestSuite) TestVerifySignatureTruncated() {
	config.SignatureSize = 8

	err := VerifySignature("oWaL7QoW5Ts", "asd")
	s.Require().NoError(err)
}

func (s *SignatureTestSuite) TestVerifySignatureInvalid() {
	err := VerifySignature("oWaL7QoW5Ts", "asd")
	s.Require().Error(err)
}

func (s *SignatureTestSuite) TestVerifySignatureMultiplePairs() {
	config.Keys = append(config.Keys, []byte("test-key2"))
	config.Salts = append(config.Salts, []byte("test-salt2"))

	err := VerifySignature("jYz1UZ7j1BCdSzH3pZhaYf0iuz0vusoOTdqJsUT6WXI", "asd")
	s.Require().NoError(err)

	err = VerifySignature("oWaL7QoW5TsgbuiS9-5-DI8S3Ibbo1gdB2SteJh3a20", "asd")
	s.Require().NoError(err)

	err = VerifySignature("dtLwhdnPPis", "asd")
	s.Require().Error(err)
}

func (s *SignatureTestSuite) TestVerifySignatureTrusted() {
	config.TrustedSignatures = []string{"truested"}
	defer func() {
		config.TrustedSignatures = []string{}
	}()

	err := VerifySignature("truested", "asd")
	s.Require().NoError(err)

	err = VerifySignature("untrusted", "asd")
	s.Require().Error(err)
}

func TestSignature(t *testing.T) {
	suite.Run(t, new(SignatureTestSuite))
}
