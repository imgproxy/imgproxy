package processing

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type ProcessingTestSuite struct {
	testutil.LazySuite

	imageDataFactory testutil.LazyObj[*imagedata.Factory]
	securityConfig   testutil.LazyObj[*security.Config]
	security         testutil.LazyObj[*security.Checker]
	config           testutil.LazyObj[*Config]
	processor        testutil.LazyObj[*Processor]
}

func (s *ProcessingTestSuite) SetupSuite() {
	vipsCfg := vips.NewDefaultConfig()
	s.Require().NoError(vips.Init(&vipsCfg))

	logger.Mute()

	s.imageDataFactory, _ = testutil.NewLazySuiteObj(s, func() (*imagedata.Factory, error) {
		c := fetcher.NewDefaultConfig()
		f, err := fetcher.New(&c)
		if err != nil {
			return nil, err
		}

		return imagedata.NewFactory(f, nil), nil
	})

	s.securityConfig, _ = testutil.NewLazySuiteObj(s, func() (*security.Config, error) {
		c := security.NewDefaultConfig()

		c.MaxSrcResolution = 10 * 1024 * 1024
		c.MaxSrcFileSize = 10 * 1024 * 1024
		c.MaxAnimationFrames = 100
		c.MaxAnimationFrameResolution = 10 * 1024 * 1024

		return &c, nil
	})

	s.security, _ = testutil.NewLazySuiteObj(s, func() (*security.Checker, error) {
		return security.New(s.securityConfig())
	})

	s.config, _ = testutil.NewLazySuiteObj(s, func() (*Config, error) {
		c := NewDefaultConfig()
		return &c, nil
	})

	s.processor, _ = testutil.NewLazySuiteObj(s, func() (*Processor, error) {
		return New(s.config(), s.security(), nil)
	})
}

func (s *ProcessingTestSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *ProcessingTestSuite) openFile(name string) imagedata.ImageData {
	wd, err := os.Getwd()
	s.Require().NoError(err)
	path := filepath.Join(wd, "..", "testdata", name)

	imagedata, err := s.imageDataFactory().NewFromPath(path)
	s.Require().NoError(err)

	return imagedata
}

func (s *ProcessingTestSuite) checkSize(r *Result, width, height int) {
	s.Require().NotNil(r)
	s.Require().Equal(width, r.ResultWidth, "Width mismatch")
	s.Require().Equal(height, r.ResultHeight, "Height mismatch")
}

func (s *ProcessingTestSuite) processImageAndCheck(
	imgdata imagedata.ImageData,
	o *options.Options,
	expectedWidth, expectedHeight int,
) {
	result, err := s.processor().ProcessImage(s.T().Context(), imgdata, o)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.checkSize(result, expectedWidth, expectedHeight)
}

