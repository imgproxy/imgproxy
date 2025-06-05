package processing

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type ProcessingTestSuite struct {
	suite.Suite
}

func (s *ProcessingTestSuite) SetupSuite() {
	config.Reset()

	s.Require().NoError(imagedata.Init())
	s.Require().NoError(vips.Init())

	logrus.SetOutput(io.Discard)
}

func (s *ProcessingTestSuite) openFile(name string) *imagedata.ImageData {
	secopts := security.Options{
		MaxSrcResolution:            10 * 1024 * 1024,
		MaxSrcFileSize:              10 * 1024 * 1024,
		MaxAnimationFrames:          100,
		MaxAnimationFrameResolution: 10 * 1024 * 1024,
	}

	wd, err := os.Getwd()
	s.Require().NoError(err)
	path := filepath.Join(wd, "..", "testdata", name)

	imagedata, err := imagedata.FromFile(path, "test image", secopts)
	s.Require().NoError(err)

	return imagedata
}

func (s *ProcessingTestSuite) checkSize(imgdata *imagedata.ImageData, width, height int) {
	img := new(vips.Image)
	err := img.Load(imgdata, 1, 1, 1)
	s.Require().NoError(err)
	defer img.Clear()

	s.Require().Equal(width, img.Width(), "Width mismatch")
	s.Require().Equal(height, img.Height(), "Height mismatch")
}

func (s *ProcessingTestSuite) TestResizeToFit() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtend() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFitExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFill() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtend() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDown() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownEnlarge() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtend() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func (s *ProcessingTestSuite) TestResizeToFillDownExtendAR() {
	imgdata := s.openFile("test2.jpg")

	po := options.NewProcessingOptions()
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

			outImgdata, err := ProcessImage(context.Background(), imgdata, po)
			s.Require().NoError(err)
			s.Require().NotNil(outImgdata)

			s.checkSize(outImgdata, tc.outWidth, tc.outHeight)
		})
	}
}

func TestProcessing(t *testing.T) {
	suite.Run(t, new(ProcessingTestSuite))
}
