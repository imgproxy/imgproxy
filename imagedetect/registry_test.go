package imagedetect

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/imagetype_new"
	"github.com/stretchr/testify/require"
)

func TestRegisterDetector(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := &Registry{}

	// Create a test detector function
	testDetector := func(data []byte) (imagetype_new.Type, error) {
		if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
			return imagetype_new.JPEG, nil
		}
		return imagetype_new.Unknown, newUnknownFormatError()
	}

	// Register the detector using the method
	testRegistry.RegisterDetector(testDetector, 64)

	// Verify the detector is registered
	require.Len(t, testRegistry.detectors, 1)
	require.Equal(t, 64, testRegistry.detectors[0].BytesNeeded)
	require.NotNil(t, testRegistry.detectors[0].Func)
}

func TestRegisterMagicBytes(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := &Registry{}

	// Register magic bytes for JPEG using the method
	jpegMagic := []byte{0xFF, 0xD8}
	testRegistry.RegisterMagicBytes(jpegMagic, imagetype_new.JPEG)

	// Verify the magic bytes are registered
	require.Len(t, testRegistry.magicBytes, 1)
	require.Equal(t, jpegMagic, testRegistry.magicBytes[0].Signature)
	require.Equal(t, imagetype_new.JPEG, testRegistry.magicBytes[0].Type)
}
