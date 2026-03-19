package processing_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ApplyFiltersTestSuite struct {
	testSuite
}

type effectTestCase struct {
	blur     float64
	sharpen  float64
	pixelate int
}

func (r effectTestCase) ImagePath() string {
	return "test-images/png/png.png"
}

func (r effectTestCase) URLOptions() string {
	return fmt.Sprintf("blur:%f/sharpen:%f/pixelate:%d", r.blur, r.sharpen, r.pixelate)
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

func (s *ApplyFiltersTestSuite) TestEffects() {
	outSize := testSize{400, 400}

	testCases := []testCase[effectTestCase]{
		{opts: effectTestCase{10, 0, 0}, outSize: outSize},
		{opts: effectTestCase{0, 10, 0}, outSize: outSize},
		{opts: effectTestCase{0, 0, 10}, outSize: outSize},
		{opts: effectTestCase{10, 10, 10}, outSize: outSize},
	}

	for _, tc := range testCases {
		s.Run(tc.opts.String(), func() {
			s.processImageAndCheck(tc)
		})
	}
}

func TestApplyFilters(t *testing.T) {
	suite.Run(t, new(ApplyFiltersTestSuite))
}
