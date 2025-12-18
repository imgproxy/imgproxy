package processing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/stretchr/testify/suite"
)

type ProcessingTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type sizeLimitTestCase struct {
	limit         int
	width         int
	height        int
	resizingType  ResizeType
	enlarge       bool
	extend        bool
	extendAR      bool
	paddingTop    int
	paddingRight  int
	paddingBottom int
	paddingLeft   int
	rotate        int
}

func (r sizeLimitTestCase) Set(o *options.Options) {
	o.Set(keys.MaxResultDimension, r.limit)
	o.Set(keys.Width, r.width)
	o.Set(keys.Height, r.height)
	o.Set(keys.ResizingType, r.resizingType)
	o.Set(keys.Enlarge, r.enlarge)
	o.Set(keys.ExtendEnabled, r.extend)
	o.Set(keys.ExtendAspectRatioEnabled, r.extendAR)
	o.Set(keys.Rotate, r.rotate)
	o.Set(keys.PaddingTop, r.paddingTop)
	o.Set(keys.PaddingRight, r.paddingRight)
	o.Set(keys.PaddingBottom, r.paddingBottom)
	o.Set(keys.PaddingLeft, r.paddingLeft)
}

func (r sizeLimitTestCase) String() string {
	b := bytes.NewBuffer(nil)

	fmt.Fprintf(b, "%s:%dx%d:limit:%d", r.resizingType, r.width, r.height, r.limit)

	if r.enlarge {
		fmt.Fprintf(b, "_en:%t", r.enlarge)
	}

	if r.extend {
		fmt.Fprintf(b, "_ex:%t", r.extend)
	}

	if r.extendAR {
		fmt.Fprintf(b, "_exAR:%t", r.extendAR)
	}

	if r.rotate != 0 {
		fmt.Fprintf(b, "_rotate:%d", r.rotate)
	}

	if r.paddingTop > 0 || r.paddingRight > 0 || r.paddingBottom > 0 || r.paddingLeft > 0 {
		fmt.Fprintf(
			b, "_padding:%dx%dx%dx%d",
			r.paddingTop, r.paddingRight, r.paddingBottom, r.paddingLeft,
		)
	}

	return b.String()
}

func (s *ProcessingTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(s.TestData.Path("geometry.png"))
	s.Require().NoError(err)
}

