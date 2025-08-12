package imagedetect

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
)

func TestRegisterDetector(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := &Registry{}

	// Create a test detector function
	testDetector := func(r bufreader.ReadPeeker) (imagetype.Type, error) {
		b, err := r.Peek(2)
		if err != nil {
			return imagetype.Unknown, err
		}
		if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xD8 {
			return imagetype.JPEG, nil
		}
		return imagetype.Unknown, newUnknownFormatError()
	}

	// Register the detector using the method
	testRegistry.RegisterDetector(testDetector, 64)

	// Verify the detector is registered
	require.Len(t, testRegistry.detectors, 1)
	require.NotNil(t, testRegistry.detectors[0])
}

func TestRegisterMagicBytes(t *testing.T) {
	// Create a test registry to avoid interfering with global state
	testRegistry := &Registry{}

	require.Empty(t, testRegistry.detectors)

	// Register magic bytes for JPEG using the method
	jpegMagic := []byte{0xFF, 0xD8}
	testRegistry.RegisterMagicBytes(jpegMagic, imagetype.JPEG)

	// Verify the magic bytes are registered
	require.Len(t, testRegistry.detectors, 1)
}
