package testutil

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

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

// ImageHashCacheMatcher is a helper struct for image hash comparison in tests
type ImageHashCacheMatcher struct {
	hashesPath          string
	createMissingHashes bool
	saveTmpImagesPath   string
	hashType            ImageHashType
}

// NewImageHashCacheMatcher creates a new ImageHashRegressionMatcher instance
func NewImageHashCacheMatcher(testDataProvider *TestDataProvider, hashType ImageHashType) *ImageHashCacheMatcher {
	hashesPath := testDataProvider.Path(hashPath)
	createMissingHashes := len(os.Getenv(createMissingHashesEnv)) > 0
	saveTmpImagesPath := os.Getenv(saveTmpImagesPathEnv)

	return &ImageHashCacheMatcher{
		hashesPath:          hashesPath,
		createMissingHashes: createMissingHashes,
		saveTmpImagesPath:   saveTmpImagesPath,
		hashType:            hashType,
	}
}

// calculateHash converts image data to RGBA using VIPS and calculates hash
func (m *ImageHashCacheMatcher) calculateHash(t *testing.T, key string, buf []byte) *ImageHash {
	t.Helper()

	// Load image as RGBA
	goImg, err := LoadImage(bytes.NewReader(buf))
	require.NoError(t, err, "failed to load image for key %s", key)

	// Calculate hash
	hash, err := NewImageHash(goImg, m.hashType)
	require.NoError(t, err)

	return hash
}

// ImageMatches is a testing helper, which accepts image as reader, calculates
// hash and compares it with a hash saved to testdata/test-hashes folder.
func (m *ImageHashCacheMatcher) ImageMatches(t *testing.T, img io.Reader, key string, maxDistance int) {
	t.Helper()

	// Read image in memory
	buf, err := io.ReadAll(img)
	require.NoError(t, err)

	// Save tmp image if requested
	m.saveTmpImage(t, key, buf)

	// Calculate hash using shared helper
	sourceHash := m.calculateHash(t, key, buf)

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
	targetHash, err := LoadImageHash(f)
	require.NoError(t, err, "failed to load target hash from %s", hashPath)

	// Ensure distance is OK
	distance, err := sourceHash.Distance(targetHash)
	require.NoError(t, err, "failed to calculate hash distance for %s", key)

	require.LessOrEqual(t, distance, maxDistance, "image hashes are too different for %s: distance %d", key, distance)
}

// makeTargetPath creates the target directory and returns file path for saving
// the image or hash.
func (m *ImageHashCacheMatcher) makeTargetPath(t *testing.T, base, folder, filename, ext string) (string, error) {
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
func (m *ImageHashCacheMatcher) saveTmpImage(t *testing.T, key string, buf []byte) {
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
