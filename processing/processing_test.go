package processing_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/testutil/servertest"
	"github.com/stretchr/testify/suite"
)

type ProcessingTestSuite struct {
	testSuite
}

type sizeTestCase struct {
	limit         int
	width         int
	height        int
	resizingType  processing.ResizeType
	enlarge       bool
	extend        bool
	extendAR      bool
	paddingTop    int
	paddingRight  int
	paddingBottom int
	paddingLeft   int
	rotate        int
}

func (r sizeTestCase) ImagePath() string {
	return "geometry.png"
}

func (r sizeTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	if r.limit > 0 {
		opts.Add("max_result_dimension").Set(0, r.limit)
	}

	opts.Add("resize").
		Set(0, r.resizingType).
		Set(1, r.width).
		Set(2, r.height)

	if r.enlarge {
		opts.Add("enlarge").Set(0, 1)
	}

	if r.extend {
		opts.Add("extend").Set(0, 1)
	}

	if r.extendAR {
		opts.Add("extend_aspect_ratio").Set(0, 1)
	}

	if r.rotate != 0 {
		opts.Add("rotate").Set(0, r.rotate)
	}

	if r.paddingTop > 0 || r.paddingRight > 0 || r.paddingBottom > 0 || r.paddingLeft > 0 {
		opts.Add("padding").
			Set(0, r.paddingTop).
			Set(1, r.paddingRight).
			Set(2, r.paddingBottom).
			Set(3, r.paddingLeft)
	}

	return opts.String()
}

func (r sizeTestCase) String() string {
	b := bytes.NewBuffer(nil)

	fmt.Fprintf(b, "%s:%dx%d", r.resizingType, r.width, r.height)

	if r.limit > 0 {
		fmt.Fprintf(b, ":limit:%d", r.limit)
	}

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

func (r sizeTestCase) ShortName() string {
	return fmt.Sprintf("%dx%d", r.width, r.height)
}

func (s *ProcessingTestSuite) TestResizeToFit() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{40, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 10}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 100, height: 300}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFit

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitEnlarge() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{40, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 10}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 150}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{250, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 125}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{400, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFit
			tc.opts.enlarge = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtend() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 300}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{300, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFit
			tc.opts.extend = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtendAR() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{200, 200}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{240, 100}},
		{opts: sizeTestCase{width: 250, height: 500}, outSize: testSize{200, 400}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFit
			tc.opts.extendAR = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFill() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 50}, outSize: testSize{200, 50}},
		{opts: sizeTestCase{width: 100, height: 300}, outSize: testSize{100, 100}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFill

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillEnlarge() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 300}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{300, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{400, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFill
			tc.opts.enlarge = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtend() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 300}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{300, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 300}},
		{opts: sizeTestCase{width: 300, height: 50}, outSize: testSize{300, 50}},
		{opts: sizeTestCase{width: 100, height: 300}, outSize: testSize{100, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFill
			tc.opts.extend = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtendAR() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{200, 200}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{240, 100}},
		{opts: sizeTestCase{width: 250, height: 500}, outSize: testSize{200, 400}},
		{opts: sizeTestCase{width: 300, height: 50}, outSize: testSize{300, 50}},
		{opts: sizeTestCase{width: 100, height: 300}, outSize: testSize{100, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFill
			tc.opts.extendAR = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDown() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{100, 100}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{200, 83}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{83, 100}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFillDown

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownEnlarge() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 300}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{300, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{400, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 150}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFillDown
			tc.opts.enlarge = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtend() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{300, 300}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{300, 125}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{250, 300}},
		{opts: sizeTestCase{width: 300, height: 50}, outSize: testSize{300, 50}},
		{opts: sizeTestCase{width: 100, height: 300}, outSize: testSize{100, 300}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 200}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{300, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFillDown
			tc.opts.extend = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtendAR() {
	testCases := []testCase[sizeTestCase]{
		{opts: sizeTestCase{width: 50, height: 50}, outSize: testSize{50, 50}},
		{opts: sizeTestCase{width: 50, height: 20}, outSize: testSize{50, 20}},
		{opts: sizeTestCase{width: 20, height: 50}, outSize: testSize{20, 50}},
		{opts: sizeTestCase{width: 300, height: 300}, outSize: testSize{100, 100}},
		{opts: sizeTestCase{width: 300, height: 125}, outSize: testSize{200, 83}},
		{opts: sizeTestCase{width: 250, height: 300}, outSize: testSize{83, 100}},
		{opts: sizeTestCase{width: 0, height: 50}, outSize: testSize{100, 50}},
		{opts: sizeTestCase{width: 50, height: 0}, outSize: testSize{50, 25}},
		{opts: sizeTestCase{width: 0, height: 200}, outSize: testSize{200, 100}},
		{opts: sizeTestCase{width: 300, height: 0}, outSize: testSize{200, 100}},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.ShortName(), func() {
			tc.opts.resizingType = processing.ResizeFillDown
			tc.opts.extendAR = true

			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestResultSizeLimit() {
	testCases := []testCase[sizeTestCase]{
		{
			opts: sizeTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFit,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFit,
				enlarge:      true,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFit,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFit,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        100,
				height:       150,
				resizingType: processing.ResizeFit,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFit,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  processing.ResizeFit,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 129},
		},
		{
			opts: sizeTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        1000,
				height:       50,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{50, 13},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        100,
				height:       1000,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFill,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFill,
				enlarge:      true,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFill,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFill,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        100,
				height:       150,
				resizingType: processing.ResizeFill,
				rotate:       90,
			},
			outSize: testSize{67, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFill,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  processing.ResizeFill,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 144},
		},
		{
			opts: sizeTestCase{
				limit:        1000,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        100,
				height:       100,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{50, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        1000,
				height:       50,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{50, 3},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        100,
				height:       1000,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{5, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       100,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{100, 50},
		},
		{
			opts: sizeTestCase{
				limit:        50,
				width:        150,
				height:       0,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{50, 25},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFillDown,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1000,
				resizingType: processing.ResizeFillDown,
				enlarge:      true,
			},
			outSize: testSize{100, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFillDown,
				extend:       true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       2000,
				resizingType: processing.ResizeFillDown,
				extendAR:     true,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        1000,
				height:       1500,
				resizingType: processing.ResizeFillDown,
				rotate:       90,
			},
			outSize: testSize{67, 100},
		},
		{
			opts: sizeTestCase{
				limit:        100,
				width:        0,
				height:       0,
				resizingType: processing.ResizeFillDown,
				rotate:       90,
			},
			outSize: testSize{50, 100},
		},
		{
			opts: sizeTestCase{
				limit:         200,
				width:         100,
				height:        100,
				resizingType:  processing.ResizeFillDown,
				paddingTop:    100,
				paddingRight:  200,
				paddingBottom: 300,
				paddingLeft:   400,
			},
			outSize: testSize{200, 144},
		},
		{
			opts: sizeTestCase{
				limit:         200,
				width:         1000,
				height:        1000,
				resizingType:  processing.ResizeFillDown,
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
			s.processImageAndCheck(tc)
		})
	}
}

func (s *ProcessingTestSuite) TestImageResolutionTooLarge() {
	resp := s.GET("/unsafe/max_src_resolution:0.00001/plain/local:///geometry.png")
	defer resp.Body.Close()

	s.Require().Equal(
		422, resp.StatusCode,
		"Expected status code 422 for too large image resolution",
	)
}

func TestProcessing(t *testing.T) {
	suite.Run(t, new(ProcessingTestSuite))
}

func TestMain(m *testing.M) {
	servertest.TestMain(m)
}
