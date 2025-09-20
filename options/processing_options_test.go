package options

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

type ProcessingOptionsTestSuite struct {
	testutil.LazySuite

	config  testutil.LazyObj[*Config]
	factory testutil.LazyObj[*Factory]
}

func (s *ProcessingOptionsTestSuite) SetupSuite() {
	s.config, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Config, error) {
			c := NewDefaultConfig()
			return &c, nil
		},
	)

	s.factory, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Factory, error) {
			return NewFactory(s.config())
		},
	)
}

func (s *ProcessingOptionsTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URL() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithFilename() {
	s.config().Base64URLIncludesFilename = true

	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png/puppy.jpg", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithReplacement() {
	s.config().URLReplacements = []config.URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg?param=value", imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURL() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s", originURL)

	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, Get(po, keys.Format, imagetype.Unknown))
}
func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscaped() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithReplacement() {
	s.config().URLReplacements = []config.URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg", imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseWithArgumentsSeparator() {
	s.config().ArgumentsSeparator = ","

	path := "/size,100,100,1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, GetInt(po, keys.Width, 0))
	s.Require().Equal(100, GetInt(po, keys.Height, 0))
	s.Require().True(Get(po, keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathFormat() {
	path := "/format:webp/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(imagetype.WEBP, Get(po, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePathResize() {
	path := "/resize:fill:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, Get(po, keys.ResizingType, ResizeFit))
	s.Require().Equal(100, GetInt(po, keys.Width, 0))
	s.Require().Equal(200, GetInt(po, keys.Height, 0))
	s.Require().True(Get(po, keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathResizingType() {
	path := "/resizing_type:fill/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, Get(po, keys.ResizingType, ResizeFit))
}

func (s *ProcessingOptionsTestSuite) TestParsePathSize() {
	path := "/size:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, GetInt(po, keys.Width, 0))
	s.Require().Equal(200, GetInt(po, keys.Height, 0))
	s.Require().True(Get(po, keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidth() {
	path := "/width:100/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, GetInt(po, keys.Width, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathHeight() {
	path := "/height:100/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, GetInt(po, keys.Height, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathEnlarge() {
	path := "/enlarge:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtend() {
	path := "/extend:1:so:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.ExtendEnabled, false))
	s.Require().Equal(GravitySouth, Get(po, keys.ExtendGravityType, GravityUnknown))
	s.Require().InDelta(10.0, GetFloat(po, keys.ExtendGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, GetFloat(po, keys.ExtendGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendSmartGravity() {
	path := "/extend:1:sm/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendReplicateGravity() {
	path := "/extend:1:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravity() {
	path := "/gravity:soea/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(GravitySouthEast, Get(po, keys.GravityType, GravityUnknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityFocusPoint() {
	path := "/gravity:fp:0.5:0.75/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(GravityFocusPoint, Get(po, keys.GravityType, GravityUnknown))
	s.Require().InDelta(0.5, GetFloat(po, keys.GravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(0.75, GetFloat(po, keys.GravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityReplicate() {
	path := "/gravity:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCrop() {
	path := "/crop:100:200/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(100.0, GetFloat(po, keys.CropWidth, 0.0), 0.0001)
	s.Require().InDelta(200.0, GetFloat(po, keys.CropHeight, 0.0), 0.0001)
	s.Require().Equal(GravityUnknown, Get(po, keys.CropGravityType, GravityUnknown))
	s.Require().InDelta(0.0, GetFloat(po, keys.CropGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(0.0, GetFloat(po, keys.CropGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravity() {
	path := "/crop:100:200:nowe:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(100.0, GetFloat(po, keys.CropWidth, 0.0), 0.0001)
	s.Require().InDelta(200.0, GetFloat(po, keys.CropHeight, 0.0), 0.0001)
	s.Require().Equal(GravityNorthWest, Get(po, keys.CropGravityType, GravityUnknown))
	s.Require().InDelta(10.0, GetFloat(po, keys.CropGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, GetFloat(po, keys.CropGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravityReplicate() {
	path := "/crop:100:200:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathQuality() {
	path := "/quality:55/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(55, GetInt(po, keys.Quality, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackground() {
	path := "/background:128:129:130/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.Flatten, false))
	s.Require().Equal(
		vips.Color{R: 128, G: 129, B: 130},
		Get(po, keys.Background, vips.Color{}),
	)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundHex() {
	path := "/background:ffddee/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.Flatten, false))
	s.Require().Equal(
		vips.Color{R: 0xff, G: 0xdd, B: 0xee},
		Get(po, keys.Background, vips.Color{}),
	)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundDisable() {
	path := "/background:fff/background:/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().False(Get(po, keys.Flatten, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathBlur() {
	path := "/blur:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(0.2, GetFloat(po, keys.Blur, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSharpen() {
	path := "/sharpen:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(0.2, GetFloat(po, keys.Sharpen, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDpr() {
	path := "/dpr:2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(2.0, GetFloat(po, keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWatermark() {
	path := "/watermark:0.5:soea:10:20:0.6/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.WatermarkEnabled, false))
	s.Require().Equal(GravitySouthEast, Get(po, keys.WatermarkPosition, GravityUnknown))
	s.Require().InDelta(10.0, GetFloat(po, keys.WatermarkXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, GetFloat(po, keys.WatermarkYOffset, 0.0), 0.0001)
	s.Require().InDelta(0.6, GetFloat(po, keys.WatermarkScale, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPreset() {
	s.config().Presets = []string{
		"test1=resizing_type:fill",
		"test2=blur:0.2/quality:50",
	}

	path := "/preset:test1:test2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, Get(po, keys.ResizingType, ResizeFit))
	s.Require().InDelta(float32(0.2), GetFloat(po, keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, GetInt(po, keys.Quality, 0))
	s.Require().ElementsMatch([]string{"test1", "test2"}, Get(po, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetDefault() {
	s.config().Presets = []string{
		"default=resizing_type:fill/blur:0.2/quality:50",
	}

	path := "/quality:70/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, Get(po, keys.ResizingType, ResizeFit))
	s.Require().InDelta(float32(0.2), GetFloat(po, keys.Blur, 0.0), 0.0001)
	s.Require().Equal(70, GetInt(po, keys.Quality, 0))
	s.Require().ElementsMatch([]string{"default"}, Get(po, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetLoopDetection() {
	s.config().Presets = []string{
		"test1=resizing_type:fill/preset:test2",
		"test2=blur:0.2/preset:test1",
	}

	path := "/preset:test1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().ElementsMatch([]string{"test1", "test2"}, Get(po, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathCachebuster() {
	path := "/cachebuster:123/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal("123", Get(po, keys.CacheBuster, ""))
}

func (s *ProcessingOptionsTestSuite) TestParsePathStripMetadata() {
	path := "/strip_metadata:true/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.StripMetadata, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpDetection() {
	s.config().AutoWebp = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	headers := http.Header{"Accept": []string{"image/webp"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferWebP, false))
	s.Require().False(Get(po, keys.EnforceWebP, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpEnforce() {
	s.config().EnforceWebp = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Accept": []string{"image/webp"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferWebP, false))
	s.Require().True(Get(po, keys.EnforceWebP, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathAvifDetection() {
	s.config().AutoWebp = true
	s.config().AutoAvif = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	headers := http.Header{"Accept": []string{"image/webp,image/avif"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferAvif, false))
	s.Require().False(Get(po, keys.EnforceAvif, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathAvifEnforce() {
	s.config().EnforceWebp = true
	s.config().EnforceAvif = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Accept": []string{"image/webp,image/avif"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferAvif, false))
	s.Require().True(Get(po, keys.EnforceAvif, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathJxlDetection() {
	s.config().AutoWebp = true
	s.config().AutoAvif = true
	s.config().AutoJxl = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	headers := http.Header{"Accept": []string{"image/webp,image/avif,image/jxl"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferJxl, false))
	s.Require().False(Get(po, keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathJxlEnforce() {
	s.config().EnforceWebp = true
	s.config().EnforceAvif = true
	s.config().EnforceJxl = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Accept": []string{"image/webp,image/avif,image/jxl"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(Get(po, keys.PreferJxl, false))
	s.Require().True(Get(po, keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeader() {
	s.config().EnableClientHints = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(100, GetInt(po, keys.Width, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderDisabled() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(0, GetInt(po, keys.Width, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderRedefine() {
	s.config().EnableClientHints = true

	path := "/width:150/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(150, GetInt(po, keys.Width, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeader() {
	s.config().EnableClientHints = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Dpr": []string{"2"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().InDelta(2.0, GetFloat(po, keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeaderDisabled() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Dpr": []string{"2"}}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().InDelta(1.0, GetFloat(po, keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthAndDprHeaderCombined() {
	s.config().EnableClientHints = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{
		"Width": []string{"100"},
		"Dpr":   []string{"2"},
	}
	po, _, err := s.factory().ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(50, GetInt(po, keys.Width, 0))
	s.Require().InDelta(2.0, GetFloat(po, keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessing() {
	path := "/skp:jpg:png/plain/http://images.dev/lorem/ipsum.jpg"

	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().ElementsMatch(
		[]imagetype.Type{imagetype.JPEG, imagetype.PNG},
		Get(po, keys.SkipProcessing, []imagetype.Type(nil)),
	)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessingInvalid() {
	path := "/skp:jpg:png:bad_format/plain/http://images.dev/lorem/ipsum.jpg"

	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err)
	s.Require().Equal("Invalid image format in skip_processing: bad_format", err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParseExpires() {
	path := "/exp:32503669200/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(time.Unix(32503669200, 0), GetTime(po, keys.Expires))
}

func (s *ProcessingOptionsTestSuite) TestParseExpiresExpired() {
	path := "/exp:1609448400/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().Error(err, "Expired URL")
}

func (s *ProcessingOptionsTestSuite) TestParsePathOnlyPresets() {
	s.config().OnlyPresets = true
	s.config().Presets = []string{
		"test1=blur:0.2",
		"test2=quality:50",
	}

	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := "/test1:test2/plain/" + originURL + "@png"

	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(0.2, GetFloat(po, keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, GetInt(po, keys.Quality, 0))
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
	s.Require().Equal(originURL, imageURL)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLOnlyPresets() {
	s.config().OnlyPresets = true
	s.config().Presets = []string{
		"test1=blur:0.2",
		"test2=quality:50",
	}

	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/test1:test2/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))

	po, imageURL, err := s.factory().ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(0.2, GetFloat(po, keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, GetInt(po, keys.Quality, 0))
	s.Require().Equal(imagetype.PNG, Get(po, keys.Format, imagetype.Unknown))
	s.Require().Equal(originURL, imageURL)
}

func (s *ProcessingOptionsTestSuite) TestParseAllowedOptions() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"

	testCases := []struct {
		options       string
		expectedError string
	}{
		{options: "w:100/h:200", expectedError: ""},
		{options: "w:100/h:200/blur:10", expectedError: "Forbidden processing option blur"},
		{options: "w:100/h:200/pr:test1", expectedError: ""},
		{options: "w:100/h:200/pr:test1/blur:10", expectedError: "Forbidden processing option blur"},
	}

	for _, tc := range testCases {
		s.Run(strings.ReplaceAll(tc.options, "/", "_"), func() {
			s.config().AllowedProcessingOptions = []string{"w", "h", "pr"}
			s.config().Presets = []string{
				"test1=blur:0.2",
			}

			path := fmt.Sprintf("/%s/%s.png", tc.options, base64.RawURLEncoding.EncodeToString([]byte(originURL)))
			_, _, err := s.factory().ParsePath(path, make(http.Header))

			if len(tc.expectedError) > 0 {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// func (s *ProcessingOptionsTestSuite) TestProcessingOptionsClone() {
// 	now := time.Now()

// 	// Create Options using factory
// 	original := s.factory().NewProcessingOptions()
// 	original.SkipProcessingFormats = []imagetype.Type{
// 		imagetype.PNG, imagetype.JPEG,
// 	}
// 	original.UsedPresets = []string{"preset1", "preset2"}
// 	original.Expires = &now

// 	// Clone the original
// 	cloned := original.clone()

// 	testutil.EqualButNotSame(s.T(), original, cloned)
// }

func TestProcessingOptions(t *testing.T) {
	suite.Run(t, new(ProcessingOptionsTestSuite))
}
