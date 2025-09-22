package processing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type ProcessingTestSuite struct {
	testutil.LazySuite

	imageDataFactory testutil.LazyObj[*imagedata.Factory]
	securityConfig   testutil.LazyObj[*security.Config]
	security         testutil.LazyObj[*security.Checker]
	poConfig         testutil.LazyObj[*options.Config]
	po               testutil.LazyObj[*options.Factory]
}

func (s *ProcessingTestSuite) SetupSuite() {
	s.Require().NoError(vips.Init())

	logger.Mute()

	s.imageDataFactory, _ = testutil.NewLazySuiteObj(s, func() (*imagedata.Factory, error) {
		c := fetcher.NewDefaultConfig()
		f, err := fetcher.New(&c)
		if err != nil {
			return nil, err
		}

		return imagedata.NewFactory(f), nil
	})

	s.securityConfig, _ = testutil.NewLazySuiteObj(s, func() (*security.Config, error) {
		c := security.NewDefaultConfig()

		c.DefaultOptions.MaxSrcResolution = 10 * 1024 * 1024
		c.DefaultOptions.MaxSrcFileSize = 10 * 1024 * 1024
		c.DefaultOptions.MaxAnimationFrames = 100
		c.DefaultOptions.MaxAnimationFrameResolution = 10 * 1024 * 1024

		return &c, nil
	})

	s.security, _ = testutil.NewLazySuiteObj(s, func() (*security.Checker, error) {
		return security.New(s.securityConfig())
	})

	s.poConfig, _ = testutil.NewLazySuiteObj(s, func() (*options.Config, error) {
		c := options.NewDefaultConfig()
		return &c, nil
	})

	s.po, _ = testutil.NewLazySuiteObj(s, func() (*options.Factory, error) {
		return options.NewFactory(s.poConfig(), s.security())
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

func (s *ProcessingTestSuite) TestResizeToFit() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFit

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFit
	po.Enlarge = true

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtend() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFit
	po.Extend = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFit
	po.ExtendAspectRatio = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFill() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFill

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFill
	po.Enlarge = true

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtend() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFill
	po.Extend = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFill
	po.ExtendAspectRatio = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDown() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFillDown

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFillDown
	po.Enlarge = true

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtend() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFillDown
	po.Extend = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()
	po.ResizingType = options.ResizeFillDown
	po.ExtendAspectRatio = options.ExtendOptions{
		Enabled: true,
		Gravity: options.GravityOptions{
			Type: options.GravityCenter,
		},
	}

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
			po.Width = tc.width
			po.Height = tc.height

			result, err := ProcessImage(context.Background(), imgdata, po, nil)
			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResultSizeLimit() {
	imgdata := s.openFile("test2.jpg")

	po := s.po().NewProcessingOptions()

	testCases := []struct {
		limit        int
		width        int
		height       int
		resizingType options.ResizeType
		enlarge      bool
		extend       bool
		extendAR     bool
		padding      options.PaddingOptions
		rotate       int
		outWidth     int
		outHeight    int
	}{
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: options.ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: options.ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: options.ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: options.ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: options.ResizeFit,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFit,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFit,
			enlarge:      true,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFit,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFit,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        100,
			height:       150,
			resizingType: options.ResizeFit,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: options.ResizeFit,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        200,
			width:        100,
			height:       100,
			resizingType: options.ResizeFit,
			padding: options.PaddingOptions{
				Enabled: true,
				Top:     100,
				Right:   200,
				Bottom:  300,
				Left:    400,
			},
			outWidth:  200,
			outHeight: 129,
		},
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: options.ResizeFill,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: options.ResizeFill,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        1000,
			height:       50,
			resizingType: options.ResizeFill,
			outWidth:     50,
			outHeight:    13,
		},
		{
			limit:        50,
			width:        100,
			height:       1000,
			resizingType: options.ResizeFill,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: options.ResizeFill,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: options.ResizeFill,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: options.ResizeFill,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFill,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFill,
			enlarge:      true,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFill,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFill,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        100,
			height:       150,
			resizingType: options.ResizeFill,
			rotate:       90,
			outWidth:     67,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: options.ResizeFill,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        200,
			width:        100,
			height:       100,
			resizingType: options.ResizeFill,
			padding: options.PaddingOptions{
				Enabled: true,
				Top:     100,
				Right:   200,
				Bottom:  300,
				Left:    400,
			},
			outWidth:  200,
			outHeight: 144,
		},
		{
			limit:        1000,
			width:        100,
			height:       100,
			resizingType: options.ResizeFillDown,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        50,
			width:        100,
			height:       100,
			resizingType: options.ResizeFillDown,
			outWidth:     50,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        1000,
			height:       50,
			resizingType: options.ResizeFillDown,
			outWidth:     50,
			outHeight:    3,
		},
		{
			limit:        50,
			width:        100,
			height:       1000,
			resizingType: options.ResizeFillDown,
			outWidth:     5,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        0,
			height:       0,
			resizingType: options.ResizeFillDown,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        0,
			height:       100,
			resizingType: options.ResizeFillDown,
			outWidth:     100,
			outHeight:    50,
		},
		{
			limit:        50,
			width:        150,
			height:       0,
			resizingType: options.ResizeFillDown,
			outWidth:     50,
			outHeight:    25,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFillDown,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFillDown,
			enlarge:      true,
			outWidth:     100,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFillDown,
			extend:       true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       2000,
			resizingType: options.ResizeFillDown,
			extendAR:     true,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        1000,
			height:       1500,
			resizingType: options.ResizeFillDown,
			rotate:       90,
			outWidth:     67,
			outHeight:    100,
		},
		{
			limit:        100,
			width:        0,
			height:       0,
			resizingType: options.ResizeFillDown,
			rotate:       90,
			outWidth:     50,
			outHeight:    100,
		},
		{
			limit:        200,
			width:        100,
			height:       100,
			resizingType: options.ResizeFillDown,
			padding: options.PaddingOptions{
				Enabled: true,
				Top:     100,
				Right:   200,
				Bottom:  300,
				Left:    400,
			},
			outWidth:  200,
			outHeight: 144,
		},
		{
			limit:        200,
			width:        1000,
			height:       1000,
			resizingType: options.ResizeFillDown,
			padding: options.PaddingOptions{
				Enabled: true,
				Top:     100,
				Right:   200,
				Bottom:  300,
				Left:    400,
			},
			outWidth:  200,
			outHeight: 144,
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
		if tc.padding.Enabled {
			name += fmt.Sprintf("_padding_%dx%dx%dx%d", tc.padding.Top, tc.padding.Right, tc.padding.Bottom, tc.padding.Left)
		}

		s.Run(name, func() {
			po.SecurityOptions.MaxResultDimension = tc.limit
			po.Width = tc.width
			po.Height = tc.height
			po.ResizingType = tc.resizingType
			po.Enlarge = tc.enlarge
			po.Extend.Enabled = tc.extend
			po.ExtendAspectRatio.Enabled = tc.extendAR
			po.Rotate = tc.rotate
			po.Padding = tc.padding

			result, err := ProcessImage(context.Background(), imgdata, po, nil)

			s.Require().NoError(err)
			s.Require().NotNil(result)

			s.checkSize(result, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestImageResolutionTooLarge() {
	po := s.po().NewProcessingOptions()
	po.SecurityOptions.MaxSrcResolution = 1

	imgdata := s.openFile("test2.jpg")
	_, err := ProcessImage(context.Background(), imgdata, po, nil)

	s.Require().Error(err)
	s.Require().Equal(422, ierrors.Wrap(err, 0).StatusCode())
}

func TestProcessing(t *testing.T) {
	suite.Run(t, new(ProcessingTestSuite))
}
