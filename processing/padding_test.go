package processing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type PaddingTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type paddingTestCase struct {
	dpr           float64
	paddingLeft   int
	paddingTop    int
	paddingRight  int
	paddingBottom int
}

func (r paddingTestCase) Set(o *options.Options) {
	o.Set(keys.Dpr, r.dpr)
	o.Set(keys.PaddingLeft, r.paddingLeft)
	o.Set(keys.PaddingTop, r.paddingTop)
	o.Set(keys.PaddingRight, r.paddingRight)
	o.Set(keys.PaddingBottom, r.paddingBottom)
	o.Set(keys.Background, color.RGB{R: 255, G: 0, B: 0})
	o.Set(keys.Enlarge, true) // for drp > 1
}

func (r paddingTestCase) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "dpr_%g", r.dpr)
	fmt.Fprintf(&b, "_pad_%d_%d_%d_%d", r.paddingLeft, r.paddingTop, r.paddingRight, r.paddingBottom)

	return b.String()
}

func (s *PaddingTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("geometry.png"),
	)
	s.Require().NoError(err)
}

func (s *PaddingTestSuite) TestPadding() {
	o := options.New()

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
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func TestPadding(t *testing.T) {
	suite.Run(t, new(PaddingTestSuite))
}
