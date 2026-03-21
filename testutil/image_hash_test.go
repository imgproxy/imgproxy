package testutil_test

import (
	"bytes"
	"testing"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ImageHashTestSuite struct {
	suite.Suite

	TestData *testutil.TestDataProvider
}

func (s *ImageHashTestSuite) SetupSuite() {
	s.TestData = testutil.NewTestDataProvider(s.T)
}

func (s *ImageHashTestSuite) TestDCT2HashCalc() {
	testCases := []struct {
		filename           string
		expectedHashLength int
	}{
		// PNG files
		{"test-images/png/1-bpp.png", 64},
		{"test-images/png/16-bpp-grayscale.png", 64},
		{"test-images/png/16-bpp-linear.png", 192},
		{"test-images/png/16-bpp.png", 192},
		{"test-images/png/8-bpp-grayscale.png", 64},
		{"test-images/png/8-bpp.png", 192},
		{"test-images/png/png.png", 192},

		// HEIF files
		{"test-images/heif/8-bpp.heif", 192},
		{"test-images/heif/heif.heif", 192},

		// GIF files
		{"test-images/gif/gif.gif", 192},

		// TIFF files
		{"test-images/tiff/16-bpp-grayscale.tiff", 64},
		{"test-images/tiff/16-bpp.tiff", 192},
		{"test-images/tiff/32-bpp-linear.tiff", 192},
		{"test-images/tiff/4-bpp-grayscale.tiff", 64},
		{"test-images/tiff/8-bpp-grayscale.tiff", 64},
		{"test-images/tiff/8-bpp.tiff", 192},
		{"test-images/tiff/tiff.tiff", 192},

		// SVG files
		{"test-images/svg/svg.svg", 192},
	}

	for _, tc := range testCases {
		s.Run(tc.filename, func() {
			// Calculate DCT hash
			hash := s.calcHash(tc.filename, testutil.HashTypeDct)

			// Check that dumped hash length matches expected length
			// (multiply by 4 because DCT hash is stored as float32, which is 4 bytes)
			s.Require().Equal(tc.expectedHashLength*4, hash.Len(), "unexpected hash length")

			// Dump hash to buffer and load it back to check consistency
			var buf bytes.Buffer
			err := hash.Dump(&buf)
			s.Require().NoError(err)

			loadedHash, err := testutil.LoadImageHash(&buf)
			s.Require().NoError(err)
			s.Require().NotNil(loadedHash)

			// Check that loaded hash matches original hash
			distance, err := hash.Distance(loadedHash)
			s.Require().NoError(err)
			s.Require().InDelta(0, distance, 1e-10, "loaded hash does not match original hash")
		})
	}
}

func (s *ImageHashTestSuite) TestDCT2HashDifference() {
	// Load original image hash
	originalHash := s.calcHash("dct2/original.png", testutil.HashTypeDct)

	testCases := []struct {
		filename         string
		expectedDistance float32
	}{
		{"dct2/brightness-minus1.png", 19.090097},
		{"dct2/brightness-minus10.png", 2008.260088},
		{"dct2/brightness-minus30.png", 19580.743012},
		{"dct2/brightness-minus50.png", 59795.423867},
		{"dct2/brightness-minus1-8bit.png", 35.761968},
		{"dct2/brightness-minus10-8bit.png", 2170.649351},
		{"dct2/brightness-minus30-8bit.png", 20112.324396},
		{"dct2/brightness-minus50-8bit.png", 60799.682862},
		{"dct2/jpeg-q99.jpg", 1.230422},
		{"dct2/jpeg-q80.jpg", 1.275000},
		{"dct2/jpeg-q60.jpg", 1.422987},
		{"dct2/jpeg-q40.jpg", 1.904982},
		{"dct2/jpeg-q20.jpg", 2.367389},
		{"dct2/jpeg-q1.jpg", 95.573785},
		{"dct2/hue-plus1.png", 50.101667},
		{"dct2/hue-plus10.png", 4834.754080},
		{"dct2/hue-plus50.png", 87091.950577},
		{"dct2/region-5x5-hue-plus1.png", 0.000002},
		{"dct2/region-5x5-hue-plus10.png", 0.000472},
		{"dct2/region-5x5-hue-plus50.png", 0.212683},
		{"dct2/region-10x10-hue-plus1.png", 0.000044},
		{"dct2/region-10x10-hue-plus10.png", 0.010890},
		{"dct2/region-10x10-hue-plus50.png", 3.514549},
		{"dct2/region-30x30-hue-plus1.png", 0.004835},
		{"dct2/region-30x30-hue-plus10.png", 0.782760},
		{"dct2/region-30x30-hue-plus50.png", 185.500887},
		{"dct2/region-50x50-hue-plus1.png", 0.035596},
		{"dct2/region-50x50-hue-plus10.png", 4.304423},
		{"dct2/region-50x50-hue-plus50.png", 840.811707},
	}

	for _, tc := range testCases {
		s.Run(tc.filename, func() {
			testHash := s.calcHash(tc.filename, testutil.HashTypeDct)

			distance, err := originalHash.Distance(testHash)
			s.Require().NoError(err)

			// Allow 1% tolerance, with a minimum absolute delta of 0.001
			delta := tc.expectedDistance*0.01 + 0.001
			s.InDelta(tc.expectedDistance, distance, float64(delta), "unexpected distance")
		})
	}
}

func (s *ImageHashTestSuite) TestSHA256HashCalc() {
	testCases := []struct {
		filename     string
		expectedHash string
	}{
		{
			"test-images/png/1-bpp.png",
			"c4e0a85ba474b6d6190d9f1d477f5d7ddab041a9951a1587f0ec8487d81cc1f9",
		},
		{
			"test-images/png/16-bpp-grayscale.png",
			"9bc0007bfa2c6a1066e045fe84b4b8f731f0c1e0665c67ba012a8ca9d714d4f6",
		},
		{
			"test-images/png/16-bpp-linear.png",
			"354bb65b9efce0b89990b49ff4609051186940bbbd261b13f225fa4710d19357",
		},
		{
			"test-images/png/16-bpp.png",
			"14c680b197a3110df5ae58a0c57092de9ef10b4559b52701ce5f7ee69fbb9715",
		},
		{
			"test-images/png/8-bpp-grayscale.png",
			"733d30d033d396f46cf2d0e99f542b5f1160129c0ca008b01903f8debb8044e0",
		},
		{
			"test-images/png/8-bpp.png",
			"880ca38b6a98787ed85272a3f9567a29ab417afd9af8999960bd149d5029c450",
		},
		{
			"test-images/png/png.png",
			"880ca38b6a98787ed85272a3f9567a29ab417afd9af8999960bd149d5029c450",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.filename, func() {
			// Calculate SHA-256 hash
			hash := s.calcHash(tc.filename, testutil.HashTypeSHA256)

			// Check that calculated hash matches expected hash
			s.Require().Equal(
				"SHA256:"+tc.expectedHash, hash.String(),
				"unexpected hash value for %s", tc.filename,
			)

			// Dump hash to buffer and load it back to check consistency
			var buf bytes.Buffer
			err := hash.Dump(&buf)
			s.Require().NoError(err)

			loadedHash, err := testutil.LoadImageHash(&buf)
			s.Require().NoError(err)
			s.Require().NotNil(loadedHash)

			// Check that loaded hash matches original hash
			distance, err := hash.Distance(loadedHash)
			s.Require().NoError(err)
			s.Require().InDelta(0, distance, 1e-10, "loaded hash does not match original hash")
		})
	}
}

func (s *ImageHashTestSuite) TestSHA256HashDifference() {
	// Load original image hash
	originalHash := s.calcHash("dct2/original.png", testutil.HashTypeSHA256)

	testCases := []string{
		"dct2/brightness-minus1.png",
		"dct2/jpeg-q80.jpg",
		"dct2/hue-plus1.png",
		"dct2/region-5x5-hue-plus1.png",
	}

	for _, filename := range testCases {
		s.Run(filename, func() {
			testHash := s.calcHash(filename, testutil.HashTypeSHA256)

			distance, err := originalHash.Distance(testHash)
			s.Require().NoError(err)

			// Check that the hash of the modified image is different from the original hash
			s.Require().NotEqual(0, distance, "hash should differ for modified image")
		})
	}
}

func (s *ImageHashTestSuite) calcHash(
	filename string, hashType testutil.ImageHashType,
) *testutil.ImageHash {
	hash, err := testutil.NewImageHash(s.TestData.Reader(filename), hashType)
	s.Require().NoError(err)
	s.Require().NotNil(hash)
	return hash
}

func TestImageHash(t *testing.T) {
	suite.Run(t, new(ImageHashTestSuite))
}
