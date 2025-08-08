package imagetype_new

import (
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
			desc := GetType(typ)
			require.NotNil(t, desc)

			// Verify that the description has non-empty fields
			require.NotEmpty(t, desc.String)
			require.NotEmpty(t, desc.Ext)
			require.NotEqual(t, "application/octet-stream", desc.Mime)
		})
	}
}
