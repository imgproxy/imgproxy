package svg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/processing/svg"
)

func TestHasUnsafeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty", "", false},
		{"path", "M0 0 L10 10 Z", false},
		{"color", "#ff00ff", false},
		{"base64", "data:image/png;base64,iVBORw0KGgo=", false},
		{"lone lt", "a < b", false},
		{"lone gt", "a > b", false},
		{"trailing close", "abc</", false},
		{"safe end tag", "</g>", false},
		{"comment close", "x-->y", true},
		{"bang comment close", "x--!>y", true},
		{"cdata close", "x]>y", true},
		{"script end", "</script>", true},
		{"style end upper", "</STYLE", true},
		{"title end", "a</title>", true},
		{"after safe gt", "a>b</script", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, svg.HasUnsafeValue(tt.value))
		})
	}
}
