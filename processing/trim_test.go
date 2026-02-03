package processing

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type TrimTestSuite struct {
	testSuite

	imgThreshold imagedata.ImageData
	imgColor     imagedata.ImageData
	imgAlpha     imagedata.ImageData
}

type trimTestCase struct {
	threshold int
	color     *color.RGB
	equalHor  bool
	equalVer  bool
}

func (r trimTestCase) Set(o *options.Options) {
	if r.threshold > 0 {
		o.Set(keys.TrimThreshold, r.threshold)
	}

	if r.color != nil {
		o.Set(keys.TrimColor, *r.color)
	} else {
		o.Delete(keys.TrimColor)
	}

	o.Set(keys.TrimEqualHor, r.equalHor)
	o.Set(keys.TrimEqualVer, r.equalVer)
}

func (r trimTestCase) String() string {
	var b bytes.Buffer

	b.WriteString("_trim_")
	b.WriteString("threshold_")
	b.WriteString(strconv.Itoa(r.threshold))

	if r.color != nil {
		b.WriteString("_color_")
		b.WriteString(r.color.String())
	}

	if r.equalHor {
		b.WriteString("_equalHor")
	}

	if r.equalVer {
		b.WriteString("_equalVer")
	}

	if b.String() == "" {
		b.WriteString("default")
	}

	n, _ := strings.CutPrefix(b.String(), "_")
	return n
}

func (s *TrimTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.imgThreshold, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("trim1.png"),
	)
	s.Require().NoError(err)

	s.imgColor, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("trim2.png"),
	)
	s.Require().NoError(err)

	s.imgAlpha, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("trim3.png"),
	)
	s.Require().NoError(err)
}

func (s *TrimTestSuite) TestThreshold() {
	o := options.New()

	testCases := []testCase[trimTestCase]{
		{opts: trimTestCase{threshold: 5}, outSize: testSize{320, 220}},
		{opts: trimTestCase{threshold: 100}, outSize: testSize{290, 190}},
		{opts: trimTestCase{threshold: 150}, outSize: testSize{260, 160}},
		{opts: trimTestCase{threshold: 200}, outSize: testSize{230, 130}},
		{opts: trimTestCase{threshold: 220}, outSize: testSize{200, 100}},
		{
			opts:    trimTestCase{color: &color.RGB{R: 255, G: 0, B: 0}},
			outSize: testSize{350, 250},
		},
		{
			opts:    trimTestCase{threshold: 5, equalHor: true},
			outSize: testSize{330, 220},
		},
		{
			opts:    trimTestCase{threshold: 5, equalVer: true},
			outSize: testSize{320, 230},
		},
		{
			opts:    trimTestCase{threshold: 5, equalHor: true, equalVer: true},
			outSize: testSize{330, 230},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.imgThreshold, o, tc)
		})
	}
}

func (s *TrimTestSuite) TestColor() {
	o := options.New()

	testCases := []testCase[trimTestCase]{
		{
			opts:    trimTestCase{threshold: 1},
			outSize: testSize{200, 170},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.RGB{R: 255, G: 0, B: 0},
			},
			outSize: testSize{200, 170},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.RGB{R: 0, G: 0, B: 255},
			},
			outSize: testSize{200, 130},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.RGB{R: 255, G: 255, B: 255},
			},
			outSize: testSize{200, 200},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.imgColor, o, tc)
		})
	}
}

func (s *TrimTestSuite) TestAlpha() {
	o := options.New()

	testCases := []testCase[trimTestCase]{
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.RGB{R: 255, G: 0, B: 255},
			},
			outSize: testSize{200, 130},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.imgAlpha, o, tc)
		})
	}
}

func TestTrim(t *testing.T) {
	suite.Run(t, new(TrimTestSuite))
}