func (s *ProcessingTestSuite) TestResizeToFit() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 25}},
		{opts: testSize{50, 20}, outSize: testSize{40, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 10}},
		{opts: testSize{300, 300}, outSize: testSize{200, 100}},
		{opts: testSize{300, 50}, outSize: testSize{100, 50}},
		{opts: testSize{100, 300}, outSize: testSize{100, 50}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitEnlarge() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.Enlarge, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 25}},
		{opts: testSize{50, 20}, outSize: testSize{40, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 10}},
		{opts: testSize{300, 300}, outSize: testSize{300, 150}},
		{opts: testSize{300, 125}, outSize: testSize{250, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 125}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{400, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtend() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.ExtendEnabled, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{300, 300}},
		{opts: testSize{300, 125}, outSize: testSize{300, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtendAR() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.ExtendAspectRatioEnabled, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{200, 200}},
		{opts: testSize{300, 125}, outSize: testSize{240, 100}},
		{opts: testSize{250, 500}, outSize: testSize{200, 400}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFill() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{200, 100}},
		{opts: testSize{300, 50}, outSize: testSize{200, 50}},
		{opts: testSize{100, 300}, outSize: testSize{100, 100}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillEnlarge() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.Enlarge, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{300, 300}},
		{opts: testSize{300, 125}, outSize: testSize{300, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{400, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtend() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.ExtendEnabled, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{300, 300}},
		{opts: testSize{300, 125}, outSize: testSize{300, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 300}},
		{opts: testSize{300, 50}, outSize: testSize{300, 50}},
		{opts: testSize{100, 300}, outSize: testSize{100, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtendAR() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.ExtendAspectRatioEnabled, true)
	o.Set(keys.ExtendAspectRatioGravityType, GravityCenter)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{200, 200}},
		{opts: testSize{300, 125}, outSize: testSize{240, 100}},
		{opts: testSize{250, 500}, outSize: testSize{200, 400}},
		{opts: testSize{300, 50}, outSize: testSize{300, 50}},
		{opts: testSize{100, 300}, outSize: testSize{100, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDown() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{100, 100}},
		{opts: testSize{300, 125}, outSize: testSize{200, 83}},
		{opts: testSize{250, 300}, outSize: testSize{83, 100}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownEnlarge() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.Enlarge, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{300, 300}},
		{opts: testSize{300, 125}, outSize: testSize{300, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{400, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtend() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.ExtendEnabled, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{300, 300}},
		{opts: testSize{300, 125}, outSize: testSize{300, 125}},
		{opts: testSize{250, 300}, outSize: testSize{250, 300}},
		{opts: testSize{300, 50}, outSize: testSize{300, 50}},
		{opts: testSize{100, 300}, outSize: testSize{100, 300}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 200}},
		{opts: testSize{300, 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtendAR() {
	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.ExtendAspectRatioEnabled, true)

	testCases := []testCase[testSize]{
		{opts: testSize{50, 50}, outSize: testSize{50, 50}},
		{opts: testSize{50, 20}, outSize: testSize{50, 20}},
		{opts: testSize{20, 50}, outSize: testSize{20, 50}},
		{opts: testSize{300, 300}, outSize: testSize{100, 100}},
		{opts: testSize{300, 125}, outSize: testSize{200, 83}},
		{opts: testSize{250, 300}, outSize: testSize{83, 100}},
		{opts: testSize{0, 50}, outSize: testSize{100, 50}},
		{opts: testSize{50, 0}, outSize: testSize{50, 25}},
		{opts: testSize{0, 200}, outSize: testSize{200, 100}},
		{opts: testSize{300, 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestResultSizeLimit() {
	testCases := []testCase[sizeLimitTestCase]{
		{
			opts: sizeLimitTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFit,
				enlarge:      true,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFit,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFit,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        100,
				height:       150,
				resizingType: ResizeFit,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: ResizeFit,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  ResizeFit,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 129},
		},
		{
			opts: sizeLimitTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: ResizeFill,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: ResizeFill,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        1000,
				height:       50,
				resizingType: ResizeFill,
			},
			outSize: testSize{50, 13},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        100,
				height:       1000,
				resizingType: ResizeFill,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: ResizeFill,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: ResizeFill,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: ResizeFill,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFill,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFill,
				enlarge:      true,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFill,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFill,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        100,
				height:       150,
				resizingType: ResizeFill,
				rotate:       90,
			},
			outSize: testSize{67, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: ResizeFill,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  ResizeFill,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 144},
		},
		{
			opts: sizeLimitTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        1000,
				height:       50,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{50, 3},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        100,
				height:       1000,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{5, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeLimitTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFillDown,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: ResizeFillDown,
				enlarge:      true,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFillDown,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: ResizeFillDown,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        1000,
				height:       1500,
				resizingType: ResizeFillDown,
				rotate:       90,
			},
			outSize: testSize{67, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: ResizeFillDown,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeLimitTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  ResizeFillDown,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 144},
		},
		{
			opts: sizeLimitTestCase{
				limit:         200,
				width:         1000,
				height:        1000,
				resizingType:  ResizeFillDown,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 144},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			o := options.New()
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func (s *ProcessingTestSuite) TestImageResolutionTooLarge() {
	o := options.New()
	o.Set(keys.MaxSrcResolution, 1)

	_, err := s.Processor().ProcessImage(s.T().Context(), s.img, o)

	s.Require().Error(err)
	s.Require().Equal(422, errctx.Wrap(err).StatusCode())
}

func TestProcessing(t *testing.T) {
	suite.Run(t, new(ProcessingTestSuite))
}
