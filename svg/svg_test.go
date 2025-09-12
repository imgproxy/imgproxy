package svg

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type SvgTestSuite struct {
	suite.Suite

	testData *testutil.TestDataProvider
}

func (s *SvgTestSuite) SetupSuite() {
	s.testData = testutil.NewTestDataProvider(s.T)
}

func (s *SvgTestSuite) readTestFile(name string) imagedata.ImageData {
	data := s.testData.Read(name)
	return imagedata.NewFromBytesWithFormat(imagetype.SVG, data)
}

func (s *SvgTestSuite) compare(expected, actual imagedata.ImageData) {
	expectedData, err := io.ReadAll(expected.Reader())
	s.Require().NoError(err)

	actualData, err := io.ReadAll(actual.Reader())
	s.Require().NoError(err)

	// Trim whitespace to not care about ending newlines in test files
	expectedData = bytes.TrimSpace(expectedData)
	actualData = bytes.TrimSpace(actualData)

	s.Require().Equal(expectedData, actualData)
}

func (s *SvgTestSuite) TestSanitize() {
	origin := s.readTestFile("test1.svg")
	expected := s.readTestFile("test1.sanitized.svg")

	actual, err := Sanitize(origin)
	s.Require().NoError(err)

	s.compare(expected, actual)
}

func TestSvg(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}
