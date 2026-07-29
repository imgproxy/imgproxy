package storage_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/httpheaders"
	"github.com/imgproxy/imgproxy/v4/storage"
)

func TestObjectReaderContentLength(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   int64
	}{
		{name: "valid header", header: "12345", want: 12345},
		{name: "invalid header", header: "not-a-number", want: -1},
		{name: "no header", header: "", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := make(http.Header)
			if len(tt.header) > 0 {
				h.Set(httpheaders.ContentLength, tt.header)
			}

			r := storage.NewObjectOK(h, http.NoBody)

			require.Equal(t, tt.want, r.ContentLength())
		})
	}
}
