package imagetype

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultTypesRegistered(t *testing.T) {
	// Test that all default types are properly registered by init()
	defaultTypes := []Type{
		JPEG, JXL, PNG, WEBP, GIF, ICO, SVG, HEIC, AVIF, BMP, TIFF,
	}

	for _, typ := range defaultTypes {
		t.Run(typ.String(), func(t *testing.T) {
			desc := GetTypeDesc(typ)
			require.NotNil(t, desc)

			// Verify that the description has non-empty fields
			require.NotEmpty(t, desc.String)
			require.NotEmpty(t, desc.Ext)
			require.NotEqual(t, "application/octet-stream", desc.Mime)
		})
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		file string
		want Type
	}{
		{"JPEG", "../testdata/test-images/jpg/jpg.jpg", JPEG},
		{"JXL", "../testdata/test-images/jxl/jxl.jxl", JXL},
		{"PNG", "../testdata/test-images/png/png.png", PNG},
		{"WEBP", "../testdata/test-images/webp/webp.webp", WEBP},
		{"GIF", "../testdata/test-images/gif/gif.gif", GIF},
		{"ICO", "../testdata/test-images/ico/png-256x256.ico", ICO},
		{"SVG", "../testdata/test-images/svg/svg.svg", SVG},
		{"HEIC", "../testdata/test-images/heif/heif.heif", HEIC},
		{"BMP", "../testdata/test-images/bmp/24-bpp.bmp", BMP},
		{"TIFF", "../testdata/test-images/tiff/tiff.tiff", TIFF},
		{"SVG", "../testdata/test-images/svg/svg.svg", SVG},
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
