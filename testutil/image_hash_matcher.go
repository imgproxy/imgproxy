package testutil

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include <vips/vips.h>
#include "image_hash_matcher.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"unsafe"

	"github.com/corona10/goimagehash"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/require"
)

const (
	// hashPath is a path to hash data in testdata folder
	hashPath = "test-hashes"

	// If TEST_CREATE_MISSING_HASHES is set, matcher would create missing hash files
	createMissingHashesEnv = "TEST_CREATE_MISSING_HASHES"

	// If this is set, the images are saved to this folder before hash is calculated
	saveTmpImagesPathEnv = "TEST_SAVE_TMP_IMAGES_PATH"
)

// ImageHashMatcher is a helper struct for image hash comparison in tests
type ImageHashMatcher struct {
	hashesPath          string
	createMissingHashes bool
	saveTmpImagesPath   string
}

// NewImageHashMatcher creates a new ImageHashMatcher instance
func NewImageHashMatcher(testDataProvider *TestDataProvider) *ImageHashMatcher {
	hashesPath := testDataProvider.Path(hashPath)
	createMissingHashes := len(os.Getenv(createMissingHashesEnv)) > 0
	saveTmpImagesPath := os.Getenv(saveTmpImagesPathEnv)

	return &ImageHashMatcher{
		hashesPath:          hashesPath,
		createMissingHashes: createMissingHashes,
		saveTmpImagesPath:   saveTmpImagesPath,
	}
}

// ImageHashMatches is a testing helper, which accepts image as reader, calculates
// difference hash and compares it with a hash saved to testdata/test-hashes
// folder.
func (m *ImageHashMatcher) ImageMatches(t *testing.T, img io.Reader, key string, maxDistance int) {
	t.Helper()

	// Read image in memory
	buf, err := io.ReadAll(img)
	require.NoError(t, err)

	// Save tmp image if requested
	m.saveTmpImage(t, key, buf)

	// Convert to RGBA and read into memory using VIPS
	var data unsafe.Pointer
	var size C.size_t
	var width, height C.int

	// no one knows why this triggers linter
	//nolint:gocritic
	readErr := C.vips_image_read_from_to_memory(unsafe.Pointer(unsafe.SliceData(buf)), C.size_t(len(buf)), &data, &size, &width, &height)
	if readErr != 0 {
		t.Fatalf("failed to read image from memory, key: %s, error: %s", key, vipsErrorMessage())
	}

	defer C.vips_memory_buffer_free(data)

	// Convert raw RGBA pixel data to Go image.Image
	goImg, err := createRGBAFromRGBAPixels(int(width), int(height), data, size)
	require.NoError(t, err)

	sourceHash, err := goimagehash.DifferenceHash(goImg)
	require.NoError(t, err)

	// Calculate image hash path (create folder if missing)
	hashPath, err := m.makeTargetPath(t, m.hashesPath, t.Name(), key, "hash")
	require.NoError(t, err)

	// Try to read or create the hash file
	f, err := os.Open(hashPath)
	if os.IsNotExist(err) {
		// If the hash file does not exist, and we are not allowed to create it, fail
		if !m.createMissingHashes {
			require.NoError(t, err, "failed to read target hash from %s, use TEST_CREATE_MISSING_HASHES=true to create it, TEST_SAVE_TMP_IMAGES_PATH=/some/path to check resulting images", hashPath)
		}

		// Create missing hash file
		h, hashErr := os.Create(hashPath)
		require.NoError(t, hashErr, "failed to create target hash file %s", hashPath)
		defer h.Close()

		// Dump calculated source hash to this hash file
		hashErr = sourceHash.Dump(h)
		require.NoError(t, hashErr, "failed to write target hash to %s", hashPath)

		t.Logf("Created missing hash in %s", hashPath)
		return
	}

	// Otherwise, if there is no error or error is something else
	require.NoError(t, err)

	// Load image hash from hash file
	targetHash, err := goimagehash.LoadImageHash(f)
	require.NoError(t, err, "failed to load target hash from %s", hashPath)

	// Ensure distance is OK
	distance, err := sourceHash.Distance(targetHash)
	require.NoError(t, err, "failed to calculate hash distance for %s", key)

	require.LessOrEqual(t, distance, maxDistance, "image hashes are too different for %s: distance %d", key, distance)
}

// makeTargetPath creates the target directory and returns file path for saving
// the image or hash.
func (m *ImageHashMatcher) makeTargetPath(t *testing.T, base, folder, filename, ext string) (string, error) {
	// Create the target directory if it doesn't exist
	targetDir := path.Join(base, folder)
	err := os.MkdirAll(targetDir, 0755)
	require.NoError(t, err, "failed to create %s target directory", targetDir)

	// Replace the extension with the detected one
	filename = filename + "." + ext

	// Create the target file
	targetPath := path.Join(targetDir, filename)

	return targetPath, nil
}

// saveTmpImage saves the provided image data to a temporary file
func (m *ImageHashMatcher) saveTmpImage(t *testing.T, key string, buf []byte) {
	if m.saveTmpImagesPath == "" {
		return
	}

	// Detect the image type to get the correct extension
	ext, err := imagetype.Detect(bytes.NewReader(buf), "", "")
	require.NoError(t, err)

	targetPath, err := m.makeTargetPath(t, m.saveTmpImagesPath, t.Name(), key, ext.String())
	require.NoError(t, err, "failed to create TEST_SAVE_TMP_IMAGES target path for %s/%s", t.Name(), key)

	targetFile, err := os.Create(targetPath)
	require.NoError(t, err, "failed to create TEST_SAVE_TMP_IMAGES target file %s", targetPath)
	defer targetFile.Close()

	// Write the image data to the file
	_, err = io.Copy(targetFile, bytes.NewReader(buf))
	require.NoError(t, err, "failed to write to TEST_SAVE_TMP_IMAGES target file %s", targetPath)

	t.Logf("Saved temporary image to %s", targetPath)
}

// createRGBAFromRGBAPixels creates a Go image.Image from raw RGBA VIPS pixel data
func createRGBAFromRGBAPixels(width, height int, data unsafe.Pointer, size C.size_t) (*image.RGBA, error) {
	// RGBA should have 4 bands
	expectedSize := width * height * 4
	if int(size) != expectedSize {
		return nil, fmt.Errorf("size mismatch: expected %d bytes for RGBA, got %d", expectedSize, int(size))
	}

	pixels := unsafe.Slice((*byte)(data), int(size))

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy RGBA pixel data directly
	copy(img.Pix, pixels)

	return img, nil
}

// vipsErrorMessage reads VIPS error message
func vipsErrorMessage() string {
	defer C.vips_error_clear()
	return strings.TrimSpace(C.GoString(C.vips_error_buffer()))
}
