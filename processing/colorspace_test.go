package processing

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

type ColorspaceTestSuite struct {
	testSuite
}

var colorspaceTestOutSize = testSize{100, 100}

type colorspaceTestCase struct {
	name          string
	sourceFile    string
	watermarkFile string
	outFormat     imagetype.Type
}

func (tc colorspaceTestCase) Set(o *options.Options) {
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.Width, 100)
	o.Set(keys.Height, 100)
	o.Set(keys.Enlarge, true)
	o.Set(keys.Format, tc.outFormat)

	if tc.watermarkFile != "" {
		o.Set(keys.WatermarkOpacity, 0.5)
	}
}

func (tc colorspaceTestCase) String() string {
	return tc.name
}

func (s *ColorspaceTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ColorspaceTestSuite) runTestCase(tc testCase[colorspaceTestCase]) {
	if tc.opts.watermarkFile != "" {
		s.WatermarkConfig().Path = s.TestData.Path(tc.opts.watermarkFile)
	}

	img, err := s.ImageDataFactory().NewFromPath(s.TestData.Path(tc.opts.sourceFile))
	s.Require().NoError(err)

	o := options.New()
	tc.opts.Set(o)

	s.processImageAndCheck(img, o, tc)
}

func (s *ColorspaceTestSuite) TestColorspace() {
	testCases := []testCase[colorspaceTestCase]{
		{
			opts: colorspaceTestCase{
				name:       "8-bpp-png-srgb",
				sourceFile: "test-images/png/8-bpp.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-png-rgb16",
				sourceFile: "test-images/png/16-bpp.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-png-jpeg-srgb",
				sourceFile: "test-images/png/16-bpp.png",
				outFormat:  imagetype.JPEG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "8-bpp-tiff-srgb",
				sourceFile: "test-images/tiff/8-bpp.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-tiff-rgb16",
				sourceFile: "test-images/tiff/16-bpp.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:       "8-bpp-grayscale-png-bw",
				sourceFile: "test-images/png/8-bpp-grayscale.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-grayscale-png-grey16",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationGrey16,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-grayscale-png-jpeg-bw",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				outFormat:  imagetype.JPEG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:       "8-bpp-grayscale-tiff-bw",
				sourceFile: "test-images/tiff/8-bpp-grayscale.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-grayscale-tiff-grey16",
				sourceFile: "test-images/tiff/16-bpp-grayscale.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationGrey16,
		},
		{
			opts: colorspaceTestCase{
				name:       "jxl-png-rgb16",
				sourceFile: "test-images/jxl/jxl.jxl",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:       "jxl-jpeg-srgb",
				sourceFile: "test-images/jxl/jxl.jxl",
				outFormat:  imagetype.JPEG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.Config().PreserveHDR = true
			s.runTestCase(tc)
		})
	}
}

func (s *ColorspaceTestSuite) TestLinearColorspace() {
	testCases := []testCase[colorspaceTestCase]{
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-linear-png-rgb16",
				sourceFile: "test-images/png/16-bpp-linear.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:       "32-bpp-linear-tiff-rgb16",
				sourceFile: "test-images/tiff/32-bpp-linear.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-linear-png-jpeg-srgb",
				sourceFile: "test-images/png/16-bpp-linear.png",
				outFormat:  imagetype.JPEG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "8-bpp-png-jpeg-srgb",
				sourceFile: "test-images/png/8-bpp.png",
				outFormat:  imagetype.JPEG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String()+"_linear", func() {
			s.ImageMatcher, _ = testutil.NewLazySuiteObj(s, func() (*testutil.ImageHashCacheMatcher, error) {
				return testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeDifference), nil
			})
			s.Config().PreserveHDR = true
			s.Config().UseLinearColorspace = true
			s.runTestCase(tc)
		})
	}
}

func (s *ColorspaceTestSuite) TestDownscaleHDR() {
	testCases := []testCase[colorspaceTestCase]{
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-png-srgb",
				sourceFile: "test-images/png/16-bpp.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-tiff-srgb",
				sourceFile: "test-images/tiff/16-bpp.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-grayscale-png-bw",
				sourceFile: "test-images/png/16-bpp-grayscale.png",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:       "16-bpp-grayscale-tiff-bw",
				sourceFile: "test-images/tiff/16-bpp-grayscale.tiff",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:       "jxl-png-srgb",
				sourceFile: "test-images/jxl/jxl.jxl",
				outFormat:  imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String()+"_no_preserve_hdr", func() {
			s.Config().PreserveHDR = false
			s.runTestCase(tc)
		})
	}
}

func (s *ColorspaceTestSuite) TestWatermarkColorspace() {
	testCases := []testCase[colorspaceTestCase]{
		{
			opts: colorspaceTestCase{
				name:          "8-bpp-wm-16-bpp-srgb",
				sourceFile:    "test-images/png/8-bpp.png",
				watermarkFile: "test-images/png/16-bpp.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:          "8-bpp-wm-16-bpp-grayscale-srgb",
				sourceFile:    "test-images/png/8-bpp.png",
				watermarkFile: "test-images/png/16-bpp-grayscale.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
		{
			opts: colorspaceTestCase{
				name:          "8-bpp-grayscale-wm-16-bpp-grayscale-bw",
				sourceFile:    "test-images/png/8-bpp-grayscale.png",
				watermarkFile: "test-images/png/16-bpp-grayscale.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationBW,
		},
		{
			opts: colorspaceTestCase{
				name:          "16-bpp-grayscale-wm-8-bpp-grayscale-grey16",
				sourceFile:    "test-images/png/16-bpp-grayscale.png",
				watermarkFile: "test-images/png/8-bpp-grayscale.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationGrey16,
		},
		{
			opts: colorspaceTestCase{
				name:          "16-bpp-grayscale-wm-16-bpp-rgb16",
				sourceFile:    "test-images/png/16-bpp-grayscale.png",
				watermarkFile: "test-images/png/16-bpp.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: colorspaceTestCase{
				name:          "16-bpp-grayscale-wm-8-bpp-rgb16",
				sourceFile:    "test-images/png/16-bpp-grayscale.png",
				watermarkFile: "test-images/png/8-bpp.png",
				outFormat:     imagetype.PNG,
			},
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.ImageMatcher, _ = testutil.NewLazySuiteObj(s, func() (*testutil.ImageHashCacheMatcher, error) {
				return testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypePerception), nil
			})

			s.Config().PreserveHDR = true
			s.runTestCase(tc)
		})
	}
}

func TestColorspace(t *testing.T) {
	suite.Run(t, new(ColorspaceTestSuite))
}
