package options

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type ProcessingOptionsTestSuite struct{ suite.Suite }

func (s *ProcessingOptionsTestSuite) SetupTest() {
	config.Reset()
	// Reset presets
	presets = make(map[string]urlOptions)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URL() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithBase() {
	config.BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", config.BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithReplacement() {
	config.URLReplacements = []config.URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg?param=value", imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURL() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithoutExtension() {
	originURL := "http://images.dev/lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s", originURL)

	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.Unknown, po.Format)
}
func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscaped() {
	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(originURL, imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithBase() {
	config.BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", config.BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithReplacement() {
	config.URLReplacements = []config.URLReplacement{
		{Regexp: regexp.MustCompile("^test://([^/]*)/"), Replacement: "test2://images.dev/${1}/dolor/"},
		{Regexp: regexp.MustCompile("^test2://"), Replacement: "http://"},
	}

	originURL := "test://lorem/ipsum.jpg"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", originURL)
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal("http://images.dev/lorem/dolor/ipsum.jpg", imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedWithBase() {
	config.BaseURL = "http://images.dev/"

	originURL := "lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/size:100:100/plain/%s@png", url.PathEscape(originURL))
	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf("%s%s", config.BaseURL, originURL), imageURL)
	s.Require().Equal(imagetype.PNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParseWithArgumentsSeparator() {
	config.ArgumentsSeparator = ","

	path := "/size,100,100,1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, po.Width)
	s.Require().Equal(100, po.Height)
	s.Require().True(po.Enlarge)
}

// func (s *ProcessingOptionsTestSuite) TestParseURLAllowedSource() {
// 	config.AllowedSources = []string{"local://", "http://images.dev/"}

// 	path := "/plain/http://images.dev/lorem/ipsum.jpg"
// 	_, _, err := ParsePath(path, make(http.Header))

// 	s.Require().NoError(err)
// }

// func (s *ProcessingOptionsTestSuite) TestParseURLNotAllowedSource() {
// 	config.AllowedSources = []string{"local://", "http://images.dev/"}

// 	path := "/plain/s3://images/lorem/ipsum.jpg"
// 	_, _, err := ParsePath(path, make(http.Header))

// 	s.Require().Error(err)
// }

func (s *ProcessingOptionsTestSuite) TestParsePathFormat() {
	path := "/format:webp/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(imagetype.WEBP, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePathResize() {
	path := "/resize:fill:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, po.ResizingType)
	s.Require().Equal(100, po.Width)
	s.Require().Equal(200, po.Height)
	s.Require().True(po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathResizingType() {
	path := "/resizing_type:fill/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, po.ResizingType)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSize() {
	path := "/size:100:200:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, po.Width)
	s.Require().Equal(200, po.Height)
	s.Require().True(po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidth() {
	path := "/width:100/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathHeight() {
	path := "/height:100/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(100, po.Height)
}

func (s *ProcessingOptionsTestSuite) TestParsePathEnlarge() {
	path := "/enlarge:1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtend() {
	path := "/extend:1:so:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.Extend.Enabled)
	s.Require().Equal(GravitySouth, po.Extend.Gravity.Type)
	s.Require().InDelta(10.0, po.Extend.Gravity.X, 0.0001)
	s.Require().InDelta(20.0, po.Extend.Gravity.Y, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendSmartGravity() {
	path := "/extend:1:sm/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtendReplicateGravity() {
	path := "/extend:1:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravity() {
	path := "/gravity:soea/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(GravitySouthEast, po.Gravity.Type)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityFocusPoint() {
	path := "/gravity:fp:0.5:0.75/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(GravityFocusPoint, po.Gravity.Type)
	s.Require().InDelta(0.5, po.Gravity.X, 0.0001)
	s.Require().InDelta(0.75, po.Gravity.Y, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityReplicate() {
	path := "/gravity:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCrop() {
	path := "/crop:100:200/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(100.0, po.Crop.Width, 0.0001)
	s.Require().InDelta(200.0, po.Crop.Height, 0.0001)
	s.Require().Equal(GravityUnknown, po.Crop.Gravity.Type)
	s.Require().InDelta(0.0, po.Crop.Gravity.X, 0.0001)
	s.Require().InDelta(0.0, po.Crop.Gravity.Y, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravity() {
	path := "/crop:100:200:nowe:10:20/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(100.0, po.Crop.Width, 0.0001)
	s.Require().InDelta(200.0, po.Crop.Height, 0.0001)
	s.Require().Equal(GravityNorthWest, po.Crop.Gravity.Type)
	s.Require().InDelta(10.0, po.Crop.Gravity.X, 0.0001)
	s.Require().InDelta(20.0, po.Crop.Gravity.Y, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathCropGravityReplicate() {
	path := "/crop:100:200:re/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathQuality() {
	path := "/quality:55/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(55, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackground() {
	path := "/background:128:129:130/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.Flatten)
	s.Require().Equal(uint8(128), po.Background.R)
	s.Require().Equal(uint8(129), po.Background.G)
	s.Require().Equal(uint8(130), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundHex() {
	path := "/background:ffddee/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.Flatten)
	s.Require().Equal(uint8(0xff), po.Background.R)
	s.Require().Equal(uint8(0xdd), po.Background.G)
	s.Require().Equal(uint8(0xee), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundDisable() {
	path := "/background:fff/background:/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().False(po.Flatten)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBlur() {
	path := "/blur:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(float32(0.2), po.Blur, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSharpen() {
	path := "/sharpen:0.2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(float32(0.2), po.Sharpen, 0.0001)
}
func (s *ProcessingOptionsTestSuite) TestParsePathDpr() {
	path := "/dpr:2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(2.0, po.Dpr, 0.0001)
}
func (s *ProcessingOptionsTestSuite) TestParsePathWatermark() {
	path := "/watermark:0.5:soea:10:20:0.6/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.Watermark.Enabled)
	s.Require().Equal(GravitySouthEast, po.Watermark.Gravity.Type)
	s.Require().InDelta(10.0, po.Watermark.Gravity.X, 0.0001)
	s.Require().InDelta(20.0, po.Watermark.Gravity.Y, 0.0001)
	s.Require().InDelta(0.6, po.Watermark.Scale, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPreset() {
	presets["test1"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
	}

	presets["test2"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	path := "/preset:test1:test2/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, po.ResizingType)
	s.Require().InDelta(float32(0.2), po.Blur, 0.0001)
	s.Require().Equal(50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetDefault() {
	presets["default"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	path := "/quality:70/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal(ResizeFill, po.ResizingType)
	s.Require().InDelta(float32(0.2), po.Blur, 0.0001)
	s.Require().Equal(70, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetLoopDetection() {
	presets["test1"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
		urlOption{Name: "preset", Args: []string{"test2"}},
	}

	presets["test2"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "preset", Args: []string{"test1"}},
	}

	path := "/preset:test1/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().ElementsMatch(po.UsedPresets, []string{"test1", "test2"})
}

func (s *ProcessingOptionsTestSuite) TestParsePathCachebuster() {
	path := "/cachebuster:123/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal("123", po.CacheBuster)
}

func (s *ProcessingOptionsTestSuite) TestParsePathStripMetadata() {
	path := "/strip_metadata:true/plain/http://images.dev/lorem/ipsum.jpg"
	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().True(po.StripMetadata)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpDetection() {
	config.EnableWebpDetection = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg"
	headers := http.Header{"Accept": []string{"image/webp"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(po.PreferWebP)
	s.Require().False(po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpEnforce() {
	config.EnforceWebp = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Accept": []string{"image/webp"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().True(po.PreferWebP)
	s.Require().True(po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeader() {
	config.EnableClientHints = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderDisabled() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(0, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderRedefine() {
	config.EnableClientHints = true

	path := "/width:150/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Width": []string{"100"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().Equal(150, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeader() {
	config.EnableClientHints = true

	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Dpr": []string{"2"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().InDelta(2.0, po.Dpr, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeaderDisabled() {
	path := "/plain/http://images.dev/lorem/ipsum.jpg@png"
	headers := http.Header{"Dpr": []string{"2"}}
	po, _, err := ParsePath(path, headers)

	s.Require().NoError(err)

	s.Require().InDelta(1.0, po.Dpr, 0.0001)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessing() {
	path := "/skp:jpg:png/plain/http://images.dev/lorem/ipsum.jpg"

	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().Equal([]imagetype.Type{imagetype.JPEG, imagetype.PNG}, po.SkipProcessingFormats)
}

func (s *ProcessingOptionsTestSuite) TestParseSkipProcessingInvalid() {
	path := "/skp:jpg:png:bad_format/plain/http://images.dev/lorem/ipsum.jpg"

	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
	s.Require().Equal("Invalid image format in skip processing: bad_format", err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParseExpires() {
	path := "/exp:32503669200/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)
}

func (s *ProcessingOptionsTestSuite) TestParseExpiresExpired() {
	path := "/exp:1609448400/plain/http://images.dev/lorem/ipsum.jpg"
	_, _, err := ParsePath(path, make(http.Header))

	s.Require().Error(err)
	s.Require().Equal(errExpiredURL.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePathOnlyPresets() {
	config.OnlyPresets = true
	presets["test1"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
	}
	presets["test2"] = urlOptions{
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	path := "/test1:test2/plain/http://images.dev/lorem/ipsum.jpg"

	po, _, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(float32(0.2), po.Blur, 0.0001)
	s.Require().Equal(50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLOnlyPresets() {
	config.OnlyPresets = true
	presets["test1"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
	}
	presets["test2"] = urlOptions{
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	originURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	path := fmt.Sprintf("/test1:test2/%s.png", base64.RawURLEncoding.EncodeToString([]byte(originURL)))

	po, imageURL, err := ParsePath(path, make(http.Header))

	s.Require().NoError(err)

	s.Require().InDelta(float32(0.2), po.Blur, 0.0001)
	s.Require().Equal(50, po.Quality)
	s.Require().Equal(originURL, imageURL)
}

func TestProcessingOptions(t *testing.T) {
	suite.Run(t, new(ProcessingOptionsTestSuite))
}
