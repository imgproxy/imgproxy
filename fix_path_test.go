package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/signature/width:100/height:100/plain/http:/example.com/image.jpg",
			expected: "/signature/width:100/height:100/plain/http://example.com/image.jpg",
		},
		{
			input:    "/signature/width:100/height:100/plain/local:/image.jpg",
			expected: "/signature/width:100/height:100/plain/local:///image.jpg",
		},
		{
			input:    "/signature/width%3A100/height%3A100/plain/local:/image.jpg",
			expected: "/signature/width:100/height:100/plain/local:///image.jpg",
		},
		{
			input:    "/signature/width%3A100/height%3A100/abc/abc",
			expected: "/signature/width:100/height:100/abc/abc",
		},
		{
			input:    "/signature/width%3A100/height%3A100/plain/",
			expected: "/signature/width:100/height:100/plain/",
		},
	}

	for _, test := range tests {
		actual := fixPath(test.input)
		require.Equal(
			t, test.expected, actual,
			"fixPath(%q) = %q; expected %q",
			test.input, actual, test.expected,
		)
	}
}
