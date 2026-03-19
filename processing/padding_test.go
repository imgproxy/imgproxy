package processing_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PaddingTestSuite struct {
	testSuite
}

type paddingTestCase struct {
	dpr           float64
	paddingTop    int
	paddingRight  int
	paddingBottom int
	paddingLeft   int
}

func (r paddingTestCase) ImagePath() string {
	return "geometry.png"
}

func (r paddingTestCase) URLOptions() string {
	return fmt.Sprintf(
		"dpr:%f/padding:%d:%d:%d:%d/background:f00/enlarge:1", // enlarge:1 for dpr > 1
		r.dpr,
		r.paddingTop,
		r.paddingRight,
		r.paddingBottom,
		r.paddingLeft,
	)
}

func (r paddingTestCase) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "dpr_%g", r.dpr)
	fmt.Fprintf(&b, "_pad_%d_%d_%d_%d", r.paddingTop, r.paddingRight, r.paddingBottom, r.paddingLeft)

	return b.String()
}

func (s *PaddingTestSuite) TestPadding() {
	testCases := []testCase[paddingTestCase]{
		{
			opts: paddingTestCase{
				dpr:           0.5,
				paddingLeft:   10,
				paddingTop:    20,
				paddingRight:  30,
				paddingBottom: 40,
			},
			outSize: testSize{120, 80},
		},
		{
			opts: paddingTestCase{
				dpr:           1,
				paddingLeft:   10,
				paddingTop:    20,
				paddingRight:  30,
				paddingBottom: 40,
			},
			outSize: testSize{240, 160},
		},
		{
			opts: paddingTestCase{
				dpr:           2,
				paddingLeft:   10,
				paddingTop:    20,
				paddingRight:  30,
				paddingBottom: 40,
			},
			outSize: testSize{480, 320},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.processImageAndCheck(tc)
		})
	}
}

func TestPadding(t *testing.T) {
	suite.Run(t, new(PaddingTestSuite))
}
