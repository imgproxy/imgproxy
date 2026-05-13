package processing_test

import (
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips"
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

func (tc colorspaceTestCase) ImagePath() string {
	return tc.sourceFile
}

func (tc colorspaceTestCase) URLOptions() string {
	opts := fmt.Sprintf(
		"resize:fill:%d:%d/enlarge:1/format:%s",
		colorspaceTestOutSize.width,
		colorspaceTestOutSize.height,
		tc.outFormat,
	)
	if tc.watermarkFile != "" {
		opts += "/watermark:0.5"
	}
	return opts
}

type preserveHDRTestCase struct {
	name        string
	sourceFile  string
	outFormat   imagetype.Type
	preserveHDR bool // URL option value
}

func (tc preserveHDRTestCase) ImagePath() string {
	return tc.sourceFile
}

func (tc preserveHDRTestCase) URLOptions() string {
	return fmt.Sprintf(
		"resize:fill:%d:%d/enlarge:1/format:%s/preserve_hdr:%t",
		colorspaceTestOutSize.width,
		colorspaceTestOutSize.height,
		tc.outFormat,
		tc.preserveHDR,
	)
}

func (s *ColorspaceTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ColorspaceTestSuite) runTestCase(tc testCase[colorspaceTestCase]) {
	if tc.opts.watermarkFile != "" {
		s.Config().WatermarkImage.Path = s.TestData.Path(tc.opts.watermarkFile)
	}

	s.processImageAndCheck(tc)
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
	}

	for _, tc := range testCases {
		s.Run(tc.opts.name, func() {
			s.Config().Processing.PreserveHDR = true
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
		s.Run(tc.opts.name+"_linear", func() {
			s.ImageMatcher, _ = testutil.NewLazySuiteObj(s, func() (*testutil.ImageHashCacheMatcher, error) {
				return testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeDct), nil
			})
			s.Config().Processing.PreserveHDR = true
			s.Config().Processing.UseLinearColorspace = true
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
	}

	for _, tc := range testCases {
		s.Run(tc.opts.name+"_no_preserve_hdr", func() {
			s.Config().Processing.PreserveHDR = false
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
		s.Run(tc.opts.name, func() {
			s.ImageMatcher, _ = testutil.NewLazySuiteObj(s, func() (*testutil.ImageHashCacheMatcher, error) {
				return testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeDct), nil
			})

			s.Config().Processing.PreserveHDR = true
			s.runTestCase(tc)
		})
	}
}

func (s *ColorspaceTestSuite) TestPreserveHDROptionOverride() {
	testCases := []struct {
		opts              preserveHDRTestCase
		configValue       bool
		outSize           testSize
		outInterpretation vips.Interpretation
	}{
		{
			opts: preserveHDRTestCase{
				name:        "option-true-overrides-config-false",
				sourceFile:  "test-images/png/16-bpp.png",
				outFormat:   imagetype.PNG,
				preserveHDR: true,
			},
			configValue:       false,
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationRGB16,
		},
		{
			opts: preserveHDRTestCase{
				name:        "option-false-overrides-config-true",
				sourceFile:  "test-images/png/16-bpp.png",
				outFormat:   imagetype.PNG,
				preserveHDR: false,
			},
			configValue:       true,
			outSize:           colorspaceTestOutSize,
			outInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.name, func() {
			s.Config().Processing.PreserveHDR = tc.configValue

			testCase := testCase[preserveHDRTestCase]{
				opts:              tc.opts,
				outSize:           tc.outSize,
				outInterpretation: tc.outInterpretation,
			}

			s.processImageAndCheck(testCase)
		})
	}
}

func TestColorspace(t *testing.T) {
	suite.Run(t, new(ColorspaceTestSuite))
}
