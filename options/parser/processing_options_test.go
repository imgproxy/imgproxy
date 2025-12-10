package optionsparser

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips/color"
	"github.com/stretchr/testify/suite"
)

type ProcessingOptionsTestSuite struct {
	testutil.LazySuite

	config testutil.LazyObj[*Config]
	parser testutil.LazyObj[*Parser]
}

func (s *ProcessingOptionsTestSuite) SetupSuite() {
	s.config, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Config, error) {
			c := NewDefaultConfig()
			return &c, nil
		},
	)

	s.parser, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Parser, error) {
			return New(s.config())
		},
	)
}

func (s *ProcessingOptionsTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URL() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithFilename() {
	s.config().Base64URLIncludesFilename = true

	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png/puppy.jpg", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithReplacement() {
	s.config().URLReplacements = []URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg?param=value", imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURL() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s", originURL)

	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, options.Get(o, keys.Format, imagetype.Unknown))
}
func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscaped() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithReplacement() {
	s.config().URLReplacements = []URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg", imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedWithBase() {
	s.config().BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", s.config().BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParseWithArgumentsSeparator() {
	s.config().ArgumentsSeparator = ","

	path := "/size,100,100,1/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(100, o.GetInt(keys.Width, 0))
	s.Require().Equal(100, o.GetInt(keys.Height, 0))
	s.Require().True(o.GetBool(keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathFormat() {
	path := "/format:webp/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(imagetype.WEBP, options.Get(o, keys.Format, imagetype.Unknown))
}

func (s *ProcessingOptionsTestSuite) TestParsePathResize() {
	path := "/resize:fill:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(processing.ResizeFill, options.Get(o, keys.ResizingType, processing.ResizeFit))
	s.Require().Equal(100, o.GetInt(keys.Width, 0))
	s.Require().Equal(200, o.GetInt(keys.Height, 0))
	s.Require().True(o.GetBool(keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathResizingType() {
	path := "/resizing_type:fill/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(processing.ResizeFill, options.Get(o, keys.ResizingType, processing.ResizeFit))
}

func (s *ProcessingOptionsTestSuite) TestParsePathSize() {
	path := "/size:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(100, o.GetInt(keys.Width, 0))
	s.Require().Equal(200, o.GetInt(keys.Height, 0))
	s.Require().True(o.GetBool(keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidth() {
	path := "/width:100/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(100, o.GetInt(keys.Width, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathHeight() {
	path := "/height:100/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(100, o.GetInt(keys.Height, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathEnlarge() {
	path := "/enlarge:1/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().True(o.GetBool(keys.Enlarge, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtend() {
	path := "/extend:1:so:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().True(o.GetBool(keys.ExtendEnabled, false))
	s.Require().Equal(
		processing.GravitySouth,
		options.Get(o, keys.ExtendGravityType, processing.GravityUnknown),
	)
	s.Require().InDelta(10.0, o.GetFloat(keys.ExtendGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, o.GetFloat(keys.ExtendGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendSmartGravity() {
	path := "/extend:1:sm/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.parser().ParsePath(path, nil)

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendReplicateGravity() {
	path := "/extend:1:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.parser().ParsePath(path, nil)

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravity() {
	path := "/gravity:soea/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(
		processing.GravitySouthEast,
		options.Get(o, keys.GravityType, processing.GravityUnknown),
	)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityFocusPoint() {
	path := "/gravity:fp:0.5:0.75/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(processing.GravityFocusPoint, options.Get(o, keys.GravityType, processing.GravityUnknown))
	s.Require().InDelta(0.5, o.GetFloat(keys.GravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(0.75, o.GetFloat(keys.GravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityReplicate() {
	path := "/gravity:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.parser().ParsePath(path, nil)

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCrop() {
	path := "/crop:100:200/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(100.0, o.GetFloat(keys.CropWidth, 0.0), 0.0001)
	s.Require().InDelta(200.0, o.GetFloat(keys.CropHeight, 0.0), 0.0001)
	s.Require().Equal(
		processing.GravityUnknown,
		options.Get(o, keys.CropGravityType, processing.GravityUnknown),
	)
	s.Require().InDelta(0.0, o.GetFloat(keys.CropGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(0.0, o.GetFloat(keys.CropGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravity() {
	//nolint:misspell
	path := "/crop:100:200:nowe:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(100.0, o.GetFloat(keys.CropWidth, 0.0), 0.0001)
	s.Require().InDelta(200.0, o.GetFloat(keys.CropHeight, 0.0), 0.0001)
	s.Require().Equal(
		processing.GravityNorthWest,
		options.Get(o, keys.CropGravityType, processing.GravityUnknown),
	)
	s.Require().InDelta(10.0, o.GetFloat(keys.CropGravityXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, o.GetFloat(keys.CropGravityYOffset, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravityReplicate() {
	path := "/crop:100:200:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.parser().ParsePath(path, nil)

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathQuality() {
	path := "/quality:55/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(55, o.GetInt(keys.Quality, 0))
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackground() {
	path := "/background:128:129:130/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(
		color.RGB{R: 128, G: 129, B: 130},
		options.Get(o, keys.Background, color.RGB{}),
	)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundHex() {
	path := "/background:ffddee/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(
		color.RGB{R: 0xff, G: 0xdd, B: 0xee},
		options.Get(o, keys.Background, color.RGB{}),
	)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundDisable() {
	path := "/background:fff/background:/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().False(o.Has(keys.Background))
}

func (s *ProcessingOptionsTestSuite) TestParsePathBlur() {
	path := "/blur:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(0.2, o.GetFloat(keys.Blur, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSharpen() {
	path := "/sharpen:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(0.2, o.GetFloat(keys.Sharpen, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDpr() {
	path := "/dpr:2/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(2.0, o.GetFloat(keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWatermark() {
	path := "/watermark:0.5:soea:10:20:0.6/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(0.5, o.GetFloat(keys.WatermarkOpacity, 0.0), 0.0001)
	s.Require().Equal(
		processing.GravitySouthEast,
		options.Get(o, keys.WatermarkPosition, processing.GravityUnknown),
	)
	s.Require().InDelta(10.0, o.GetFloat(keys.WatermarkXOffset, 0.0), 0.0001)
	s.Require().InDelta(20.0, o.GetFloat(keys.WatermarkYOffset, 0.0), 0.0001)
	s.Require().InDelta(0.6, o.GetFloat(keys.WatermarkScale, 0.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPreset() {
	s.config().Presets = []string{
		"test1=resizing_type:fill",
		"test2=blur:0.2/quality:50",
	}

	path := "/preset:test1:test2/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(processing.ResizeFill, options.Get(o, keys.ResizingType, processing.ResizeFit))
	s.Require().InDelta(float32(0.2), o.GetFloat(keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, o.GetInt(keys.Quality, 0))
	s.Require().ElementsMatch([]string{"test1", "test2"}, options.Get(o, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetDefault() {
	s.config().Presets = []string{
		"default=resizing_type:fill/blur:0.2/quality:50",
	}

	path := "/quality:70/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal(processing.ResizeFill, options.Get(o, keys.ResizingType, processing.ResizeFit))
	s.Require().InDelta(float32(0.2), o.GetFloat(keys.Blur, 0.0), 0.0001)
	s.Require().Equal(70, o.GetInt(keys.Quality, 0))
	s.Require().ElementsMatch([]string{"default"}, options.Get(o, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetLoopDetection() {
	s.config().Presets = []string{
		"test1=resizing_type:fill/preset:test2",
		"test2=blur:0.2/preset:test1",
	}

	path := "/preset:test1/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().ElementsMatch([]string{"test1", "test2"}, options.Get(o, keys.UsedPresets, []string{}))
}

func (s *ProcessingOptionsTestSuite) TestParsePathCachebuster() {
	path := "/cachebuster:123/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().Equal("123", options.Get(o, keys.CacheBuster, ""))
}

func (s *ProcessingOptionsTestSuite) TestParsePathStripMetadata() {
	path := "/strip_metadata:true/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().True(o.GetBool(keys.StripMetadata, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpDetection() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	features := clientfeatures.Features{PreferWebP: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().True(o.GetBool(keys.PreferWebP, false))
	s.Require().False(o.GetBool(keys.EnforceWebP, false))
	s.Require().False(o.GetBool(keys.PreferAvif, false))
	s.Require().False(o.GetBool(keys.EnforceAvif, false))
	s.Require().False(o.GetBool(keys.PreferJxl, false))
	s.Require().False(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpEnforce() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	features := clientfeatures.Features{EnforceWebP: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().True(o.GetBool(keys.PreferWebP, false))
	s.Require().True(o.GetBool(keys.EnforceWebP, false))
	s.Require().False(o.GetBool(keys.PreferAvif, false))
	s.Require().False(o.GetBool(keys.EnforceAvif, false))
	s.Require().False(o.GetBool(keys.PreferJxl, false))
	s.Require().False(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathAvifDetection() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	features := clientfeatures.Features{PreferAvif: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().False(o.GetBool(keys.PreferWebP, false))
	s.Require().False(o.GetBool(keys.EnforceWebP, false))
	s.Require().True(o.GetBool(keys.PreferAvif, false))
	s.Require().False(o.GetBool(keys.EnforceAvif, false))
	s.Require().False(o.GetBool(keys.PreferJxl, false))
	s.Require().False(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathAvifEnforce() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	features := clientfeatures.Features{EnforceAvif: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().False(o.GetBool(keys.PreferWebP, false))
	s.Require().False(o.GetBool(keys.EnforceWebP, false))
	s.Require().True(o.GetBool(keys.PreferAvif, false))
	s.Require().True(o.GetBool(keys.EnforceAvif, false))
	s.Require().False(o.GetBool(keys.PreferJxl, false))
	s.Require().False(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathJxlDetection() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	features := clientfeatures.Features{PreferJxl: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().False(o.GetBool(keys.PreferWebP, false))
	s.Require().False(o.GetBool(keys.EnforceWebP, false))
	s.Require().False(o.GetBool(keys.PreferAvif, false))
	s.Require().False(o.GetBool(keys.EnforceAvif, false))
	s.Require().True(o.GetBool(keys.PreferJxl, false))
	s.Require().False(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathJxlEnforce() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	features := clientfeatures.Features{EnforceJxl: true}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().False(o.GetBool(keys.PreferWebP, false))
	s.Require().False(o.GetBool(keys.EnforceWebP, false))
	s.Require().False(o.GetBool(keys.PreferAvif, false))
	s.Require().False(o.GetBool(keys.EnforceAvif, false))
	s.Require().True(o.GetBool(keys.PreferJxl, false))
	s.Require().True(o.GetBool(keys.EnforceJxl, false))
}

func (s *ProcessingOptionsTestSuite) TestParsePathClientHints() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"

	testCases := []struct {
		name     string
		features clientfeatures.Features
		width    int
		dpr      float64
	}{
		{
			name:     "NoClientHints",
			features: clientfeatures.Features{},
			width:    0,
			dpr:      1.0,
		},
		{
			name:     "WidthOnly",
			features: clientfeatures.Features{ClientHintsWidth: 100},
			width:    100,
			dpr:      1.0,
		},
		{
			name:     "DprOnly",
			features: clientfeatures.Features{ClientHintsDPR: 2.0},
			width:    0,
			dpr:      2.0,
		},
		{
			name:     "WidthAndDpr",
			features: clientfeatures.Features{ClientHintsWidth: 100, ClientHintsDPR: 2.0},
			width:    50,
			dpr:      2.0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			o, _, err := s.parser().ParsePath(path, &tc.features)

			s.Require().NoError(err)
			s.Require().Equal(tc.width, o.GetInt(keys.Width, 0))
			s.Require().InDelta(tc.dpr, o.GetFloat(keys.Dpr, 1.0), 0.0001)
		})
	}
}

func (s *ProcessingOptionsTestSuite) TestParsePathClientHintsRedefine() {
	path := "/width:150/dpr:3.0/plain/http://images.dev/lorem/ipsum.jpg@png"
	features := clientfeatures.Features{
		ClientHintsWidth: 100,
		ClientHintsDPR:   2.0,
	}
	o, _, err := s.parser().ParsePath(path, &features)

	s.Require().NoError(err)

	s.Require().Equal(150, o.GetInt(keys.Width, 0))
	s.Require().InDelta(3.0, o.GetFloat(keys.Dpr, 1.0), 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessing() {
	path := "/skp:jpg:png/plain/http://images.dev/lorem/ipsum.jpg"

	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().ElementsMatch(
		[]imagetype.Type{imagetype.JPEG, imagetype.PNG},
		options.Get(o, keys.SkipProcessing, []imagetype.Type(nil)),
	)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessingInvalid() {
	path := "/skp:jpg:png:bad_format/plain/http://images.dev/lorem/ipsum.jpg"

	_, _, err := s.parser().ParsePath(path, nil)

	s.Require().Error(err)
	s.Require().Equal("Invalid image format in skip_processing: bad_format", err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParseExpires() {
	path := "/exp:32503669200/plain/http://images.dev/lorem/ipsum.jpg"
	o, _, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)
	s.Require().Equal(time.Unix(32503669200, 0), o.GetTime(keys.Expires))
}

func (s *ProcessingOptionsTestSuite) TestParseExpiresExpired() {
	path := "/exp:1609448400/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := s.parser().ParsePath(path, nil)

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

	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(0.2, o.GetFloat(keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, o.GetInt(keys.Quality, 0))
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
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

	o, imageURL, err := s.parser().ParsePath(path, nil)

	s.Require().NoError(err)

	s.Require().InDelta(0.2, o.GetFloat(keys.Blur, 0.0), 0.0001)
	s.Require().Equal(50, o.GetInt(keys.Quality, 0))
	s.Require().Equal(imagetype.PNG, options.Get(o, keys.Format, imagetype.Unknown))
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
			_, _, err := s.parser().ParsePath(path, nil)

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

// 	// Create Options using parser
// 	original := s.parser().NewProcessingOptions()
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
