package processing

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/stretchr/testify/suite"
)

type ApplyFiltersTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type effectTestCase struct {
	blur     float64
	sharpen  float64
	pixelate int
}

func (r effectTestCase) Set(o *options.Options) {
	o.Set(keys.Blur, r.blur)
	o.Set(keys.Sharpen, r.sharpen)
	o.Set(keys.Pixelate, r.pixelate)
}

func (r effectTestCase) String() string {
	b := bytes.NewBuffer(nil)

	if r.blur > 0 {
		fmt.Fprintf(b, "_blur_%f", r.blur)
	}

	if r.sharpen > 0 {
		fmt.Fprintf(b, "_sharpen_%f", r.sharpen)
	}

	if r.pixelate > 0 {
		fmt.Fprintf(b, "_pixelate_%d", r.pixelate)
	}

	name, _ := strings.CutPrefix(b.String(), "_")
	return name
}

func (s *ApplyFiltersTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("test-images/png/png.png"),
	)

	s.Require().NoError(err)
}

func (s *ApplyFiltersTestSuite) TestEffects() {
	o := options.New()

	outSize := testSize{400, 400}

	testCases := []testCase[effectTestCase]{
		{opts: effectTestCase{10, 0, 0}, outSize: outSize},
		{opts: effectTestCase{0, 10, 0}, outSize: outSize},
		{opts: effectTestCase{0, 0, 10}, outSize: outSize},
		{opts: effectTestCase{10, 10, 10}, outSize: outSize},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func TestApplyFilters(t *testing.T) {
	suite.Run(t, new(ApplyFiltersTestSuite))
}
