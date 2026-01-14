package security

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/testutil"
)

type SignatureTestSuite struct {
	testutil.LazySuite

	config   testutil.LazyObj[*Config]
	security testutil.LazyObj[*Checker]
}

func (s *SignatureTestSuite) SetupSuite() {
	s.config, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Config, error) {
			c := NewDefaultConfig()
			return &c, nil
		},
	)

	s.security, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Checker, error) {
			return New(s.config())
		},
	)
}

func (s *SignatureTestSuite) SetupTest() {
	s.config().Keys = [][]byte{[]byte("test-key")}
	s.config().Salts = [][]byte{[]byte("test-salt")}
}

func (s *SignatureTestSuite) TestVerifySignature() {
	err := s.security().VerifySignature(s.T().Context(), "oWaL7QoW5TsgbuiS9-5-DI8S3Ibbo1gdB2SteJh3a20", "asd")
	s.Require().NoError(err)
}

func (s *SignatureTestSuite) TestVerifySignatureTruncated() {
	s.config().SignatureSize = 8

	err := s.security().VerifySignature(s.T().Context(), "oWaL7QoW5Ts", "asd")
	s.Require().NoError(err)
}

func (s *SignatureTestSuite) TestVerifySignatureInvalid() {
	err := s.security().VerifySignature(s.T().Context(), "oWaL7QoW5Ts", "asd")
	s.Require().Error(err)
}

func (s *SignatureTestSuite) TestVerifySignatureMultiplePairs() {
	s.config().Keys = append(s.config().Keys, []byte("test-key2"))
	s.config().Salts = append(s.config().Salts, []byte("test-salt2"))

	err := s.security().VerifySignature(s.T().Context(), "jYz1UZ7j1BCdSzH3pZhaYf0iuz0vusoOTdqJsUT6WXI", "asd")
	s.Require().NoError(err)

	err = s.security().VerifySignature(s.T().Context(), "oWaL7QoW5TsgbuiS9-5-DI8S3Ibbo1gdB2SteJh3a20", "asd")
	s.Require().NoError(err)

	err = s.security().VerifySignature(s.T().Context(), "dtLwhdnPPis", "asd")
	s.Require().Error(err)
}

func (s *SignatureTestSuite) TestVerifySignatureTrusted() {
	s.config().TrustedSignatures = []string{"truested"}

	err := s.security().VerifySignature(s.T().Context(), "truested", "asd")
	s.Require().NoError(err)

	err = s.security().VerifySignature(s.T().Context(), "untrusted", "asd")
	s.Require().Error(err)
}

func TestSignature(t *testing.T) {
	suite.Run(t, new(SignatureTestSuite))
}
