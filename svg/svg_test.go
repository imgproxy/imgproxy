package svg

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type SvgTestSuite struct {
	suite.Suite
}

func (s *SvgTestSuite) SetupSuite() {
	config.Reset()

	err := imagedata.Init()
	s.Require().NoError(err)
}

func (s *SvgTestSuite) readTestFile(name string) *imagedata.ImageData {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(wd, "..", "testdata", name))
	s.Require().NoError(err)

	return &imagedata.ImageData{
		Type: imagetype.SVG,
		Data: data,
		Headers: map[string]string{
			"Content-Type":  "image/svg+xml",
			"Cache-Control": "public, max-age=12345",
		},
	}
}

func (s *SvgTestSuite) TestSanitize() {
	origin := s.readTestFile("test1.svg")
	expected := s.readTestFile("test1.sanitized.svg")

	actual, err := Sanitize(origin)

	s.Require().NoError(err)
	s.Require().Equal(string(expected.Data), string(actual.Data))
	s.Require().Equal(origin.Headers, actual.Headers)
}

func (s *SvgTestSuite) TestFixUnsupportedDropShadow() {
	origin := s.readTestFile("test1.drop-shadow.svg")
	expected := s.readTestFile("test1.drop-shadow.fixed.svg")

	actual, changed, err := FixUnsupported(origin)

	// `FixUnsupported` generates random IDs, we need to replace them for the test
	re := regexp.MustCompile(`"ds(in|of)-.+?"`)
	actualData := re.ReplaceAllString(string(actual.Data), `"ds$1-test"`)

	s.Require().NoError(err)
	s.Require().True(changed)
	s.Require().Equal(string(expected.Data), actualData)
	s.Require().Equal(origin.Headers, actual.Headers)
}

func (s *SvgTestSuite) TestFixUnsupportedNothingChanged() {
	origin := s.readTestFile("test1.svg")

	actual, changed, err := FixUnsupported(origin)

	s.Require().NoError(err)
	s.Require().False(changed)
	s.Require().Equal(origin, actual)
}

func TestSvg(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}
