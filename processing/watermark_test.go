package processing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/stretchr/testify/suite"
)

type watermarkTestCase struct {
	position GravityType
	opacity  float64
	xOffset  float64
	yOffset  float64
	scale    float64
	dpr      float64
}

func (w watermarkTestCase) String() string {
	var b bytes.Buffer

	b.WriteString(w.position.String())
	fmt.Fprintf(&b, "_opacity_%g", w.opacity)

	if w.xOffset != 0 || w.yOffset != 0 {
		fmt.Fprintf(&b, "_offset_%g_%g", w.xOffset, w.yOffset)
	}

	if w.scale != 0 {
		fmt.Fprintf(&b, "_scale_%g", w.scale)
	}

	if w.dpr != 0 {
		fmt.Fprintf(&b, "_dpr_%g", w.dpr)
	}

	return b.String()
}

func (w watermarkTestCase) Set(o *options.Options) {
	o.Set(keys.WatermarkPosition, w.position)

	if w.opacity != 0 {
		o.Set(keys.WatermarkOpacity, w.opacity)
	} else {
		o.Delete(keys.WatermarkOpacity)
	}

	if w.xOffset != 0 {
		o.Set(keys.WatermarkXOffset, w.xOffset)
	} else {
		o.Delete(keys.WatermarkXOffset)
	}

	if w.yOffset != 0 {
		o.Set(keys.WatermarkYOffset, w.yOffset)
	} else {
		o.Delete(keys.WatermarkYOffset)
	}

	if w.scale != 0 {
		o.Set(keys.WatermarkScale, w.scale)
	} else {
		o.Delete(keys.WatermarkScale)
	}

	if w.dpr != 0 {
		o.Set(keys.Dpr, w.dpr)
	} else {
		o.Delete(keys.Dpr)
	}
}

type WatermarkTestSuite struct {
	testSuite

	img     imagedata.ImageData
	imgAnim imagedata.ImageData
}

func (s *WatermarkTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("test-images/bmp/24-bpp.bmp"),
	)
	s.Require().NoError(err)

	s.imgAnim, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("test-images/gif/gif.gif"),
	)
	s.Require().NoError(err)
}

func (s *WatermarkTestSuite) TestWatermark() {
	o := options.New()
	o.Set(keys.Format, imagetype.PNG)

	outSize := testSize{1080, 902}

	testCases := []testCase[watermarkTestCase]{
		// All positions
		{
			opts:    watermarkTestCase{position: GravityCenter, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityNorth, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravitySouth, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityNorthWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityNorthEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravitySouthWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravitySouthEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: GravityReplicate, opacity: 1},
			outSize: outSize,
		},

		// Offset
		{
			opts: watermarkTestCase{
				position: GravityNorth,
				xOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityNorth,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      0.5,
			},
			outSize: testSize{outSize.width / 2, outSize.height / 2},
		},
		{
			opts: watermarkTestCase{
				position: GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      2,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				xOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      0.5,
			},
			outSize: testSize{outSize.width / 2, outSize.height / 2},
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      2,
			},
			outSize: outSize,
		},

		// Opacity
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				opacity:  0,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				opacity:  0.5,
			},
			outSize: outSize,
		},

		// Scale
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				opacity:  1,
				scale:    0,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: GravityReplicate,
				opacity:  1,
				scale:    0.5,
			},
			outSize: outSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc)
		})
	}
}

func (s *WatermarkTestSuite) TestWatermarkAnimation() {
	o := options.New()

	testCases := []testCase[watermarkTestCase]{
		{
			opts: watermarkTestCase{
				position: GravityNorthWest,
				xOffset:  10,
				yOffset:  20,
				dpr:      0.5,
				opacity:  0.5,
			},
			outSize: testSize{246, 115},
		},
		{
			opts: watermarkTestCase{
				position: GravityNorthWest,
				xOffset:  10,
				yOffset:  20,
				dpr:      2,
				opacity:  0.5,
			},
			outSize: testSize{492, 229},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.imgAnim, o, tc)
		})
	}
}

func TestWatermark(t *testing.T) {
	suite.Run(t, new(WatermarkTestSuite))
}
