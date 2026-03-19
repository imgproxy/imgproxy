package processing_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/testutil/servertest"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type testSuite struct {
	servertest.Suite

	ImageMatcher testutil.LazyObj[*testutil.ImageHashCacheMatcher]
}

type optsFactory interface {
	URLOptions() string
	ImagePath() string
}

type testCaseParams interface {
	Options() optsFactory
	OutSize() testSize
	OutInterpretation() vips.Interpretation
}

type testCase[T optsFactory] struct {
	opts              T
	outSize           testSize
	outInterpretation vips.Interpretation
}

func (c testCase[T]) Options() optsFactory {
	return c.opts
}

func (c testCase[T]) URLOptions() string {
	return c.opts.URLOptions()
}

func (c testCase[T]) ImagePath() string {
	return c.opts.ImagePath()
}

func (c testCase[T]) OutSize() testSize {
	return c.outSize
}

func (c testCase[T]) OutInterpretation() vips.Interpretation {
	return c.outInterpretation
}

type testSize struct {
	width  int
	height int
}

func (c testSize) String() string {
	return fmt.Sprintf("%dx%d", c.width, c.height)
}

func (s *testSuite) SetupSuite() {
	s.Suite.SetupSuite()

	s.ImageMatcher, _ = testutil.NewLazySuiteObj(s, func() (*testutil.ImageHashCacheMatcher, error) {
		return testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeSHA256), nil
	})
}

func (s *testSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *testSuite) processImage(opts optsFactory) imagedata.ImageData {
	s.T().Helper()

	reqPath := "/unsafe/" + opts.URLOptions() + "/plain/local:///" + opts.ImagePath()
	resp := s.GET(reqPath)

	resultBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(
		http.StatusOK,
		resp.StatusCode,
		"Expected status code 200, got %d; Path: %s; Body: %s",
		resp.StatusCode,
		reqPath,
		string(resultBytes),
	)

	resultData, err := s.Imgproxy().ImageDataFactory().NewFromBytes(resultBytes)
	s.Require().NoError(err)

	return resultData
}

func (s *testSuite) processImageAndCheck(tc testCaseParams) {
	s.T().Helper()

	resultData := s.processImage(tc.Options())
	defer resultData.Close()

	// Load the result image to check its size and interpretation
	resultImg := new(vips.Image)
	defer resultImg.Clear()

	err := resultImg.Load(resultData, 1.0, 0, 1)
	s.Require().NoError(err)

	outSize := tc.OutSize()
	outInterpretation := tc.OutInterpretation()

	s.Require().Equal(resultImg.Width(), outSize.width, "Width mismatch")
	s.Require().Equal(resultImg.Height(), outSize.height, "Height mismatch")

	if outInterpretation != vips.InterpretationMultiBand {
		// Check the interpretation
		actualInterpretation := resultImg.Type()
		s.Require().Equal(outInterpretation, actualInterpretation)
	}

	s.ImageMatcher().ImageMatches(s.T(), resultData.Reader(), "test", 0.0005)
}
