package svg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type SvgTestSuite struct {
	suite.Suite
}

func (s *SvgTestSuite) SetupSuite() {
	config.Reset()

	err := imagedata.Init()
	s.Require().NoError(err)
}

func (s *SvgTestSuite) readTestFile(name string) imagedata.ImageData {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(wd, "..", "testdata", name))
	s.Require().NoError(err)

	d, err := imagedata.NewFromBytes(data)
	s.Require().NoError(err)

	return d
}

func (s *SvgTestSuite) TestSanitize() {
	origin := s.readTestFile("test1.svg")
	expected := s.readTestFile("test1.sanitized.svg")
	actual, err := Sanitize(origin)

	s.Require().NoError(err)
	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), actual.Reader()))
}

func TestSvg(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}
