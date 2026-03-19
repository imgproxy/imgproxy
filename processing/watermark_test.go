package processing_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type watermarkTestCase struct {
	sourceFile string
	position   processing.GravityType
	opacity    float64
	xOffset    float64
	yOffset    float64
	scale      float64
	dpr        float64
	format     imagetype.Type
}

func (w watermarkTestCase) ImagePath() string {
	return w.sourceFile
}

func (w watermarkTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	wmArgs := opts.Add("watermark")
	wmArgs.Set(0, w.opacity).Set(1, w.position)

	if w.xOffset != 0 {
		wmArgs.Set(2, w.xOffset)
	}
	if w.yOffset != 0 {
		wmArgs.Set(3, w.yOffset)
	}
	if w.scale != 0 {
		wmArgs.Set(4, w.scale)
	}

	if w.dpr != 0 {
		opts.Add("dpr").Set(0, w.dpr)
	}

	if w.format != imagetype.Unknown {
		opts.Add("format").Set(0, w.format)
	}

	return opts.String()
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

type WatermarkTestSuite struct {
	testSuite
}

func (s *WatermarkTestSuite) SetupSubTest() {
	s.Config().WatermarkImage.Path = s.TestData.Path("geometry.png")
}

func (s *WatermarkTestSuite) TestWatermark() {
	outSize := testSize{1080, 902}

	testCases := []testCase[watermarkTestCase]{
		// All positions
		{
			opts:    watermarkTestCase{position: processing.GravityCenter, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityNorth, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravitySouth, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityNorthWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityNorthEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravitySouthWest, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravitySouthEast, opacity: 1},
			outSize: outSize,
		},
		{
			opts:    watermarkTestCase{position: processing.GravityReplicate, opacity: 1},
			outSize: outSize,
		},

		// Offset
		{
			opts: watermarkTestCase{
				position: processing.GravityNorth,
				xOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityNorth,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      0.5,
			},
			outSize: testSize{outSize.width / 2, outSize.height / 2},
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityNorth,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      2,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				xOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				xOffset:  50,
				yOffset:  50,
				opacity:  1,
				dpr:      0.5,
			},
			outSize: testSize{outSize.width / 2, outSize.height / 2},
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
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
				position: processing.GravityReplicate,
				opacity:  0,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				opacity:  0.5,
			},
			outSize: outSize,
		},

		// Scale
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				opacity:  1,
				scale:    0,
			},
			outSize: outSize,
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityReplicate,
				opacity:  1,
				scale:    0.5,
			},
			outSize: outSize,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.sourceFile = "test-images/bmp/24-bpp.bmp"
			tc.opts.format = imagetype.PNG

			s.processImageAndCheck(tc)
		})
	}
}

func (s *WatermarkTestSuite) TestWatermarkAnimation() {
	testCases := []testCase[watermarkTestCase]{
		{
			opts: watermarkTestCase{
				position: processing.GravityNorthWest,
				xOffset:  10,
				yOffset:  20,
				dpr:      0.5,
				opacity:  0.5,
			},
			outSize: testSize{246, 115},
		},
		{
			opts: watermarkTestCase{
				position: processing.GravityNorthWest,
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
			tc.opts.sourceFile = "test-images/gif/gif.gif"

			s.processImageAndCheck(tc)
		})
	}
}

func TestWatermark(t *testing.T) {
	suite.Run(t, new(WatermarkTestSuite))
}
