package processing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/stretchr/testify/suite"
)

type ExtendTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type extendTestCase struct {
	gravity GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (e extendTestCase) Set(o *options.Options) {
	o.Set(keys.ExtendEnabled, true)
	o.Set(keys.ExtendGravityType, e.gravity)
	o.Set(keys.Width, e.size.width)
	o.Set(keys.Height, e.size.height)

	if e.xOffset != 0 {
		o.Set(keys.ExtendGravityXOffset, e.xOffset)
	} else {
		o.Delete(keys.ExtendGravityXOffset)
	}

	if e.yOffset != 0 {
		o.Set(keys.ExtendGravityYOffset, e.yOffset)
	} else {
		o.Delete(keys.ExtendGravityYOffset)
	}

	if e.dpr != 0 {
		o.Set(keys.Dpr, e.dpr)
	} else {
		o.Delete(keys.Dpr)
	}
}

func (e extendTestCase) String() string {
	var b bytes.Buffer

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
	gravity GravityType
	size    testSize
	xOffset float64
	yOffset float64
	dpr     float64
}

func (e extendArTestCase) Set(o *options.Options) {
	o.Set(keys.ExtendAspectRatioEnabled, true)
	o.Set(keys.ExtendAspectRatioGravityType, e.gravity)
	o.Set(keys.Width, e.size.width)
	o.Set(keys.Height, e.size.height)

	if e.xOffset != 0 {
		o.Set(keys.ExtendAspectRatioGravityXOffset, e.xOffset)
	} else {
		o.Delete(keys.ExtendAspectRatioGravityXOffset)
	}

	if e.yOffset != 0 {
		o.Set(keys.ExtendAspectRatioGravityYOffset, e.yOffset)
	} else {
		o.Delete(keys.ExtendAspectRatioGravityYOffset)
	}

	if e.dpr != 0 {
		o.Set(keys.Dpr, e.dpr)
	} else {
		o.Delete(keys.Dpr)
	}
}

func (e extendArTestCase) String() string {
	var b bytes.Buffer

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

func (s *ExtendTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("geometry.png"),
	)
	s.Require().NoError(err)
}

func (s *ExtendTestSuite) TestExtend() {
	o := options.New()

	extendSize := testSize{500, 500}

	//nolint:dupl
	testCases := []testCase[extendTestCase]{
		{
			opts: extendTestCase{
				gravity: GravityCenter,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravitySouth,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityNorthEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravitySouthEast,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravitySouthWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityNorthWest,
				size:    extendSize,
			},
			outSize: extendSize,
		},

		// With offsets
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 5,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 5,
				dpr:     2,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityEast,
				size:    extendSize,
				xOffset: 5,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityEast,
				size:    extendSize,
				xOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: GravityEast,
				size:    extendSize,
				xOffset: 5,
				dpr:     2,
			},
			outSize: extendSize,
		},

		// With relative offsets
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
			},
			outSize: extendSize,
		},
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
				dpr:     0.5,
			},
			outSize: testSize{extendSize.width / 2, extendSize.height / 2},
		},
		{
			opts: extendTestCase{
				gravity: GravityNorth,
				size:    extendSize,
				yOffset: 0.1,
				dpr:     2,
			},
			outSize: extendSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc)
		})
	}
}

func (s *ExtendTestSuite) TestExtendAspectRatio() {
	o := options.New()

	targetSize := testSize{300, 600}
	expectedSize := testSize{200, 400}

	//nolint:dupl
	testCases := []testCase[extendArTestCase]{
		{
			opts: extendArTestCase{
				gravity: GravityCenter,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravitySouth,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorthEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravitySouthEast,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravitySouthWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorthWest,
				size:    targetSize,
			},
			outSize: expectedSize,
		},

		// With offsets
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 5,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 5,
				dpr:     2,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityEast,
				size:    targetSize,
				xOffset: 5,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityEast,
				size:    targetSize,
				xOffset: 5,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: GravityEast,
				size:    targetSize,
				xOffset: 5,
				dpr:     2,
			},
			outSize: expectedSize,
		},

		// With relative offsets
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
			},
			outSize: expectedSize,
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
				dpr:     0.5,
			},
			outSize: testSize{targetSize.width / 2, targetSize.height / 2},
		},
		{
			opts: extendArTestCase{
				gravity: GravityNorth,
				size:    targetSize,
				yOffset: 0.1,
				dpr:     2,
			},
			outSize: expectedSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc)
		})
	}
}

func TestExtend(t *testing.T) {
	suite.Run(t, new(ExtendTestSuite))
}
