package processing_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type CropTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type resizeFillGravityTestCase struct {
	gravity processing.GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (r resizeFillGravityTestCase) Set(o *options.Options) {
	o.Set(keys.ResizingType, processing.ResizeFill)
	o.Set(keys.GravityType, r.gravity)
	o.Set(keys.Width, r.size.width)
	o.Set(keys.Height, r.size.height)

	if r.xOffset != 0 {
		o.Set(keys.GravityXOffset, r.xOffset)
	} else {
		o.Delete(keys.GravityXOffset)
	}

	if r.yOffset != 0 {
		o.Set(keys.GravityYOffset, r.yOffset)
	} else {
		o.Delete(keys.GravityYOffset)
	}

	if r.dpr != 0 {
		o.Set(keys.Dpr, r.dpr)
	} else {
		o.Delete(keys.Dpr)
	}
}

func (r resizeFillGravityTestCase) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "resizeFill_%dx%d", r.size.width, r.size.height)

	if r.gravity != 0 {
		fmt.Fprintf(&b, "_gravity_%s", r.gravity.String())
	}

	if r.xOffset != 0 || r.yOffset != 0 {
		fmt.Fprintf(&b, "_offset_%f_%f", r.xOffset, r.yOffset)
	}

	if r.dpr > 0 {
		fmt.Fprintf(&b, "_dpr_%g", r.dpr)
	}

	return b.String()
}

type cropTestCase struct {
	gravity     processing.GravityType
	cropGravity processing.GravityType
	cropSize    testSize
	xOffset     float64
	yOffset     float64
	dpr         float64
}

func (c cropTestCase) Set(o *options.Options) {
	o.Set(keys.GravityType, c.gravity)
	o.Set(keys.CropGravityType, c.cropGravity)
	o.Set(keys.CropWidth, c.cropSize.width)
	o.Set(keys.CropHeight, c.cropSize.height)

	if c.xOffset != 0 {
		o.Set(keys.CropGravityXOffset, c.xOffset)
	} else {
		o.Delete(keys.CropGravityXOffset)
	}

	if c.yOffset != 0 {
		o.Set(keys.CropGravityYOffset, c.yOffset)
	} else {
		o.Delete(keys.CropGravityYOffset)
	}

	if c.dpr != 0 {
		o.Set(keys.Dpr, c.dpr)
	} else {
		o.Delete(keys.Dpr)
	}
}

func (c cropTestCase) String() string {
	var b bytes.Buffer

	if c.gravity != 0 {
		fmt.Fprintf(&b, "_gravity_%s", c.gravity.String())
	}

	if c.cropGravity != 0 {
		fmt.Fprintf(&b, "_crop_gravity_%s", c.cropGravity.String())
	}

	if c.cropSize.width > 0 && c.cropSize.height > 0 {
		fmt.Fprintf(&b, "_crop_%dx%d", c.cropSize.width, c.cropSize.height)
	}

	if c.xOffset != 0 || c.yOffset != 0 {
		fmt.Fprintf(&b, "_offset_%f_%f", c.xOffset, c.yOffset)
	}

	if c.dpr > 0 {
		fmt.Fprintf(&b, "_dpr_%g", c.dpr)
	}

	n, _ := strings.CutPrefix(b.String(), "_")
	return n
}

func (s *CropTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("geometry.png"),
	)
	s.Require().NoError(err)
}
func (s *CropTestSuite) TestResizeFill() {
	o := options.New()

	widerSize := testSize{100, 26}
	tallerSize := testSize{26, 100}

	testCases := []testCase[resizeFillGravityTestCase]{
		// When target has less vertical space
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityCenter,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravitySouth,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorthEast,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravitySouthEast,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravitySouthWest,
				size:    widerSize,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorthWest,
				size:    widerSize,
			},
			outSize: widerSize,
		},

		// When target has less horizontal space
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityCenter,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityEast,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityWest,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorthEast,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravitySouthEast,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravitySouthWest,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorthWest,
				size:    tallerSize,
			},
			outSize: tallerSize,
		},

		// With DPR
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityCenter,
				size:    widerSize,
				dpr:     2,
			},
			outSize: testSize{widerSize.width * 2, widerSize.height * 2},
		},

		// With offsets
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 5,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{width: widerSize.width / 2, height: widerSize.height / 2},
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 5,
				dpr:     2,
			},
			outSize: testSize{width: widerSize.width * 2, height: widerSize.height * 2},
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    tallerSize,
				xOffset: 5,
			},
			outSize: tallerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    tallerSize,
				xOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{width: tallerSize.width / 2, height: tallerSize.height / 2},
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    tallerSize,
				xOffset: 5,
				dpr:     2,
			},
			outSize: tallerSize, // dpr does not affect output size since it higher than input
		},

		// With relative offsets
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 0.1,
			},
			outSize: widerSize,
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 0.1,
				dpr:     0.5,
			},
			outSize: testSize{width: widerSize.width / 2, height: 13},
		},
		{
			opts: resizeFillGravityTestCase{
				gravity: processing.GravityNorth,
				size:    widerSize,
				yOffset: 0.1,
				dpr:     2,
			},
			outSize: testSize{width: widerSize.width * 2, height: widerSize.height * 2},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc)
		})
	}
}

func (s *CropTestSuite) TestCrop() {
	o := options.New()

	cropSize := testSize{50, 50}

	testCases := []testCase[cropTestCase]{
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityNorth,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravitySouth,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityEast,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityWest,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityNorthEast,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityNorthWest,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravitySouthEast,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravitySouthWest,
				cropSize:    cropSize,
			},
			outSize: cropSize,
		},

		// Crop with offsets
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				xOffset:     10,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				yOffset:     10,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				xOffset:     10,
				yOffset:     10,
			},
			outSize: cropSize,
		},

		// Relative offsets
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				xOffset:     0.1,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				yOffset:     0.1,
			},
			outSize: cropSize,
		},
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				xOffset:     0.1,
				yOffset:     0.1,
			},
			outSize: cropSize,
		},

		// Dpr
		{
			opts: cropTestCase{
				cropGravity: processing.GravityCenter,
				cropSize:    cropSize,
				xOffset:     10,
				yOffset:     10,
				dpr:         0.5,
			},
			outSize: testSize{25, 25}, // since dpr 0.5 halves the size
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc)
		})
	}
}

func (s *CropTestSuite) getCropGravityResult(c cropTestCase) imagedata.ImageData {
	o := options.New()
	c.Set(o)

	result, err := s.Processor().ProcessImage(s.T().Context(), s.img, o)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	return result.OutData
}

func (s *CropTestSuite) TestCropGravityPriority() {
	cropSize := testSize{50, 50}

	r1 := s.getCropGravityResult(cropTestCase{
		gravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r1.Close()

	r2 := s.getCropGravityResult(cropTestCase{
		cropGravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r2.Close()

	r3 := s.getCropGravityResult(cropTestCase{
		gravity: processing.GravitySouth, cropGravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r3.Close()

	s.Require().True(testutil.ReadersEqual(s.T(), r1.Reader(), r2.Reader()))
	s.Require().True(testutil.ReadersEqual(s.T(), r1.Reader(), r3.Reader()))
}

func TestCrop(t *testing.T) {
	suite.Run(t, new(CropTestSuite))
}
