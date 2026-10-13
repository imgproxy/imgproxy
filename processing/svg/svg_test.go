package svg_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/processing/svg"
	"github.com/imgproxy/imgproxy/v4/testutil"
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

	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	actual, err := svgProc.Process(options.New(), origin)
	s.Require().NoError(err)

	s.compare(expected, actual)
}

func (s *SvgTestSuite) TestSanitizeAnimation() {
	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "animate href",
			in:       `<svg><a href="#"><animate attributeName="href" to="javascript:alert(1)"/><text>x</text></a></svg>`,
			expected: `<svg><a href="#"><text>x</text></a></svg>`,
		},
		{
			name:     "set xlink:href",
			in:       `<svg><a href="#"><set attributeName="xlink:href" to="javascript:alert(1)"/></a></svg>`,
			expected: `<svg><a href="#"/></svg>`,
		},
		{
			name:     "uppercase with spaces",
			in:       `<svg><a href="#"><animate attributeName=" HREF " to="x"/></a></svg>`,
			expected: `<svg><a href="#"/></svg>`,
		},
		{
			name:     "other attribute kept",
			in:       `<svg><rect><animate attributeName="opacity" to="0"/></rect></svg>`,
			expected: `<svg><rect><animate attributeName="opacity" to="0"/></rect></svg>`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			in := imagedata.NewFromBytesWithFormat(imagetype.SVG, []byte(tt.in))
			actual, err := svgProc.Process(options.New(), in)
			s.Require().NoError(err)

			expected := imagedata.NewFromBytesWithFormat(imagetype.SVG, []byte(tt.expected))
			s.compare(expected, actual)
		})
	}
}

func TestSvg(t *testing.T) {
	suite.Run(t, new(SvgTestSuite))
}

func BenchmarkSvgProcessing(b *testing.B) {
	testImagesPath, err := filepath.Abs("../../testdata/test-images/svg-test-suite")
	if err != nil {
		b.Fatal(err)
	}

	samples := []imagedata.ImageData{}

	err = filepath.Walk(testImagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			b.Fatal(err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-SVG files
		if filepath.Ext(path) != ".svg" {
			return nil
		}

		// Read SVG file
		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}

		samples = append(samples, imagedata.NewFromBytesWithFormat(imagetype.SVG, data))
		return nil
	})

	if err != nil {
		b.Fatal(err)
	}

	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	opts := options.New()

	b.ResetTimer()

	for b.Loop() {
		for _, sample := range samples {
			_, err := svgProc.Process(opts, sample)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
