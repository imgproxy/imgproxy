package httpheaders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentDispositionValue(t *testing.T) {
	// Test cases for ContentDispositionValue function that generates content-disposition headers
	tests := []struct {
		name             string
		url              string
		filename         string
		ext              string
		returnAttachment bool
		expected         string
		contentType      string
	}{
		{
			name:             "BasicURL",
			url:              "http://example.com/test.jpg",
			filename:         "",
			ext:              "",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"test.jpg\"",
		},
		{
			name:             "EmptyFilename",
			url:              "http://example.com/path/to/",
			filename:         "",
			ext:              "",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"image\"",
		},
		{
			name:             "EmptyFilenameWithExt",
			url:              "http://example.com/path/to/",
			filename:         "",
			ext:              ".png",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"image.png\"",
		},
		{
			name:             "EmptyFilenameWithFilenameAndExt",
			url:              "http://example.com/path/to/",
			filename:         "example",
			ext:              ".png",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"example.png\"",
		},
		{
			name:             "EmptyFilenameWithFilenameOverride",
			url:              "http://example.com/path/to/",
			filename:         "example",
			ext:              ".jpg",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"example.jpg\"",
		},
		{
			name:             "PresentFilenameWithExtOverride",
			url:              "http://example.com/path/to/face.png",
			filename:         "",
			ext:              ".jpg",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"face.jpg\"",
		},
		{
			name:             "PresentFilenameWithFilenameOverride",
			url:              "http://example.com/path/to/123.png",
			filename:         "face",
			ext:              ".jpg",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"face.jpg\"",
		},
		{
			name:             "EmptyFilenameWithFallback",
			url:              "http://example.com/path/to/",
			filename:         "",
			ext:              ".png",
			contentType:      "",
			returnAttachment: false,
			expected:         "inline; filename=\"image.png\"",
		},
		{
			name:             "AttachmentDisposition",
			url:              "http://example.com/test.jpg",
			filename:         "",
			ext:              "",
			contentType:      "",
			returnAttachment: true,
			expected:         "attachment; filename=\"test.jpg\"",
		},
		{
			name:             "FilenameWithQuotes",
			url:              "http://example.com/test.jpg",
			filename:         "my\"file",
			ext:              ".png",
			returnAttachment: false,
			contentType:      "",
			expected:         "inline; filename=\"my%22file.png\"",
		},
		{
			name:             "FilenameWithContentType",
			url:              "http://example.com/test",
			filename:         "my\"file",
			ext:              "",
			contentType:      "image/png",
			returnAttachment: false,
			expected:         "inline; filename=\"my%22file.png\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ContentDispositionValue(tc.url, tc.filename, tc.ext, tc.contentType, tc.returnAttachment)
			require.Equal(t, tc.expected, result)
		})
	}
}
