package imagetype_new

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterType(t *testing.T) {
	// Create a separate registry for testing to avoid conflicts with global registry
	testRegistry := &Registry{}

	// Use a high type number to avoid conflicts with default types
	const customType Type = 1000

	// First, verify the type is not registered
	desc := testRegistry.GetType(customType)
	require.Nil(t, desc)

	// Register a custom type
	customDesc := &TypeDesc{
		String:                "custom",
		Ext:                   ".custom",
		Mime:                  "image/custom",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
	}

	testRegistry.RegisterType(customType, customDesc)

	// Verify the type is now registered
	result := testRegistry.GetType(customType)
	require.NotNil(t, result)
	require.Equal(t, customDesc.String, result.String)
	require.Equal(t, customDesc.Ext, result.Ext)
	require.Equal(t, customDesc.Mime, result.Mime)
	require.Equal(t, customDesc.IsVector, result.IsVector)
	require.Equal(t, customDesc.SupportsAlpha, result.SupportsAlpha)
	require.Equal(t, customDesc.SupportsColourProfile, result.SupportsColourProfile)
}

func TestRegisterTypeError(t *testing.T) {
	// Create a separate registry for testing to avoid conflicts with global registry
	testRegistry := &Registry{}

	// Use a high type number to avoid conflicts
	const testType Type = 2000

	desc1 := &TypeDesc{
		String:                "test1",
		Ext:                   ".test1",
		Mime:                  "image/test1",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
	}

	desc2 := &TypeDesc{
		String:                "test2",
		Ext:                   ".test2",
		Mime:                  "image/test2",
		IsVector:              true,
		SupportsAlpha:         false,
		SupportsColourProfile: false,
	}

	// Register the first type
	testRegistry.RegisterType(testType, desc1)

	// Attempting to register the same type again should return an error
	require.Error(t, testRegistry.RegisterType(testType, desc2))
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
		{
			name:                "Unknown",
			typ:                 Unknown,
			expectVector:        false,
			expectAlpha:         false,
			expectColourProfile: false,
			expectQuality:       false,
			expectAnimationLoad: false,
			expectAnimationSave: false,
			expectThumbnail:     false,
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
