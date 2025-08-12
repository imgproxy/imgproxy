package imagedetect

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		file string
		want imagetype.Type
	}{
		{"JPEG", "../testdata/test-images/jpg/jpg.jpg", imagetype.JPEG},
		{"JXL", "../testdata/test-images/jxl/jxl.jxl", imagetype.JXL},
		{"PNG", "../testdata/test-images/png/png.png", imagetype.PNG},
		{"WEBP", "../testdata/test-images/webp/webp.webp", imagetype.WEBP},
		{"GIF", "../testdata/test-images/gif/gif.gif", imagetype.GIF},
		{"ICO", "../testdata/test-images/ico/png-256x256.ico", imagetype.ICO},
		{"SVG", "../testdata/test-images/svg/svg.svg", imagetype.SVG},
		{"HEIC", "../testdata/test-images/heif/heif.heif", imagetype.HEIC},
		{"BMP", "../testdata/test-images/bmp/24-bpp.bmp", imagetype.BMP},
		{"TIFF", "../testdata/test-images/tiff/tiff.tiff", imagetype.TIFF},
		{"SVG", "../testdata/test-images/svg/svg.svg", imagetype.SVG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
			require.NoError(t, err)
			defer f.Close()

			got, err := Detect(f)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
