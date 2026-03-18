package imagetype_test

import (
	"bytes"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
)

func TestRegisterType(t *testing.T) {
	// Create a separate registry for testing to avoid conflicts with global registry
	testRegistry := imagetype.NewRegistry()

	// Register a custom type
	customDesc := &imagetype.TypeDesc{
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
		typ                 imagetype.Type
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
			typ:                 imagetype.JPEG,
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
			typ:                 imagetype.PNG,
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
			typ:                 imagetype.WEBP,
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
			typ:                 imagetype.SVG,
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
			typ:                 imagetype.GIF,
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
			typ:                 imagetype.HEIC,
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
			typ:                 imagetype.AVIF,
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
	testRegistry := imagetype.NewRegistry()

	functionsEqual := func(fn1, fn2 imagetype.DetectFunc) {
		// Compare function names to check if they are the same
		fnName1 := runtime.FuncForPC(reflect.ValueOf(fn1).Pointer()).Name()
		fnName2 := runtime.FuncForPC(reflect.ValueOf(fn2).Pointer()).Name()
		require.Equal(t, fnName1, fnName2)
	}

	// Create a test detector functions
	testDetector1 := func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) { return imagetype.JPEG, nil }
	testDetector2 := func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) { return imagetype.PNG, nil }
	testDetector3 := func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) { return imagetype.GIF, nil }
	testDetector4 := func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) { return imagetype.SVG, nil }

	// Register the detectors using the method
	testRegistry.RegisterDetector(0, testDetector1)
	testRegistry.RegisterDetector(0, testDetector2)
	testRegistry.RegisterDetector(10, testDetector3)
	testRegistry.RegisterDetector(5, testDetector4)

	// Verify the detectors are registered
	detectors := testRegistry.Detectors()
	require.Len(t, detectors, 4)

	// Verify the order of detectors based on priority
	require.Equal(t, 0, detectors[0].Priority)
	functionsEqual(testDetector1, detectors[0].Fn)
	require.Equal(t, 0, detectors[1].Priority)
	functionsEqual(testDetector2, detectors[1].Fn)
	require.Equal(t, 5, detectors[2].Priority)
	functionsEqual(testDetector4, detectors[2].Fn)
	require.Equal(t, 10, detectors[3].Priority)
	functionsEqual(testDetector3, detectors[3].Fn)
}

func TestRegisterMagicBytes(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := imagetype.NewRegistry()

	require.Empty(t, testRegistry.Detectors())

	// Register magic bytes for JPEG using the method
	jpegMagic := []byte{0xFF, 0xD8}
	testRegistry.RegisterMagicBytes(imagetype.JPEG, jpegMagic)

	// Verify the magic bytes are registered
	detectors := testRegistry.Detectors()
	require.Len(t, detectors, 1)
	require.Equal(t, -1, detectors[0].Priority)

	typ, err := testRegistry.Detect(bufreader.New(bytes.NewReader(jpegMagic)), "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.JPEG, typ)
}

func TestDetectionErrorReturns(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := imagetype.NewRegistry()

	detErr := error(nil)

	// The first detector will fail with detErr
	testRegistry.RegisterDetector(0, func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) {
		return imagetype.Unknown, detErr
	})

	// The second detector will succeed
	testRegistry.RegisterDetector(1, func(r bufreader.ReadPeeker, _, _ string) (imagetype.Type, error) {
		return imagetype.JPEG, nil
	})

	// We don't actually need to read anything, just create a reader
	r := strings.NewReader("")

	// Should not fail with io.EOF
	detErr = io.EOF
	typ, err := testRegistry.Detect(r, "", "")
	require.Equal(t, imagetype.JPEG, typ)
	require.NoError(t, err)

	// Should not fail with io.ErrUnexpectedEOF
	detErr = io.ErrUnexpectedEOF
	typ, err = testRegistry.Detect(r, "", "")
	require.Equal(t, imagetype.JPEG, typ)
	require.NoError(t, err)

	// Should fail with other read errors
	detErr = io.ErrClosedPipe
	typ, err = testRegistry.Detect(r, "", "")
	require.Equal(t, imagetype.Unknown, typ)
	require.Error(t, err)
	require.ErrorAs(t, err, &imagetype.TypeDetectionError{})
	require.ErrorIs(t, err, io.ErrClosedPipe)
}
