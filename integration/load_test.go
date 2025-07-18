//go:build integration
// +build integration

package integration

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corona10/goimagehash"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	similarityThreshold = 5 // Distance between images to be considered similar
)

// testLoadFolder fetches images iterates over images in the specified folder,
// runs imgproxy on each image, and compares the result with the reference image
// which is expected to be in the `integration` folder with the same name
// but with `.png` extension.
func testLoadFolder(t *testing.T, cs, sourcePath, folder string) {
	t.Logf("Testing folder: %s", folder)

	walkPath := path.Join(sourcePath, folder)

	// Iterate over the files in the source folder
	err := filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// get the base name of the file (8-bpp.png)
		basePath := filepath.Base(path)

		// Replace the extension with .png
		referencePath := strings.TrimSuffix(basePath, filepath.Ext(basePath)) + ".png"

		// Construct the full path to the reference image (integration/ folder)
		referencePath = filepath.Join(sourcePath, "integration", folder, referencePath)

		// Construct the source URL for imgproxy (no processing)
		sourceUrl := fmt.Sprintf("insecure/plain/local:///%s/%s@png", folder, basePath)

		imgproxyImageBytes := fetchImage(t, cs, sourceUrl)
		imgproxyImage, err := png.Decode(bytes.NewReader(imgproxyImageBytes))
		require.NoError(t, err, "Failed to decode PNG image from imgproxy for %s", basePath)

		referenceFile, err := os.Open(referencePath)
		require.NoError(t, err)
		defer referenceFile.Close()

		referenceImage, err := png.Decode(referenceFile)
		require.NoError(t, err, "Failed to decode PNG reference image for %s", referencePath)

		hash1, err := goimagehash.DifferenceHash(imgproxyImage)
		require.NoError(t, err)

		hash2, err := goimagehash.DifferenceHash(referenceImage)
		require.NoError(t, err)

		distance, err := hash1.Distance(hash2)
		require.NoError(t, err)

		assert.LessOrEqual(t, distance, similarityThreshold,
			"Image %s differs from reference image %s by %d, which is greater than the allowed threshold of %d",
			basePath, referencePath, distance, similarityThreshold)

		return nil
	})

	require.NoError(t, err)
}

// TestLoadSaveToPng ensures that our load pipeline works,
// including standard and custom loaders. For each source image
// in the folder, it does the passthrough request through imgproxy:
// no processing, just convert format of the source file to png.
// Then, it compares the result with the reference image.
func TestLoadSaveToPng(t *testing.T) {
	ctx := t.Context()

	// TODO: Will be moved to test suite (like in processing_test.go)
	// Since we use SupportsLoad, we need to initialize vips
	defer vips.Shutdown() // either way it needs to be deinitialized
	err := vips.Init()
	require.NoError(t, err, "Failed to initialize vips")

	path := downloadTestImages(t)
	cs := startImgproxy(t, ctx, path)

	if vips.SupportsLoad(imagetype.GIF) {
		testLoadFolder(t, cs, path, "gif")
	}

	if vips.SupportsLoad(imagetype.JPEG) {
		testLoadFolder(t, cs, path, "jpg")
	}

	if vips.SupportsLoad(imagetype.HEIC) {
		testLoadFolder(t, cs, path, "heif")
	}

	if vips.SupportsLoad(imagetype.JXL) {
		testLoadFolder(t, cs, path, "jxl")
	}

	if vips.SupportsLoad(imagetype.SVG) {
		testLoadFolder(t, cs, path, "svg")
	}

	if vips.SupportsLoad(imagetype.TIFF) {
		testLoadFolder(t, cs, path, "tiff")
	}

	if vips.SupportsLoad(imagetype.WEBP) {
		testLoadFolder(t, cs, path, "webp")
	}

	if vips.SupportsLoad(imagetype.BMP) {
		testLoadFolder(t, cs, path, "bmp")
	}

	if vips.SupportsLoad(imagetype.ICO) {
		testLoadFolder(t, cs, path, "ico")
	}
}
