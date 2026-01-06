package processing

import (
	"fmt"
	"path/filepath"
	"strings"
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

type colorspaceTestCase struct {
	sourceFile             string
	watermarkFile          string
	outFormat              imagetype.Type
	expectedInterpretation vips.Interpretation
}

func (tc colorspaceTestCase) testName() string {
	fileName := strings.ReplaceAll(filepath.Base(tc.sourceFile), ".", "_")
	name := fmt.Sprintf("%s_%s_%d", fileName, tc.outFormat, tc.expectedInterpretation)
	if tc.watermarkFile != "" {
		watermarkName := strings.ReplaceAll(filepath.Base(tc.watermarkFile), ".", "_")
		name = fmt.Sprintf("%s_wm_%s", name, watermarkName)
	}
	return name
}

func (s *ColorspaceTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ColorspaceTestSuite) runTestCase(
	tc colorspaceTestCase,
	imageMatcher *testutil.ImageHashCacheMatcher,
	distance int,
) {
	// Load source image
	img, err := s.ImageDataFactory().NewFromPath(
		s.TestData.Path(tc.sourceFile),
	)
	s.Require().NoError(err)

	// Create options with resize to 100x100 and enlarge enabled
	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.Width, 100)
	o.Set(keys.Height, 100)
	o.Set(keys.Enlarge, true)
	o.Set(keys.Format, tc.outFormat)

	if tc.watermarkFile != "" {
		o.Set(keys.WatermarkOpacity, 0.5)
		s.WatermarkConfig().Path = s.TestData.Path(tc.watermarkFile)
	}

	// Process the image
	result, err := s.Processor().ProcessImage(s.T().Context(), img, o)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	// Load the result image to check its interpretation
	resultImg := new(vips.Image)
	defer resultImg.Clear()

	err = resultImg.Load(result.OutData, 1, 1.0, 1)
	s.Require().NoError(err)

	// Check the interpretation
	actualInterpretation := resultImg.Type()
	s.Require().Equal(tc.expectedInterpretation, actualInterpretation)

	// Match against stored hash
	imageMatcher.ImageMatches(s.T(), result.OutData.Reader(), "hash", distance)
}

func (s *ColorspaceTestSuite) TestColorspace() {
	testCases := []colorspaceTestCase{
		{
			sourceFile:             "test-images/png/8-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/png/16-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/png/16-bpp.png",
			outFormat:              imagetype.JPEG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/tiff/8-bpp.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/tiff/16-bpp.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/png/8-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationGrey16,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			outFormat:              imagetype.JPEG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/tiff/8-bpp-grayscale.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/tiff/16-bpp-grayscale.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationGrey16,
		},
		{
			sourceFile:             "test-images/jxl/jxl.jxl",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/jxl/jxl.jxl",
			outFormat:              imagetype.JPEG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.testName(), func() {
			s.Config().PreserveHDR = true
			s.runTestCase(tc, s.ImageMatcher, 0)
		})
	}
}

func (s *ColorspaceTestSuite) TestLinearColorspace() {
	imageMatcher := testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeDifference)

	testCases := []colorspaceTestCase{
		{
			sourceFile:             "test-images/png/16-bpp-linear.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/tiff/32-bpp-linear.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/png/16-bpp-linear.png",
			outFormat:              imagetype.JPEG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/png/8-bpp.png",
			outFormat:              imagetype.JPEG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.testName()+"_linear", func() {
			s.Config().PreserveHDR = true
			s.Config().UseLinearColorspace = true
			s.runTestCase(tc, imageMatcher, 1)
		})
	}
}

func (s *ColorspaceTestSuite) TestDownscaleHDR() {
	testCases := []colorspaceTestCase{
		{
			sourceFile:             "test-images/png/16-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/tiff/16-bpp.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/tiff/16-bpp-grayscale.tiff",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/jxl/jxl.jxl",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.testName()+"_no_preserve_hdr", func() {
			s.Config().PreserveHDR = false
			s.runTestCase(tc, s.ImageMatcher, 0)
		})
	}
}

func (s *ColorspaceTestSuite) TestWatermarkColorspace() {
	imageMatcher := testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypePerception)

	testCases := []colorspaceTestCase{
		{
			sourceFile:             "test-images/png/8-bpp.png",
			watermarkFile:          "test-images/png/16-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/png/8-bpp.png",
			watermarkFile:          "test-images/png/16-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationSRGB,
		},
		{
			sourceFile:             "test-images/png/8-bpp-grayscale.png",
			watermarkFile:          "test-images/png/16-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationBW,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			watermarkFile:          "test-images/png/8-bpp-grayscale.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationGrey16,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			watermarkFile:          "test-images/png/16-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
		{
			sourceFile:             "test-images/png/16-bpp-grayscale.png",
			watermarkFile:          "test-images/png/8-bpp.png",
			outFormat:              imagetype.PNG,
			expectedInterpretation: vips.InterpretationRGB16,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.testName(), func() {
			s.Config().PreserveHDR = true
			s.runTestCase(tc, imageMatcher, 1)
		})
	}
}

func TestColorspace(t *testing.T) {
	suite.Run(t, new(ColorspaceTestSuite))
}
