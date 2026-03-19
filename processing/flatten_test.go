package processing_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type FlattenTestSuite struct {
	testSuite
}

var flattenTestOutSize = testSize{500, 500}

type flattenTestCase struct {
	name       string
	sourceFile string
	background *color.RGB
	format     imagetype.Type
}

func (c flattenTestCase) ImagePath() string {
	return c.sourceFile
}

func (c flattenTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	opts.Add("resize").Set(0, "fill").Set(1, 300).Set(2, 300)
	opts.Add("enlarge").Set(0, 1)
	opts.Add("padding").Set(0, 100)
	opts.Add("format").Set(0, c.format)

	if c.background != nil {
		opts.Add("background").
			Set(0, c.background.R).
			Set(1, c.background.G).
			Set(2, c.background.B)
	}

	return opts.String()
}

func (s *FlattenTestSuite) TestBackground() {
	testCases := []testCase[flattenTestCase]{
		// Basic background tests with 32-bpp-with-alpha.bmp
		{
			opts: flattenTestCase{
				name:       "32-bpp-red-jpeg",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: &color.Red,
				format:     imagetype.JPEG,
			},
			outSize: flattenTestOutSize,
		},
		{
			opts: flattenTestCase{
				name:       "32-bpp-red-png",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: &color.Red,
				format:     imagetype.PNG,
			},
			outSize: flattenTestOutSize,
		},
		{
			opts: flattenTestCase{
				name:       "32-bpp-none-jpeg",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: nil,
				format:     imagetype.JPEG,
			},
			outSize: flattenTestOutSize,
		},
		{
			opts: flattenTestCase{
				name:       "32-bpp-none-png",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: nil,
				format:     imagetype.PNG,
			},
			outSize: flattenTestOutSize,
		},
		// RGB16 source should stay RGB16
		{
			opts: flattenTestCase{
				name:       "16-bpp-gray-rgb16",
				sourceFile: "test-images/png/16-bpp.png",
				background: &color.Gray,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: flattenTestCase{
				name:       "16-bpp-red-rgb16",
				sourceFile: "test-images/png/16-bpp.png",
				background: &color.Red,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		// 8-bit grayscale source stays BW with gray color, becomes sRGB with red color
		{
			opts: flattenTestCase{
				name:       "8-bpp-grayscale-gray-bw",
				sourceFile: "test-images/png/8-bpp-grayscale.png",
				background: &color.Gray,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: flattenTestCase{
				name:       "8-bpp-grayscale-red-srgb",
				sourceFile: "test-images/png/8-bpp-grayscale.png",
				background: &color.Red,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		// 16-bit grayscale source stays Grey16 with gray color, becomes RGB16 with red color
		{
			opts: flattenTestCase{
				name:       "16-bpp-grayscale-gray-grey16",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				background: &color.Gray,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationGrey16,
		},
		{
			opts: flattenTestCase{
				name:       "16-bpp-grayscale-red-rgb16",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				background: &color.Red,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		// Regular 8-bit RGB source stays sRGB
		{
			opts: flattenTestCase{
				name:       "8-bpp-gray-srgb",
				sourceFile: "test-images/png/8-bpp.png",
				background: &color.Gray,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: flattenTestCase{
				name:       "8-bpp-red-srgb",
				sourceFile: "test-images/png/8-bpp.png",
				background: &color.Red,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.name, func() {
			s.Config().Processing.PreserveHDR = true

			s.processImageAndCheck(tc)
		})
	}
}

func TestFlatten(t *testing.T) {
	suite.Run(t, new(FlattenTestSuite))
}
