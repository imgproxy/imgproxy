package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProcessingOptionsTestSuite struct{ MainTestSuite }

func (s *ProcessingOptionsTestSuite) getRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	return req
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URL() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithoutExtension() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypeUnknown, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLInvalid() {
	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
	assert.Equal(s.T(), errInvalidImageURL.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURL() {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithoutExtension() {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s", imageURL))

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypeUnknown, getProcessingOptions(ctx).Format)
}
func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscaped() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLInvalid() {
	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL))
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
	assert.Equal(s.T(), errInvalidImageURL.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedInvalid() {
	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)))
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
	assert.Equal(s.T(), errInvalidImageURL.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePathBasic() {
	req := s.getRequest("http://example.com/unsafe/fill/100/200/noea/1/plain/http://images.dev/lorem/ipsum.jpg@png")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.Resize)
	assert.Equal(s.T(), 100, po.Width)
	assert.Equal(s.T(), 200, po.Height)
	assert.Equal(s.T(), gravityNorthEast, po.Gravity.Type)
	assert.True(s.T(), po.Enlarge)
	assert.Equal(s.T(), imageTypePNG, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedFormat() {
	req := s.getRequest("http://example.com/unsafe/format:webp/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), imageTypeWEBP, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedResize() {
	req := s.getRequest("http://example.com/unsafe/resize:fill:100:200:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.Resize)
	assert.Equal(s.T(), 100, po.Width)
	assert.Equal(s.T(), 200, po.Height)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedResizingType() {
	req := s.getRequest("http://example.com/unsafe/resizing_type:fill/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.Resize)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedSize() {
	req := s.getRequest("http://example.com/unsafe/size:100:200:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
	assert.Equal(s.T(), 200, po.Height)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedWidth() {
	req := s.getRequest("http://example.com/unsafe/width:100/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedHeight() {
	req := s.getRequest("http://example.com/unsafe/height:100/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Height)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedEnlarge() {
	req := s.getRequest("http://example.com/unsafe/enlarge:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedGravity() {
	req := s.getRequest("http://example.com/unsafe/gravity:soea/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), gravitySouthEast, po.Gravity.Type)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedGravityFocuspoint() {
	req := s.getRequest("http://example.com/unsafe/gravity:fp:0.5:0.75/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), gravityFocusPoint, po.Gravity.Type)
	assert.Equal(s.T(), 0.5, po.Gravity.X)
	assert.Equal(s.T(), 0.75, po.Gravity.Y)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedQuality() {
	req := s.getRequest("http://example.com/unsafe/quality:55/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 55, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedBackground() {
	req := s.getRequest("http://example.com/unsafe/background:128:129:130/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Flatten)
	assert.Equal(s.T(), uint8(128), po.Background.R)
	assert.Equal(s.T(), uint8(129), po.Background.G)
	assert.Equal(s.T(), uint8(130), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedBackgroundHex() {
	req := s.getRequest("http://example.com/unsafe/background:ffddee/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Flatten)
	assert.Equal(s.T(), uint8(0xff), po.Background.R)
	assert.Equal(s.T(), uint8(0xdd), po.Background.G)
	assert.Equal(s.T(), uint8(0xee), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedBackgroundDisable() {
	req := s.getRequest("http://example.com/unsafe/background:fff/background:/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.False(s.T(), po.Flatten)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedBlur() {
	req := s.getRequest("http://example.com/unsafe/blur:0.2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedSharpen() {
	req := s.getRequest("http://example.com/unsafe/sharpen:0.2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Sharpen)
}
func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedDpr() {
	req := s.getRequest("http://example.com/unsafe/dpr:2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 2.0, po.Dpr)
}
func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedWatermark() {
	req := s.getRequest("http://example.com/unsafe/watermark:0.5:soea:10:20:0.6/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Watermark.Enabled)
	assert.Equal(s.T(), gravitySouthEast, po.Watermark.Gravity)
	assert.Equal(s.T(), 10, po.Watermark.OffsetX)
	assert.Equal(s.T(), 20, po.Watermark.OffsetY)
	assert.Equal(s.T(), 0.6, po.Watermark.Scale)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedPreset() {
	conf.Presets["test1"] = urlOptions{
		"resizing_type": []string{"fill"},
	}

	conf.Presets["test2"] = urlOptions{
		"blur":    []string{"0.2"},
		"quality": []string{"50"},
	}

	req := s.getRequest("http://example.com/unsafe/preset:test1:test2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.Resize)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetDefault() {
	conf.Presets["default"] = urlOptions{
		"resizing_type": []string{"fill"},
		"blur":          []string{"0.2"},
		"quality":       []string{"50"},
	}

	req := s.getRequest("http://example.com/unsafe/quality:70/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.Resize)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 70, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedPresetLoopDetection() {
	conf.Presets["test1"] = urlOptions{
		"resizing_type": []string{"fill"},
	}

	conf.Presets["test2"] = urlOptions{
		"blur":    []string{"0.2"},
		"quality": []string{"50"},
	}

	req := s.getRequest("http://example.com/unsafe/preset:test1:test2:test1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	require.ElementsMatch(s.T(), po.UsedPresets, []string{"test1", "test2"})
}

func (s *ProcessingOptionsTestSuite) TestParsePathAdvancedCachebuster() {
	req := s.getRequest("http://example.com/unsafe/cachebuster:123/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), "123", po.CacheBuster)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpDetection() {
	conf.EnableWebpDetection = true

	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg")
	req.Header.Set("Accept", "image/webp")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), true, po.PreferWebP)
	assert.Equal(s.T(), false, po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpEnforce() {
	conf.EnforceWebp = true

	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Accept", "image/webp")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), true, po.PreferWebP)
	assert.Equal(s.T(), true, po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderDisabled() {
	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 0, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderRedefine() {
	conf.EnableClientHints = true

	req := s.getRequest("http://example.com/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 150, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeaderDisabled() {
	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 0, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeaderRedefine() {
	conf.EnableClientHints = true

	req := s.getRequest("http://example.com/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 150, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("DPR", "2")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 2.0, po.Dpr)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeaderDisabled() {
	req := s.getRequest("http://example.com/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("DPR", "2")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 1.0, po.Dpr)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSigned() {
	conf.Keys = []securityKey{securityKey("test-key")}
	conf.Salts = []securityKey{securityKey("test-salt")}
	conf.AllowInsecure = false

	req := s.getRequest("http://example.com/HcvNognEV1bW6f8zRqxNYuOkV0IUf1xloRb57CzbT4g/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	_, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSignedInvalid() {
	conf.Keys = []securityKey{securityKey("test-key")}
	conf.Salts = []securityKey{securityKey("test-salt")}
	conf.AllowInsecure = false

	req := s.getRequest("http://example.com/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
	assert.Equal(s.T(), errInvalidSignature.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePathOnlyPresets() {
	conf.OnlyPresets = true
	conf.Presets["test1"] = urlOptions{
		"blur": []string{"0.2"},
	}
	conf.Presets["test2"] = urlOptions{
		"quality": []string{"50"},
	}

	req := s.getRequest("http://example.com/unsafe/test1:test2/plain/http://images.dev/lorem/ipsum.jpg")

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLOnlyPresets() {
	conf.OnlyPresets = true
	conf.Presets["test1"] = urlOptions{
		"blur": []string{"0.2"},
	}
	conf.Presets["test2"] = urlOptions{
		"quality": []string{"50"},
	}

	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("http://example.com/unsafe/test1:test2/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}
func TestProcessingOptions(t *testing.T) {
	suite.Run(t, new(ProcessingOptionsTestSuite))
}
