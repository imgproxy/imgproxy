package processing

import (
	"bytes"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type FlattenTestSuite struct {
	testSuite

	img imagedata.ImageData
}

type flattenTestCase struct {
	background *color.RGB
	format     imagetype.Type
}

func (r flattenTestCase) Set(o *options.Options) {
	if r.background != nil {
		o.Set(keys.Background, *r.background)
	} else {
		o.Delete(keys.Background)
	}

	o.Set(keys.Format, r.format)
}

func (r flattenTestCase) String() string {
	var b bytes.Buffer
	b.WriteString(r.format.String())
	b.WriteString("_")

	if r.background != nil {
		b.WriteString(r.background.String())
	} else {
		b.WriteString("none")
	}

	return b.String()
}

func (s *FlattenTestSuite) SetupSuite() {
	s.testSuite.SetupSuite()

	var err error

	s.img, err = s.ImageDataFactory().NewFromPath(
		s.TestData.Path("test-images/bmp/32-bpp-with-alpha.bmp"),
	)
	s.Require().NoError(err)
}

func (s *FlattenTestSuite) TestBackground() {
	o := options.New()

	outSize := testSize{1080, 902}

	testCases := []testCase[flattenTestCase]{
		{opts: flattenTestCase{background: &color.RGB{R: 255, G: 0, B: 0}, format: imagetype.JPEG}, outSize: outSize},
		{opts: flattenTestCase{background: &color.RGB{R: 255, G: 0, B: 0}, format: imagetype.PNG}, outSize: outSize},
		{opts: flattenTestCase{background: nil, format: imagetype.JPEG}, outSize: outSize},
		{opts: flattenTestCase{background: nil, format: imagetype.PNG}, outSize: outSize},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			tc.opts.Set(o)

			s.processImageAndCheck(s.img, o, tc.outSize)
		})
	}
}

func TestFlatten(t *testing.T) {
	suite.Run(t, new(FlattenTestSuite))
}
