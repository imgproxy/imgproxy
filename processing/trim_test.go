package processing_test

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips/color"
	"github.com/stretchr/testify/suite"
)

type TrimTestSuite struct {
	testSuite
}

type trimTestCase struct {
	sourceFile string
	threshold  int
	color      *color.RGB
	equalHor   bool
	equalVer   bool
}

func (r trimTestCase) ImagePath() string {
	return r.sourceFile
}

func (r trimTestCase) URLOptions() string {
	opts := testutil.NewOptionsBuilder()

	args := opts.Add("trim")
	args.Set(0, r.threshold)

	if r.color != nil {
		args.Set(1, fmt.Sprintf("%02x%02x%02x", r.color.R, r.color.G, r.color.B))
	}

	if r.equalHor {
		args.Set(2, 1)
	}

	if r.equalVer {
		args.Set(3, 1)
	}

	return opts.String()
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

func (s *TrimTestSuite) TestThreshold() {
	testCases := []testCase[trimTestCase]{
		{opts: trimTestCase{threshold: 5}, outSize: testSize{320, 220}},
		{opts: trimTestCase{threshold: 100}, outSize: testSize{290, 190}},
		{opts: trimTestCase{threshold: 150}, outSize: testSize{260, 160}},
		{opts: trimTestCase{threshold: 200}, outSize: testSize{230, 130}},
		{opts: trimTestCase{threshold: 220}, outSize: testSize{200, 100}},
		{
			opts:    trimTestCase{color: &color.Red},
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
			tc.opts.sourceFile = "trim1.png"
			s.processImageAndCheck(tc)
		})
	}
}

func (s *TrimTestSuite) TestColor() {
	testCases := []testCase[trimTestCase]{
		{
			opts:    trimTestCase{threshold: 1},
			outSize: testSize{200, 170},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.Red,
			},
			outSize: testSize{200, 170},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.Blue,
			},
			outSize: testSize{200, 130},
		},
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.White,
			},
			outSize: testSize{200, 200},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.sourceFile = "trim2.png"
			s.processImageAndCheck(tc)
		})
	}
}

func (s *TrimTestSuite) TestAlpha() {
	testCases := []testCase[trimTestCase]{
		{
			opts: trimTestCase{
				threshold: 1,
				color:     &color.Magenta,
			},
			outSize: testSize{200, 130},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.sourceFile = "trim3.png"
			s.processImageAndCheck(tc)
		})
	}
}

func TestTrim(t *testing.T) {
	suite.Run(t, new(TrimTestSuite))
}
