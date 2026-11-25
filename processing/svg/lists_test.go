package svg_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/processing/svg"
)

func TestApplyListOverrides(t *testing.T) {
	defaults := []string{"svg", "g", "circle", "rect"}

	tests := []struct {
		name     string
		entries  []string
		expected []string
		invalid  bool
	}{
		{"nil entries", nil, []string{"circle", "g", "rect", "svg"}, false},
		{"empty entries skipped", []string{"", " ", ""}, []string{"circle", "g", "rect", "svg"}, false},
		{"add", []string{"+path"}, []string{"circle", "g", "path", "rect", "svg"}, false},
		{"remove", []string{"-circle"}, []string{"g", "rect", "svg"}, false},
		{"remove unknown", []string{"-nope"}, []string{"circle", "g", "rect", "svg"}, false},
		{"replace", []string{"a", "b"}, []string{"a", "b"}, false},
		{"replace and modify", []string{"a", "b", "+c", "-a"}, []string{"b", "c"}, false},
		{"bare after plus", []string{"+x", "a"}, []string{"a", "x"}, false},
		{"bare between modifiers", []string{"+x", "a", "-x"}, []string{"a"}, false},
		{"remove then add", []string{"-circle", "+circle"}, []string{"circle", "g", "rect", "svg"}, false},
		{"add then remove", []string{"+circle", "-circle"}, []string{"g", "rect", "svg"}, false},
		{"case insensitive", []string{"+LinearGradient", "-RECT"}, []string{"circle", "g", "lineargradient", "svg"}, false},
		{"whitespace", []string{" + path ", " - g "}, []string{"circle", "path", "rect", "svg"}, false},
		{"bare plus", []string{"+", "+path"}, nil, true},
		{"bare minus", []string{"-", "-g"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := svg.ApplyListOverrides(defaults, tt.entries)

			if tt.invalid {
				require.Error(t, err)
				require.Nil(t, set)

				return
			}

			require.NoError(t, err)

			actual := make([]string, 0, len(set))
			for name := range set {
				actual = append(actual, name)
			}

			sort.Strings(actual)

			require.Equal(t, tt.expected, actual)
		})
	}
}
