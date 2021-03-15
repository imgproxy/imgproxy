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

func (s *ProcessingOptionsTestSuite) getRequest(uri string) *http.Request {
	return &http.Request{Method: "GET", RequestURI: uri, Header: make(http.Header)}
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URL() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithoutExtension() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypeUnknown, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURL() {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/plain/%s@png", imageURL))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithoutExtension() {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/plain/%s", imageURL))

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypeUnknown, getProcessingOptions(ctx).Format)
}
func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscaped() {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), imageURL, getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/plain/%s@png", imageURL))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePlainURLEscapedWithBase() {
	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)))
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
	assert.Equal(s.T(), fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
	assert.Equal(s.T(), imageTypePNG, getProcessingOptions(ctx).Format)
}

func (s *ProcessingOptionsTestSuite) TestParseURLAllowedSource() {
	conf.AllowedSources = []string{"local://", "http://images.dev/"}

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg")
	_, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
}

func (s *ProcessingOptionsTestSuite) TestParseURLNotAllowedSource() {
	conf.AllowedSources = []string{"local://", "http://images.dev/"}

	req := s.getRequest("/unsafe/plain/s3://images/lorem/ipsum.jpg")
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathFormat() {
	req := s.getRequest("/unsafe/format:webp/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), imageTypeWEBP, po.Format)
}

func (s *ProcessingOptionsTestSuite) TestParsePathResize() {
	req := s.getRequest("/unsafe/resize:fill:100:200:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.ResizingType)
	assert.Equal(s.T(), 100, po.Width)
	assert.Equal(s.T(), 200, po.Height)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathResizingType() {
	req := s.getRequest("/unsafe/resizing_type:fill/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.ResizingType)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSize() {
	req := s.getRequest("/unsafe/size:100:200:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
	assert.Equal(s.T(), 200, po.Height)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidth() {
	req := s.getRequest("/unsafe/width:100/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathHeight() {
	req := s.getRequest("/unsafe/height:100/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Height)
}

func (s *ProcessingOptionsTestSuite) TestParsePathEnlarge() {
	req := s.getRequest("/unsafe/enlarge:1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Enlarge)
}

func (s *ProcessingOptionsTestSuite) TestParsePathExtend() {
	req := s.getRequest("/unsafe/extend:1:so:10:20/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), true, po.Extend.Enabled)
	assert.Equal(s.T(), gravitySouth, po.Extend.Gravity.Type)
	assert.Equal(s.T(), 10.0, po.Extend.Gravity.X)
	assert.Equal(s.T(), 20.0, po.Extend.Gravity.Y)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravity() {
	req := s.getRequest("/unsafe/gravity:soea/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), gravitySouthEast, po.Gravity.Type)
}

func (s *ProcessingOptionsTestSuite) TestParsePathGravityFocuspoint() {
	req := s.getRequest("/unsafe/gravity:fp:0.5:0.75/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), gravityFocusPoint, po.Gravity.Type)
	assert.Equal(s.T(), 0.5, po.Gravity.X)
	assert.Equal(s.T(), 0.75, po.Gravity.Y)
}

func (s *ProcessingOptionsTestSuite) TestParsePathQuality() {
	req := s.getRequest("/unsafe/quality:55/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 55, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackground() {
	req := s.getRequest("/unsafe/background:128:129:130/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Flatten)
	assert.Equal(s.T(), uint8(128), po.Background.R)
	assert.Equal(s.T(), uint8(129), po.Background.G)
	assert.Equal(s.T(), uint8(130), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundHex() {
	req := s.getRequest("/unsafe/background:ffddee/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Flatten)
	assert.Equal(s.T(), uint8(0xff), po.Background.R)
	assert.Equal(s.T(), uint8(0xdd), po.Background.G)
	assert.Equal(s.T(), uint8(0xee), po.Background.B)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBackgroundDisable() {
	req := s.getRequest("/unsafe/background:fff/background:/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.False(s.T(), po.Flatten)
}

func (s *ProcessingOptionsTestSuite) TestParsePathBlur() {
	req := s.getRequest("/unsafe/blur:0.2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSharpen() {
	req := s.getRequest("/unsafe/sharpen:0.2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Sharpen)
}
func (s *ProcessingOptionsTestSuite) TestParsePathDpr() {
	req := s.getRequest("/unsafe/dpr:2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 2.0, po.Dpr)
}
func (s *ProcessingOptionsTestSuite) TestParsePathWatermark() {
	req := s.getRequest("/unsafe/watermark:0.5:soea:10:20:0.6/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.Watermark.Enabled)
	assert.Equal(s.T(), gravitySouthEast, po.Watermark.Gravity.Type)
	assert.Equal(s.T(), 10.0, po.Watermark.Gravity.X)
	assert.Equal(s.T(), 20.0, po.Watermark.Gravity.Y)
	assert.Equal(s.T(), 0.6, po.Watermark.Scale)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPreset() {
	conf.Presets["test1"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
	}

	conf.Presets["test2"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	req := s.getRequest("/unsafe/preset:test1:test2/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.ResizingType)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetDefault() {
	conf.Presets["default"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	req := s.getRequest("/unsafe/quality:70/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), resizeFill, po.ResizingType)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 70, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParsePathPresetLoopDetection() {
	conf.Presets["test1"] = urlOptions{
		urlOption{Name: "resizing_type", Args: []string{"fill"}},
	}

	conf.Presets["test2"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	req := s.getRequest("/unsafe/preset:test1:test2:test1/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	require.ElementsMatch(s.T(), po.UsedPresets, []string{"test1", "test2"})
}

func (s *ProcessingOptionsTestSuite) TestParsePathCachebuster() {
	req := s.getRequest("/unsafe/cachebuster:123/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), "123", po.CacheBuster)
}

func (s *ProcessingOptionsTestSuite) TestParsePathStripMetadata() {
	req := s.getRequest("/unsafe/strip_metadata:true/plain/http://images.dev/lorem/ipsum.jpg")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.True(s.T(), po.StripMetadata)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpDetection() {
	conf.EnableWebpDetection = true

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg")
	req.Header.Set("Accept", "image/webp")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), true, po.PreferWebP)
	assert.Equal(s.T(), false, po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWebpEnforce() {
	conf.EnforceWebp = true

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Accept", "image/webp")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), true, po.PreferWebP)
	assert.Equal(s.T(), true, po.EnforceWebP)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderDisabled() {
	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 0, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathWidthHeaderRedefine() {
	conf.EnableClientHints = true

	req := s.getRequest("/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 150, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 100, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeaderDisabled() {
	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 0, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathViewportWidthHeaderRedefine() {
	conf.EnableClientHints = true

	req := s.getRequest("/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("Viewport-Width", "100")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 150, po.Width)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeader() {
	conf.EnableClientHints = true

	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
	req.Header.Set("DPR", "2")
	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), 2.0, po.Dpr)
}

func (s *ProcessingOptionsTestSuite) TestParsePathDprHeaderDisabled() {
	req := s.getRequest("/unsafe/plain/http://images.dev/lorem/ipsum.jpg@png")
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

	req := s.getRequest("/HcvNognEV1bW6f8zRqxNYuOkV0IUf1xloRb57CzbT4g/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	_, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)
}

func (s *ProcessingOptionsTestSuite) TestParsePathSignedInvalid() {
	conf.Keys = []securityKey{securityKey("test-key")}
	conf.Salts = []securityKey{securityKey("test-salt")}
	conf.AllowInsecure = false

	req := s.getRequest("/unsafe/width:150/plain/http://images.dev/lorem/ipsum.jpg@png")
	_, err := parsePath(context.Background(), req)

	require.Error(s.T(), err)
	assert.Equal(s.T(), errInvalidSignature.Error(), err.Error())
}

func (s *ProcessingOptionsTestSuite) TestParsePathOnlyPresets() {
	conf.OnlyPresets = true
	conf.Presets["test1"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
	}
	conf.Presets["test2"] = urlOptions{
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	req := s.getRequest("/unsafe/test1:test2/plain/http://images.dev/lorem/ipsum.jpg")

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}

func (s *ProcessingOptionsTestSuite) TestParseBase64URLOnlyPresets() {
	conf.OnlyPresets = true
	conf.Presets["test1"] = urlOptions{
		urlOption{Name: "blur", Args: []string{"0.2"}},
	}
	conf.Presets["test2"] = urlOptions{
		urlOption{Name: "quality", Args: []string{"50"}},
	}

	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req := s.getRequest(fmt.Sprintf("/unsafe/test1:test2/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))))

	ctx, err := parsePath(context.Background(), req)

	require.Nil(s.T(), err)

	po := getProcessingOptions(ctx)
	assert.Equal(s.T(), float32(0.2), po.Blur)
	assert.Equal(s.T(), 50, po.Quality)
}
func TestProcessingOptions(t *testing.T) {
	suite.Run(t, new(ProcessingOptionsTestSuite))
}
