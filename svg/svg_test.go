package svg

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SvgTestSuite struct {
	suite.Suite
}

func (s *SvgTestSuite) SetupSuite() {
	config.Reset()

	err := imagedata.Init()
	require.Nil(s.T(), err)
}

func (s *SvgTestSuite) readTestFile(name string) *imagedata.ImageData {
	wd, err := os.Getwd()
	require.Nil(s.T(), err)

	data, err := os.ReadFile(filepath.Join(wd, "..", "testdata", name))
	require.Nil(s.T(), err)

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

	require.Nil(s.T(), err)
	require.Equal(s.T(), string(expected.Data), string(actual.Data))
	require.Equal(s.T(), origin.Headers, actual.Headers)
}

func (s *SvgTestSuite) TestFixUnsupportedDropShadow() {
	origin := s.readTestFile("test1.drop-shadow.svg")
	expected := s.readTestFile("test1.drop-shadow.fixed.svg")

	actual, changed, err := FixUnsupported(origin)

	// `FixUnsupported` generates random IDs, we need to replace them for the test
	re := regexp.MustCompile(`"ds(in|of)-.+?"`)
	actualData := re.ReplaceAllString(string(actual.Data), `"ds$1-test"`)

	require.Nil(s.T(), err)
	require.True(s.T(), changed)
	require.Equal(s.T(), string(expected.Data), actualData)
	require.Equal(s.T(), origin.Headers, actual.Headers)
}

func (s *SvgTestSuite) TestFixUnsupportedNothingChanged() {
	origin := s.readTestFile("test1.svg")

	actual, changed, err := FixUnsupported(origin)

	require.Nil(s.T(), err)
	require.False(s.T(), changed)
	require.Equal(s.T(), origin, actual)
}

func TestSvg(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}