func (s *ProcessingTestSuite) TestResizeToFit() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 25},
		{width: 50, height: 20, outWidth: 40, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 10},
		{width: 300, height: 300, outWidth: 200, outHeight: 100},
		{width: 300, height: 50, outWidth: 100, outHeight: 50},
		{width: 100, height: 300, outWidth: 100, outHeight: 50},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitEnlarge() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.Enlarge, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 25},
		{width: 50, height: 20, outWidth: 40, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 10},
		{width: 300, height: 300, outWidth: 300, outHeight: 150},
		{width: 300, height: 125, outWidth: 250, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 125},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 400, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 150},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtend() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.ExtendEnabled, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 300, outHeight: 300},
		{width: 300, height: 125, outWidth: 300, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtendAR() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFit)
	o.Set(keys.ExtendAspectRatioEnabled, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 200, outHeight: 200},
		{width: 300, height: 125, outWidth: 240, outHeight: 100},
		{width: 250, height: 500, outWidth: 200, outHeight: 400},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFill() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 200, outHeight: 100},
		{width: 300, height: 50, outWidth: 200, outHeight: 50},
		{width: 100, height: 300, outWidth: 100, outHeight: 100},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillEnlarge() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.Enlarge, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 300, outHeight: 300},
		{width: 300, height: 125, outWidth: 300, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 400, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 150},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtend() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.ExtendEnabled, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 300, outHeight: 300},
		{width: 300, height: 125, outWidth: 300, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 300},
		{width: 300, height: 50, outWidth: 300, outHeight: 50},
		{width: 100, height: 300, outWidth: 100, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtendAR() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFill)
	o.Set(keys.ExtendAspectRatioEnabled, true)
	o.Set(keys.ExtendAspectRatioGravityType, GravityCenter)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 200, outHeight: 200},
		{width: 300, height: 125, outWidth: 240, outHeight: 100},
		{width: 250, height: 500, outWidth: 200, outHeight: 400},
		{width: 300, height: 50, outWidth: 300, outHeight: 50},
		{width: 100, height: 300, outWidth: 100, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDown() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 100, outHeight: 100},
		{width: 300, height: 125, outWidth: 200, outHeight: 83},
		{width: 250, height: 300, outWidth: 83, outHeight: 100},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownEnlarge() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.Enlarge, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 300, outHeight: 300},
		{width: 300, height: 125, outWidth: 300, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 400, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 150},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtend() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.ExtendEnabled, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 300, outHeight: 300},
		{width: 300, height: 125, outWidth: 300, outHeight: 125},
		{width: 250, height: 300, outWidth: 250, outHeight: 300},
		{width: 300, height: 50, outWidth: 300, outHeight: 50},
		{width: 100, height: 300, outWidth: 100, outHeight: 300},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 200},
		{width: 300, height: 0, outWidth: 300, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtendAR() {
	imgdata := s.openFile("test2.jpg")

	o := options.New()
	o.Set(keys.ResizingType, ResizeFillDown)
	o.Set(keys.ExtendAspectRatioEnabled, true)

	testCases := []struct {
		width     int
		height    int
		outWidth  int
		outHeight int
	}{
		{width: 50, height: 50, outWidth: 50, outHeight: 50},
		{width: 50, height: 20, outWidth: 50, outHeight: 20},
		{width: 20, height: 50, outWidth: 20, outHeight: 50},
		{width: 300, height: 300, outWidth: 100, outHeight: 100},
		{width: 300, height: 125, outWidth: 200, outHeight: 83},
		{width: 250, height: 300, outWidth: 83, outHeight: 100},
		{width: 0, height: 50, outWidth: 100, outHeight: 50},
		{width: 50, height: 0, outWidth: 50, outHeight: 25},
		{width: 0, height: 200, outWidth: 200, outHeight: 100},
		{width: 300, height: 0, outWidth: 200, outHeight: 100},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("%dx%d", tc.width, tc.height), func() {
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResultSizeLimit() {
	imgdata := s.openFile("test2.jpg")

	testCases := []struct {
		limit         int
		width         int
		height        int
		resizingType  ResizeType
		enlarge       bool
		extend        bool
		extendAR      bool
		paddingTop    int
		paddingRight  int
		paddingBottom int
		paddingLeft   int
		rotate        int
		outWidth      int
		outHeight     int
	}{
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFit,
			enlarge:      true,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFit,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFit,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        100,
			height:       150,
			resizingType: ResizeFit,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: ResizeFit,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:         200,
			width:         100,
			height:        100,
			resizingType:  ResizeFit,
			paddingTop:    100,
			paddingRight:  200,
			paddingBottom: 300,
			paddingLeft:   400,
			outWidth:      200,
			outHeight:     129,
		},
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: ResizeFill,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: ResizeFill,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        1000,
			height:       50,
			resizingType: ResizeFill,
			outWidth:     50,
			outHeight:    13,
		},
		{
			limit:        50,
			width:        100,
			height:       1000,
			resizingType: ResizeFill,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: ResizeFill,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: ResizeFill,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: ResizeFill,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFill,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFill,
			enlarge:      true,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFill,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFill,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        100,
			height:       150,
			resizingType: ResizeFill,
			rotate:       90,
			outWidth:     67,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: ResizeFill,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:         200,
			width:         100,
			height:        100,
			resizingType:  ResizeFill,
			paddingTop:    100,
			paddingRight:  200,
			paddingBottom: 300,
			paddingLeft:   400,
			outWidth:      200,
			outHeight:     144,
		},
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: ResizeFillDown,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: ResizeFillDown,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        1000,
			height:       50,
			resizingType: ResizeFillDown,
			outWidth:     50,
			outHeight:    3,
		},
		{
			limit:        50,
			width:        100,
			height:       1000,
			resizingType: ResizeFillDown,
			outWidth:     5,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: ResizeFillDown,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: ResizeFillDown,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: ResizeFillDown,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFillDown,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: ResizeFillDown,
			enlarge:      true,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFillDown,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: ResizeFillDown,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       1500,
			resizingType: ResizeFillDown,
			rotate:       90,
			outWidth:     67,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: ResizeFillDown,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:         200,
			width:         100,
			height:        100,
			resizingType:  ResizeFillDown,
			paddingTop:    100,
			paddingRight:  200,
			paddingBottom: 300,
			paddingLeft:   400,
			outWidth:      200,
			outHeight:     144,
		},
		{
			limit:         200,
			width:         1000,
			height:        1000,
			resizingType:  ResizeFillDown,
			paddingTop:    100,
			paddingRight:  200,
			paddingBottom: 300,
			paddingLeft:   400,
			outWidth:      200,
			outHeight:     144,
		},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_%dx%d_limit_%d", tc.resizingType, tc.width, tc.height, tc.limit)
		if tc.enlarge {
			name += "_enlarge"
		}
		if tc.extend {
			name += "_extend"
		}
		if tc.extendAR {
			name += "_extendAR"
		}
		if tc.rotate != 0 {
			name += fmt.Sprintf("_rot_%d", tc.rotate)
		}
		if tc.paddingTop > 0 || tc.paddingRight > 0 || tc.paddingBottom > 0 || tc.paddingLeft > 0 {
			name += fmt.Sprintf(
				"_padding_%dx%dx%dx%d",
				tc.paddingTop, tc.paddingRight, tc.paddingBottom, tc.paddingLeft,
			)
		}

		s.Run(name, func() {
			o := options.New()
			o.Set(keys.MaxResultDimension, tc.limit)
			o.Set(keys.Width, tc.width)
			o.Set(keys.Height, tc.height)
			o.Set(keys.ResizingType, tc.resizingType)
			o.Set(keys.Enlarge, tc.enlarge)
			o.Set(keys.ExtendEnabled, tc.extend)
			o.Set(keys.ExtendAspectRatioEnabled, tc.extendAR)
			o.Set(keys.Rotate, tc.rotate)
			o.Set(keys.PaddingTop, tc.paddingTop)
			o.Set(keys.PaddingRight, tc.paddingRight)
			o.Set(keys.PaddingBottom, tc.paddingBottom)
			o.Set(keys.PaddingLeft, tc.paddingLeft)

			s.processImageAndCheck(imgdata, o, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestImageResolutionTooLarge() {
	o := options.New()
	o.Set(keys.MaxSrcResolution, 1)

	imgdata := s.openFile("test2.jpg")
	_, err := s.processor().ProcessImage(s.T().Context(), imgdata, o)

	s.Require().Error(err)
	s.Require().Equal(422, errctx.Wrap(err).StatusCode())
}

func TestProcessing(t *testing.T) {
	suite.Run(t, new(ProcessingTestSuite))
}
