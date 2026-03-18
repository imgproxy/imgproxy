package imagetype_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

func TestDefaultTypesRegistered(t *testing.T) {
	// Test that all default types are properly registered by init()
	defaultTypes := []imagetype.Type{
		imagetype.JPEG, imagetype.JXL, imagetype.PNG, imagetype.WEBP, imagetype.GIF,
		imagetype.ICO, imagetype.SVG, imagetype.HEIC, imagetype.AVIF, imagetype.BMP, imagetype.TIFF,
	}

	for _, typ := range defaultTypes {
		t.Run(typ.String(), func(t *testing.T) {
			desc := imagetype.GetTypeDesc(typ)
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
		{"RAW", "../testdata/test-images/raw/RAW_CANON_1DM2.CR2", imagetype.Unknown}, // RAW is not supported
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.file)
			require.NoError(t, err)
			defer f.Close()

			got, err := imagetype.Detect(f, "", path.Ext(tt.file))
			if tt.want == imagetype.Unknown {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
