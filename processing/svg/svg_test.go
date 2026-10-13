package svg_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func (s *SvgTestSuite) TestSanitizeHref() {
	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "tab entity in scheme",
			in:       `<svg><a href="java&#9;script:alert(1)"><text>x</text></a></svg>`,
			expected: `<svg><a><text>x</text></a></svg>`,
		},
		{
			name:     "DOCTYPE entity expanding to javascript",
			in:       `<!DOCTYPE svg [<!ENTITY j "javascript:alert(1)">]><svg><a href="&j;"><text>x</text></a></svg>`,
			expected: `<!DOCTYPE svg [<!ENTITY j "javascript:alert(1)">]><svg><a><text>x</text></a></svg>`,
		},
		{
			name: "DOCTYPE entity with the reported payload",
			in: `<!DOCTYPE svg [<!ENTITY x "javascript:console.log('PWNED_DTD_REALIMGPROXY');void(0)">]>` +
				`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">` +
				`<a href="&x;"><text x="10" y="50">click dtd</text></a></svg>`,
			expected: `<!DOCTYPE svg [<!ENTITY x "javascript:console.log('PWNED_DTD_REALIMGPROXY');void(0)">]>` +
				`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">` +
				`<a><text x="10" y="50">click dtd</text></a></svg>`,
		},
		{
			name: "DOCTYPE entity in xlink:href",
			in: `<!DOCTYPE svg [<!ENTITY j "javascript:alert(1)">]>` +
				`<svg xmlns:xlink="http://www.w3.org/1999/xlink">` +
				`<a xlink:href="&j;"><text>x</text></a></svg>`,
			expected: `<!DOCTYPE svg [<!ENTITY j "javascript:alert(1)">]>` +
				`<svg xmlns:xlink="http://www.w3.org/1999/xlink"><a><text>x</text></a></svg>`,
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

// TestSanitizeNestedEntities checks that entities are expanded only once:
// a nested reference must not be expanded recursively
func (s *SvgTestSuite) TestSanitizeNestedEntities() {
	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	in := imagedata.NewFromBytesWithFormat(imagetype.SVG, []byte(
		`<!DOCTYPE svg [<!ENTITY a "javascript:"><!ENTITY b "&a;alert(1)">]>`+
			`<svg><a href="&b;"><text>x</text></a></svg>`,
	))

	out, err := svgProc.Process(options.New(), in)
	s.Require().NoError(err)

	data, err := io.ReadAll(out.Reader())
	s.Require().NoError(err)

	_, body, found := bytes.Cut(data, []byte("]>"))
	s.Require().True(found)
	s.Require().NotContains(string(body), "javascript:")
}

// TestSanitizeBillionLaughs checks that the "Billion Laughs" payload
// is not expanded recursively into a huge document
func (s *SvgTestSuite) TestSanitizeBillionLaughs() {
	config := svg.NewDefaultConfig()
	svgProc := svg.New(&config)

	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE svg [<!ENTITY lol "lol">`)

	prev := "lol"
	for i := range 8 {
		name := fmt.Sprintf("lol%d", i+1)
		fmt.Fprintf(&sb, `<!ENTITY %s "%s">`, name, strings.Repeat("&"+prev+";", 10))
		prev = name
	}

	fmt.Fprintf(&sb, `]><svg><text>&%s;</text></svg>`, prev)

	in := imagedata.NewFromBytesWithFormat(imagetype.SVG, []byte(sb.String()))

	out, err := svgProc.Process(options.New(), in)
	s.Require().NoError(err)

	size, err := out.Size()
	s.Require().NoError(err)
	s.Require().Less(size, 4<<10)
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
