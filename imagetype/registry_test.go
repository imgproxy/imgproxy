package imagetype

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/stretchr/testify/require"
)

func TestRegisterType(t *testing.T) {
	// Create a separate registry for testing to avoid conflicts with global registry
	testRegistry := NewRegistry()

	// Register a custom type
	customDesc := &TypeDesc{
		String:                "custom",
		Ext:                   ".custom",
		Mime:                  "image/custom",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
	}

	customType := testRegistry.RegisterType(customDesc)

	// Verify the type is now registered
	result := testRegistry.GetTypeDesc(customType)
	require.NotNil(t, result)
	require.Equal(t, customDesc.String, result.String)
	require.Equal(t, customDesc.Ext, result.Ext)
	require.Equal(t, customDesc.Mime, result.Mime)
	require.Equal(t, customDesc.IsVector, result.IsVector)
	require.Equal(t, customDesc.SupportsAlpha, result.SupportsAlpha)
	require.Equal(t, customDesc.SupportsColourProfile, result.SupportsColourProfile)
}

func TestTypeProperties(t *testing.T) {
	// Test that Type methods use TypeDesc fields correctly
	tests := []struct {
		name                string
		typ                 Type
		expectVector        bool
		expectAlpha         bool
		expectColourProfile bool
		expectQuality       bool
		expectAnimationLoad bool
		expectAnimationSave bool
		expectThumbnail     bool
	}{
		{
			name:                "JPEG",
			typ:                 JPEG,
			expectVector:        false,
			expectAlpha:         false,
			expectColourProfile: true,
			expectQuality:       true,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     false,
		},
		{
			name:                "PNG",
			typ:                 PNG,
			expectVector:        false,
			expectAlpha:         true,
			expectColourProfile: true,
			expectQuality:       false,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     false,
		},
		{
			name:                "WEBP",
			typ:                 WEBP,
			expectVector:        false,
			expectAlpha:         true,
			expectColourProfile: true,
			expectQuality:       true,
			expectAnimationLoad: true,
			expectAnimationSave: true,
			expectThumbnail:     false,
		},
		{
			name:                "SVG",
			typ:                 SVG,
			expectVector:        true,
			expectAlpha:         true,
			expectColourProfile: false,
			expectQuality:       false,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     false,
		},
		{
			name:                "GIF",
			typ:                 GIF,
			expectVector:        false,
			expectAlpha:         true,
			expectColourProfile: false,
			expectQuality:       false,
			expectAnimationLoad: true,
			expectAnimationSave: true,
			expectThumbnail:     false,
		},
		{
			name:                "HEIC",
			typ:                 HEIC,
			expectVector:        false,
			expectAlpha:         true,
			expectColourProfile: true,
			expectQuality:       true,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     true,
		},
		{
			name:                "AVIF",
			typ:                 AVIF,
			expectVector:        false,
			expectAlpha:         true,
			expectColourProfile: true,
			expectQuality:       true,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectVector, tt.typ.IsVector())
			require.Equal(t, tt.expectAlpha, tt.typ.SupportsAlpha())
			require.Equal(t, tt.expectColourProfile, tt.typ.SupportsColourProfile())
			require.Equal(t, tt.expectQuality, tt.typ.SupportsQuality())
			require.Equal(t, tt.expectAnimationLoad, tt.typ.SupportsAnimationLoad())
			require.Equal(t, tt.expectAnimationSave, tt.typ.SupportsAnimationSave())
			require.Equal(t, tt.expectThumbnail, tt.typ.SupportsThumbnail())
		})
	}
}

func TestRegisterDetector(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := NewRegistry()

	// Create a test detector function
	testDetector := func(r bufreader.ReadPeeker) (Type, error) {
		b, err := r.Peek(2)
		if err != nil {
			return Unknown, err
		}
		if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xD8 {
			return JPEG, nil
		}
		return Unknown, newUnknownFormatError()
	}

	// Register the detector using the method
	testRegistry.RegisterDetector(testDetector)

	// Verify the detector is registered
	require.Len(t, testRegistry.detectors, 1)
	require.NotNil(t, testRegistry.detectors[0])
}

func TestRegisterMagicBytes(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := NewRegistry()

	require.Empty(t, testRegistry.detectors)

	// Register magic bytes for JPEG using the method
	jpegMagic := []byte{0xFF, 0xD8}
	testRegistry.RegisterMagicBytes(JPEG, jpegMagic)

	// Verify the magic bytes are registered
	require.Len(t, testRegistry.detectors, 1)
}
