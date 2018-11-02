package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBase64URL(t *testing.T) {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, imageURL, getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParseBase64URLWithoutExtension(t *testing.T) {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/%s", base64.RawURLEncoding.EncodeToString([]byte(imageURL))), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, imageURL, getImageURL(ctx))
		assert.Equal(t, imageTypeJPEG, getProcessingOptions(ctx).Format)
	}
}

func TestParseBase64URLWithBase(t *testing.T) {
	oldConf := conf
	defer func() { conf = oldConf }()

	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParseBase64URLInvalid(t *testing.T) {
	imageURL := "lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/%s.png", base64.RawURLEncoding.EncodeToString([]byte(imageURL))), nil)

	_, err := parsePath(context.Background(), req)

	assert.Equal(t, errInvalidImageURL, err)
}

func TestParsePlainURL(t *testing.T) {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, imageURL, getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParsePlainURLWithoutExtension(t *testing.T) {
	imageURL := "http://images.dev/lorem/ipsum.jpg"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s", imageURL), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, imageURL, getImageURL(ctx))
		assert.Equal(t, imageTypeJPEG, getProcessingOptions(ctx).Format)
	}
}
func TestParsePlainURLEscaped(t *testing.T) {
	imageURL := "http://images.dev/lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, imageURL, getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParsePlainURLWithBase(t *testing.T) {
	oldConf := conf
	defer func() { conf = oldConf }()

	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParsePlainURLEscapedWithBase(t *testing.T) {
	oldConf := conf
	defer func() { conf = oldConf }()

	conf.BaseURL = "http://images.dev/"

	imageURL := "lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)), nil)

	ctx, err := parsePath(context.Background(), req)

	if assert.Nil(t, err) {
		assert.Equal(t, fmt.Sprintf("%s%s", conf.BaseURL, imageURL), getImageURL(ctx))
		assert.Equal(t, imageTypePNG, getProcessingOptions(ctx).Format)
	}
}

func TestParsePlainURLInvalid(t *testing.T) {
	imageURL := "lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", imageURL), nil)

	_, err := parsePath(context.Background(), req)

	assert.Equal(t, errInvalidImageURL, err)
}

func TestParsePlainURLEscapedInvalid(t *testing.T) {
	imageURL := "lorem/ipsum.jpg?param=value"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://example.com/unsafe/size:100:100/plain/%s@png", url.PathEscape(imageURL)), nil)

	_, err := parsePath(context.Background(), req)

	assert.Equal(t, errInvalidImageURL, err)
}
