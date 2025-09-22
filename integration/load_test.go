package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/corona10/goimagehash"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

const (
	maxDistance = 0 // maximum image distance
)

type LoadTestSuite struct {
	Suite

	testImagesPath      string
	hashesPath          string
	saveTmpImagesPath   string
	createMissingHashes bool
}

func (s *LoadTestSuite) SetupTest() {
	s.testImagesPath = s.TestData.Path("test-images")
	s.hashesPath = s.TestData.Path("test-hashes")
	s.saveTmpImagesPath = os.Getenv("TEST_SAVE_TMP_IMAGES")
	s.createMissingHashes = len(os.Getenv("TEST_CREATE_MISSING_HASHES")) > 0

	s.Config().Security.DefaultOptions.MaxAnimationFrames = 999
	s.Config().Server.DevelopmentErrorsMode = true
	s.Config().Fetcher.Transport.Local.Root = s.testImagesPath
}

// testLoadFolder fetches images iterates over images in the specified folder,
// runs imgproxy on each image, and compares the result with the reference image
// which is expected to be in the `integration` folder with the same name
// but with `.png` extension.
func (s *LoadTestSuite) testLoadFolder(folder string) {
	walkPath := path.Join(s.testImagesPath, folder)

	// Iterate over the files in the source folder
	err := filepath.Walk(walkPath, func(path string, info os.FileInfo, err error) error {
		s.Require().NoError(err)

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// get the base name of the file (8-bpp.png)
		baseName := filepath.Base(path)

		// Construct the source URL for imgproxy (no processing)
		sourceUrl := fmt.Sprintf("/insecure/plain/local:///%s/%s@bmp", folder, baseName)

		// Read source image from imgproxy
		sourceImageData := s.fetchImage(sourceUrl)
		defer sourceImageData.Close()

		// Save the source image if requested
		s.saveTmpImage(folder, baseName, sourceImageData)

		// Calculate image hash of the image returned by imgproxy
		var sourceImage vips.Image
		s.Require().NoError(sourceImage.Load(sourceImageData, 1, 1.0, 1))
		defer sourceImage.Clear()

		sourceHash, err := testutil.ImageDifferenceHash(unsafe.Pointer(sourceImage.VipsImage))
		s.Require().NoError(err)

		// Calculate image hash path (create folder if missing)
		hashPath, err := s.makeTargetPath(s.hashesPath, s.T().Name(), baseName, "hash")
		s.Require().NoError(err)

		// Try to read or create the hash file
		f, err := os.Open(hashPath)
		if os.IsNotExist(err) {
			// If the hash file does not exist, and we are not allowed to create it, fail
			if !s.createMissingHashes {
				s.Require().NoError(err, "failed to read target hash from %s, use TEST_CREATE_MISSING_HASHES=true to create it", hashPath)
			}

			h, hashErr := os.Create(hashPath)
			s.Require().NoError(hashErr, "failed to create target hash file %s", hashPath)
			defer h.Close()

			hashErr = sourceHash.Dump(h)
			s.Require().NoError(hashErr, "failed to write target hash to %s", hashPath)

			s.T().Logf("Created missing hash in %s", hashPath)
		} else {
			// Otherwise, if there is no error or error is something else
			s.Require().NoError(err)

			targetHash, err := goimagehash.LoadImageHash(f)
			s.Require().NoError(err, "failed to load target hash from %s", hashPath)

			distance, err := sourceHash.Distance(targetHash)
			s.Require().NoError(err, "failed to calculate hash distance for %s", baseName)

			s.Require().LessOrEqual(distance, maxDistance, "image hashes are too different for %s: distance %d", baseName, distance)
		}

		return nil
	})

	s.Require().NoError(err)
}

// fetchImage fetches an image from the imgproxy server
func (s *LoadTestSuite) fetchImage(path string) imagedata.ImageData {
	resp := s.GET(path)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK, got %d, path: %s", resp.StatusCode, path)

	bytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err, "failed to read response body from %s", path)

	d, err := s.Imgproxy().ImageDataFactory().NewFromBytes(bytes)
	s.Require().NoError(err, "failed to load image from bytes for %s", path)

	return d
}

// makeTargetPath creates the target directory and returns file path for saving
// the image or hash.
func (s *LoadTestSuite) makeTargetPath(base, folder, filename, ext string) (string, error) {
	// Create the target directory if it doesn't exist
	targetDir := path.Join(base, folder)
	err := os.MkdirAll(targetDir, 0755)
	s.Require().NoError(err, "failed to create %s target directory", targetDir)

	// Replace the extension with the detected one
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + "." + ext

	// Create the target file
	targetPath := path.Join(targetDir, filename)

	return targetPath, nil
}

// saveTmpImage saves the provided image data to a temporary file
func (s *LoadTestSuite) saveTmpImage(folder, filename string, imageData imagedata.ImageData) {
	if s.saveTmpImagesPath == "" {
		return
	}

	// Detect the image type to get the correct extension
	ext, err := imagetype.Detect(imageData.Reader())
	s.Require().NoError(err)

	targetPath, err := s.makeTargetPath(s.saveTmpImagesPath, folder, filename, ext.String())
	s.Require().NoError(err, "failed to create TEST_SAVE_TMP_IMAGES target path for %s/%s", folder, filename)

	targetFile, err := os.Create(targetPath)
	s.Require().NoError(err, "failed to create TEST_SAVE_TMP_IMAGES target file %s", targetPath)
	defer targetFile.Close()

	// Write the image data to the file
	_, err = io.Copy(targetFile, imageData.Reader())
	s.Require().NoError(err, "failed to write to TEST_SAVE_TMP_IMAGES target file %s", targetPath)

	s.T().Logf("Saved temporary image to %s", targetPath)
}

// TestLoadSaveToPng ensures that our load pipeline works,
// including standard and custom loaders. For each source image
// in the folder, it does the passthrough request through imgproxy:
// no processing, just convert format of the source file to png.
// Then, it compares the result with the reference image.
func (s *LoadTestSuite) TestLoadSaveToPng() {
	testCases := []struct {
		name       string
		imageType  imagetype.Type
		folderName string
	}{
		{"GIF", imagetype.GIF, "gif"},
		{"JPEG", imagetype.JPEG, "jpg"},
		{"HEIC", imagetype.HEIC, "heif"},
		{"JXL", imagetype.JXL, "jxl"},
		{"SVG", imagetype.SVG, "svg"},
		{"TIFF", imagetype.TIFF, "tiff"},
		{"WEBP", imagetype.WEBP, "webp"},
		{"BMP", imagetype.BMP, "bmp"},
		{"ICO", imagetype.ICO, "ico"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			if vips.SupportsLoad(tc.imageType) {
				s.testLoadFolder(tc.folderName)
			} else {
				t.Skipf("%s format not supported by VIPS", tc.name)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(LoadTestSuite))
}
