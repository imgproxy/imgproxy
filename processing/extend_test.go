package processing_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ExtendTestSuite struct {
	testSuite
}

type extendTestCase struct {
	gravity processing.GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (e extendTestCase) ImagePath() string {
	return "geometry.png"
}

func (e extendTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	opts.Add("width").Set(0, e.size.width)
	opts.Add("height").Set(0, e.size.height)

	args := opts.Add("extend").Set(0, 1)

	if e.gravity != processing.GravityUnknown {
		args.Set(1, e.gravity)

		if e.xOffset != 0 {
			args.Set(2, e.xOffset)
		}
		if e.yOffset != 0 {
			args.Set(3, e.yOffset)
		}
	}

	if e.dpr != 0 {
		opts.Add("dpr").Set(0, e.dpr)
	}

	return opts.String()
}

func (e extendTestCase) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "extend_%dx%d", e.size.width, e.size.height)

	if e.gravity != 0 {
		fmt.Fprintf(&b, "_gravity_%s", e.gravity.String())
	}

	if e.xOffset != 0 || e.yOffset != 0 {
		fmt.Fprintf(&b, "_offset_%f_%f", e.xOffset, e.yOffset)
	}

	if e.dpr > 0 {
		fmt.Fprintf(&b, "_dpr_%g", e.dpr)
	}

	return b.String()
}

type extendArTestCase struct {
	gravity processing.GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (e extendArTestCase) ImagePath() string {
	return "geometry.png"
}

func (e extendArTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	opts.Add("width").Set(0, e.size.width)
	opts.Add("height").Set(0, e.size.height)

	args := opts.Add("extend_aspect_ratio").Set(0, 1)

	if e.gravity != processing.GravityUnknown {
		args.Set(1, e.gravity)

		if e.xOffset != 0 {
			args.Set(2, e.xOffset)
		}
		if e.yOffset != 0 {
			args.Set(3, e.yOffset)
		}
	}

	if e.dpr != 0 {
		opts.Add("dpr").Set(0, e.dpr)
	}

	return opts.String()
}

func (e extendArTestCase) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "extendAr_%dx%d", e.size.width, e.size.height)

	if e.gravity != 0 {
		fmt.Fprintf(&b, "_gravity_%s", e.gravity.String())
	}

	if e.xOffset != 0 || e.yOffset != 0 {
		fmt.Fprintf(&b, "_offset_%f_%f", e.xOffset, e.yOffset)
	}

	if e.dpr > 0 {
		fmt.Fprintf(&b, "_dpr_%g", e.dpr)
	}

	return b.String()
}

func (s *ExtendTestSuite) TestExtend() {
	extendSize := testSize{500, 500}

	//nolint:dupl
	testCases := []testCase[extendTestCase]{
		{
			opts: extendTestCase{
				gravity: processing.GravityCenter,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravitySouth,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorthEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravitySouthEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravitySouthWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorthWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},

		// With offsets
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 5,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 5,
				dpr:     2,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityEast,
				size:    extendSize,
				xOffset: 5,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityEast,
				size:    extendSize,
				xOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityEast,
				size:    extendSize,
				xOffset: 5,
				dpr:     2,
			},
			outSize: extendSize,
		},

		// With relative offsets
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: processing.GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
				dpr:     2,
			},
			outSize: extendSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.processImageAndCheck(tc)
		})
	}
}

func (s *ExtendTestSuite) TestExtendAspectRatio() {
	targetSize := testSize{300, 600}
	expectedSize := testSize{200, 400}

	//nolint:dupl
	testCases := []testCase[extendArTestCase]{
		{
			opts: extendArTestCase{
				gravity: processing.GravityCenter,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravitySouth,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorthEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravitySouthEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravitySouthWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorthWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},

		// With offsets
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 5,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 5,
				dpr:     2,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityEast,
				size:    targetSize,
				xOffset: 5,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityEast,
				size:    targetSize,
				xOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityEast,
				size:    targetSize,
				xOffset: 5,
				dpr:     2,
			},
			outSize: expectedSize,
		},

		// With relative offsets
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: processing.GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
				dpr:     2,
			},
			outSize: expectedSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.processImageAndCheck(tc)
		})
	}
}

func TestExtend(t *testing.T) {
	suite.Run(t, new(ExtendTestSuite))
}
