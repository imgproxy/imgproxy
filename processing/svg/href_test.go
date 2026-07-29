package svg_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/processing/svg"
	"github.com/stretchr/testify/require"
)

func TestIsSafeHref(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		schemes  map[string]struct{}
		expected bool
	}{
		{"empty", "", svg.HrefSchemes, true},
		{"fragment", "#test", svg.HrefSchemes, true},
		{"relative path", "foo/bar.svg", svg.HrefSchemes, true},
		{"absolute path relative url", "/foo/bar.svg", svg.HrefSchemes, true},
		{"protocol-relative", "//evil.com/x", svg.HrefSchemes, true},
		{"https allowed", "https://example.com", svg.HrefSchemes, true},
		{"http allowed", "http://example.com", svg.HrefSchemes, true},
		{"javascript rejected", "javascript:alert(1)", svg.HrefSchemes, false},
		{"javascript uppercase", "JavaScript:alert(1)", svg.HrefSchemes, false},
		{"javascript mixed case", "jAvAsCriPt:alert(1)", svg.HrefSchemes, false},
		{"leading newline", "\njavascript:alert(1)", svg.HrefSchemes, false},
		{"vbscript rejected", "vbscript:msgbox(1)", svg.HrefSchemes, false},
		{"data rejected for generic", "data:text/html,<script>alert(1)</script>", svg.HrefSchemes, false},
		{"mailto rejected", "mailto:test@example.com", svg.HrefSchemes, false},
		{"ftp rejected", "ftp://example.com/file", svg.HrefSchemes, false},
		{"numeric entity scheme", "jav&#97;script:alert(1)", svg.HrefSchemes, false},
		{"hex numeric entity scheme", "jav&#x61;script:alert(1)", svg.HrefSchemes, false},
		{"named entity colon", "javascript&colon;alert(1)", svg.HrefSchemes, false},
		{"data allowed for image", "data:image/png;base64,AAAA", svg.ImageHrefSchemes, true},
		{"javascript still rejected for image", "javascript:alert(1)", svg.ImageHrefSchemes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, svg.IsSafeHref(tt.value, tt.schemes))
		})
	}
}
