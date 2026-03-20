package processing_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type CropTestSuite struct {
	testSuite
}

type resizeFillGravityTestCase struct {
	gravity processing.GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (r resizeFillGravityTestCase) ImagePath() string {
	return "geometry.png"
}

func (r resizeFillGravityTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	opts.Add("rs").
		Set(0, "fill").
		Set(1, r.size.width).
		Set(2, r.size.height)

	if r.gravity != processing.GravityUnknown {
		args := opts.Add("g").Set(0, r.gravity)

		if r.xOffset != 0 {
			args.Set(1, r.xOffset)
		}
		if r.yOffset != 0 {
			args.Set(2, r.yOffset)
		}
	}

	if r.dpr != 0 {
		opts.Add("dpr").Set(0, r.dpr)
	}

	return opts.String()
}

func (r resizeFillGravityTestCase) String() string {
	var b strings.Builder

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

func (c cropTestCase) ImagePath() string {
	return "geometry.png"
}

func (c cropTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	args := opts.Add("crop").
		Set(0, c.cropSize.width).
		Set(1, c.cropSize.height)

	if c.cropGravity != processing.GravityUnknown {
		args.Set(2, c.cropGravity)

		if c.xOffset != 0 {
			args.Set(3, c.xOffset)
		}
		if c.yOffset != 0 {
			args.Set(4, c.yOffset)
		}
	}

	if c.gravity != processing.GravityUnknown {
		opts.Add("gravity").Set(0, c.gravity)
	}

	if c.dpr != 0 {
		opts.Add("dpr").Set(0, c.dpr)
	}

	return opts.String()
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

func (s *CropTestSuite) TestResizeFill() {
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
			s.processImageAndCheck(tc)
		})
	}
}

func (s *CropTestSuite) TestCrop() {
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
			s.processImageAndCheck(tc)
		})
	}
}

func (s *CropTestSuite) TestCropGravityPriority() {
	cropSize := testSize{50, 50}

	r1 := s.processImage(cropTestCase{
		gravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r1.Close()

	r2 := s.processImage(cropTestCase{
		cropGravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r2.Close()

	r3 := s.processImage(cropTestCase{
		gravity: processing.GravitySouth, cropGravity: processing.GravityNorth, cropSize: cropSize,
	})
	defer r3.Close()

	s.Require().True(testutil.ReadersEqual(s.T(), r1.Reader(), r2.Reader()))
	s.Require().True(testutil.ReadersEqual(s.T(), r1.Reader(), r3.Reader()))
}

func TestCrop(t *testing.T) {
	suite.Run(t, new(CropTestSuite))
}
