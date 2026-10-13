package transport_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/fetcher/transport"
)

func TestSourcePath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		sep  string
		want string
	}{
		{
			name: "http with signed query",
			url:  "http://origin/img.png?v=1.2&X-Amz-Signature=secret",
			sep:  "?",
			want: "/img.png",
		},
		{
			name: "https with fragment and escaped chars",
			url:  "https://h/a%20b.jpg#frag",
			sep:  "?",
			want: "/a%20b.jpg",
		},
		{
			name: "s3 key with percent",
			url:  "s3://bucket/50%off.jpg",
			sep:  "?",
			want: "50%off.jpg",
		},
		{
			name: "local key with hash",
			url:  "local:///photo#1.png",
			sep:  "?",
			want: "photo#1.png",
		},
		{
			name: "s3 key with query",
			url:  "s3://bucket/dir/img.png?X-Amz=secret",
			sep:  "?",
			want: "dir/img.png",
		},
		{
			name: "s3 key with custom separator",
			url:  "s3://bucket/dir/img.png@version",
			sep:  "@",
			want: "dir/img.png",
		},
		{
			name: "http path with custom separator is not cut",
			url:  "http://origin/dir/img.png@version",
			sep:  "@",
			want: "/dir/img.png@version",
		},
		{
			name: "unparsable url",
			url:  "http://[::1",
			sep:  "?",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, transport.SourcePath(tt.url, tt.sep))
		})
	}
}
