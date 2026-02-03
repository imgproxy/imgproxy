package processing

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type FlattenTestSuite struct {
	testSuite

	img imagedata.ImageData
}

var flattenTestOutSize = testSize{500, 500}

type flattenTestCase struct {
	name       string
	sourceFile string
	background *color.RGB
	format     imagetype.Type
}

func (r flattenTestCase) Set(o *options.Options) {
	if r.background != nil {
		o.Set(keys.Background, *r.background)
	} else {
		o.Delete(keys.Background)
	}

	o.Set(keys.Format, r.format)
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.Width, 300)
	o.Set(keys.Height, 300)
	o.Set(keys.Enlarge, true)
	o.Set(keys.PaddingLeft, 100)
	o.Set(keys.PaddingRight, 100)
	o.Set(keys.PaddingTop, 100)
	o.Set(keys.PaddingBottom, 100)
}

func (r flattenTestCase) String() string {
	return r.name
}

func (s *FlattenTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("test-images/bmp/32-bpp-with-alpha.bmp"),
	)
	s.Require().NoError(err)
}

func (s *FlattenTestSuite) TestBackground() {
	var (
		grayColor = &color.RGB{R: 127, G: 127, B: 127}
		redColor  = &color.RGB{R: 255, G: 0, B: 0}
	)

	testCases := []testCase[flattenTestCase]{
		// Basic background tests with 32-bpp-with-alpha.bmp
		{
			opts: flattenTestCase{
				name:       "32-bpp-red-jpeg",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: redColor,
				format:     imagetype.JPEG,
			},
			outSize: flattenTestOutSize,
		},
		{
			opts: flattenTestCase{
				name:       "32-bpp-red-png",
				sourceFile: "test-images/bmp/32-bpp-with-alpha.bmp",
				background: redColor,
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
				background: grayColor,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: flattenTestCase{
				name:       "16-bpp-red-rgb16",
				sourceFile: "test-images/png/16-bpp.png",
				background: redColor,
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
				background: grayColor,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: flattenTestCase{
				name:       "8-bpp-grayscale-red-srgb",
				sourceFile: "test-images/png/8-bpp-grayscale.png",
				background: redColor,
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
				background: grayColor,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationGrey16,
		},
		{
			opts: flattenTestCase{
				name:       "16-bpp-grayscale-red-rgb16",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				background: redColor,
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
				background: grayColor,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: flattenTestCase{
				name:       "8-bpp-red-srgb",
				sourceFile: "test-images/png/8-bpp.png",
				background: redColor,
				format:     imagetype.PNG,
			},
			outSize:           flattenTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.Config().PreserveHDR = true

			img, err := s.ImageDataFactory().NewFromPath(s.TestData.Path(tc.opts.sourceFile))
			s.Require().NoError(err)

			o := options.New()
			tc.opts.Set(o)

			s.processImageAndCheck(img, o, tc)
		})
	}
}

func TestFlatten(t *testing.T) {
	suite.Run(t, new(FlattenTestSuite))
}
