package processing

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type testSuite struct {
	testutil.LazySuite

	TestData     *testutil.TestDataProvider
	ImageMatcher *testutil.ImageHashCacheMatcher

	ImageDataFactory  testutil.LazyObj[*imagedata.Factory]
	SecurityConfig    testutil.LazyObj[*security.Config]
	Security          testutil.LazyObj[*security.Checker]
	Config            testutil.LazyObj[*Config]
	WatermarkProvider testutil.LazyObj[auximageprovider.Provider]
	Processor         testutil.LazyObj[*Processor]
}

type optsFactory interface {
	Set(o *options.Options)
	String() string
}

type testCase[T optsFactory] struct {
	opts    T
	outSize testSize
}

type testSize struct {
	width  int
	height int
}

func (c testSize) Set(o *options.Options) {
	o.Set(keys.Width, c.width)
	o.Set(keys.Height, c.height)
}

func (c testSize) String() string {
	return fmt.Sprintf("%dx%d", c.width, c.height)
}

func (s *testSuite) SetupSuite() {
	vipsCfg := vips.NewDefaultConfig()
	s.Require().NoError(vips.Init(&vipsCfg))

	logger.Mute()

	s.TestData = testutil.NewTestDataProvider(s.T)

	s.ImageDataFactory, _ = testutil.NewLazySuiteObj(s, func() (*imagedata.Factory, error) {
		c := fetcher.NewDefaultConfig()
		f, err := fetcher.New(&c)
		if err != nil {
			return nil, err
		}

		return imagedata.NewFactory(f, nil), nil
	})

	s.SecurityConfig, _ = testutil.NewLazySuiteObj(s, func() (*security.Config, error) {
		c := security.NewDefaultConfig()

		c.MaxSrcResolution = 10 * 1024 * 1024
		c.MaxSrcFileSize = 10 * 1024 * 1024
		c.MaxAnimationFrames = 100
		c.MaxAnimationFrameResolution = 10 * 1024 * 1024

		return &c, nil
	})

	s.Security, _ = testutil.NewLazySuiteObj(s, func() (*security.Checker, error) {
		return security.New(s.SecurityConfig())
	})

	s.Config, _ = testutil.NewLazySuiteObj(s, func() (*Config, error) {
		c := NewDefaultConfig()
		return &c, nil
	})

	s.WatermarkProvider, _ = testutil.NewLazySuiteObj(s, func() (auximageprovider.Provider, error) {
		return auximageprovider.NewStaticProvider(
			s.T().Context(),
			&auximageprovider.StaticConfig{
				Path: s.TestData.Path("geometry.png"),
			},
			"watermark",
			s.ImageDataFactory(),
		)
	})

	s.Processor, _ = testutil.NewLazySuiteObj(s, func() (*Processor, error) {
		return New(s.Config(), s.Security(), s.WatermarkProvider())
	})

	s.ImageMatcher = testutil.NewImageHashCacheMatcher(s.TestData, testutil.HashTypeSHA256)
}

func (s *testSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *testSuite) processImageAndCheck(
	imgdata imagedata.ImageData,
	o *options.Options,
	outSize testSize,
) {
	result, err := s.Processor().ProcessImage(s.T().Context(), imgdata, o)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Require().Equal(result.ResultWidth, outSize.width, "Width mismatch")
	s.Require().Equal(result.ResultHeight, outSize.height, "Height mismatch")

	s.ImageMatcher.ImageMatches(s.T(), result.OutData.Reader(), "test", 0)
}
