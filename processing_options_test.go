package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodePlainUrl(t *testing.T) {
	decodeStr := func(s string) []string {
		path, ext, err := decodePlainURL(strings.Split(s, "/"))
		assert.Nil(t, err)
		return []string{path, ext}
	}
	assert.Equal(t, []string{"path/to/file.jpg", "jpg"}, decodeStr("jpg/path/to/file.jpg"))
	assert.Equal(t, []string{"path/to/file.png", "jpg"}, decodeStr("png/path/to/file.jpg"))
	assert.Equal(t, []string{"path/to/file.jpg", "png"}, decodeStr("jpg/path/to/file.png"))
	assert.Equal(t, []string{"path/to/file", "jpg"}, decodeStr("/path/to/file.jpg"))
	assert.Equal(t, []string{"path/to/file.", "jpg"}, decodeStr("./path/to/file.jpg"))
}
