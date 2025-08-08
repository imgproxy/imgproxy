package imagedetect

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype_new"
	"github.com/stretchr/testify/require"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		file string
		want imagetype_new.Type
	}{
		{"JPEG", "../testdata/test-images/jpg/jpg.jpg", imagetype_new.JPEG},
		{"JXL", "../testdata/test-images/jxl/jxl.jxl", imagetype_new.JXL},
		{"PNG", "../testdata/test-images/png/png.png", imagetype_new.PNG},
		{"WEBP", "../testdata/test-images/webp/webp.webp", imagetype_new.WEBP},
		{"GIF", "../testdata/test-images/gif/gif.gif", imagetype_new.GIF},
		{"ICO", "../testdata/test-images/ico/png-256x256.ico", imagetype_new.ICO},
		{"SVG", "../testdata/test-images/svg/svg.svg", imagetype_new.SVG},
		{"HEIC", "../testdata/test-images/heif/heif.heif", imagetype_new.HEIC},
		{"BMP", "../testdata/test-images/bmp/24-bpp.bmp", imagetype_new.BMP},
		{"TIFF", "../testdata/test-images/tiff/tiff.tiff", imagetype_new.TIFF},
		{"SVG", "../testdata/test-images/svg/svg.svg", imagetype_new.SVG},
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
