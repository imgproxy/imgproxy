package stemext

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStemExt(t *testing.T) {
	// Test cases for stem and ext detection
	tests := []struct {
		name string
		url  string
		stem string
		ext  string
		fn   func(*stemExt) (string, string)
	}{
		{
			name: "BasicURL",
			url:  "http://example.com/test.jpg",
			stem: "test",
			ext:  ".jpg",
			fn: func(se *stemExt) (string, string) {
				return se.StemExt()
			},
		},
		{
			name: "EmptyFilename",
			url:  "http://example.com/path/to/",
			stem: "",
			ext:  "",
			fn: func(se *stemExt) (string, string) {
				return se.StemExt()
			},
		},
		{
			name: "EmptyFilenameWithContentType",
			url:  "http://example.com/path/to/",
			stem: "",
			ext:  ".png",
			fn: func(se *stemExt) (string, string) {
				return se.SetExtFromContentTypeIfEmpty("image/png").StemExt()
			},
		},
		{
			name: "EmptyFilenameWithContentTypeAndOverride",
			url:  "http://example.com/path/to/",
			stem: "example",
			ext:  ".png",
			fn: func(se *stemExt) (string, string) {
				return se.OverrideStem("example").SetExtFromContentTypeIfEmpty("image/png").StemExt()
			},
		},
		{
			name: "EmptyFilenameWithOverride",
			url:  "http://example.com/path/to/",
			stem: "example",
			ext:  ".jpg",
			fn: func(se *stemExt) (string, string) {
				return se.OverrideStem("example").OverrideExt(".jpg").StemExt()
			},
		},
		{
			name: "PresentFilenameWithOverride",
			url:  "http://example.com/path/to/face",
			stem: "face",
			ext:  ".jpg",
			fn: func(se *stemExt) (string, string) {
				return se.OverrideExt(".jpg").StemExt()
			},
		},
		{
			name: "PresentFilenameWithOverride",
			url:  "http://example.com/path/to/123",
			stem: "face",
			ext:  ".jpg",
			fn: func(se *stemExt) (string, string) {
				return se.OverrideStem("face").OverrideExt(".jpg").StemExt()
			},
		},
		{
			name: "EmptyFilenameWithFallback",
			url:  "http://example.com/path/to/",
			stem: "image",
			ext:  ".png",
			fn: func(se *stemExt) (string, string) {
				return se.SetExtFromContentTypeIfEmpty("image/png").StemExtWithFallback()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			require.NoError(t, err)

			se := FromURL(u)
			stem, ext := tc.fn(se)

			require.Equal(t, tc.stem, stem)
			require.Equal(t, tc.ext, ext)
		})
	}
}
