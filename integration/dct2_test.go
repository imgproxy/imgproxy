package integration

// This file contains tests for DCT2 image hashing functionality.
// Unfortunately, Cgo can not be called from tests so we have to rely on
// test_main which initializes VIPS using imgproxy.Init().
// These test should reside in testutil.

import (
	"bytes"
	"testing"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type DCT2HashTestSuite struct {
	suite.Suite

	TestData *testutil.TestDataProvider
}

func (s *DCT2HashTestSuite) SetupSuite() {
	s.TestData = testutil.NewTestDataProvider(s.T)
}

func (s *DCT2HashTestSuite) TestDCT2HashCalc() {
	testCases := []struct {
		filename           string
		expectedHashLength int
	}{
		// BMP files
		{"test-images/bmp/1-bpp.bmp", 95},
		{"test-images/bmp/16-bpp.bmp", 95},
		{"test-images/bmp/24-bpp-no-alpha-mask.bmp", 95},
		{"test-images/bmp/24-bpp.bmp", 95},
		{"test-images/bmp/32-bpp-with-alpha-self-gen.bmp", 95},
		{"test-images/bmp/32-bpp-with-alpha.bmp", 95},
		{"test-images/bmp/4-bpp.bmp", 95},
		{"test-images/bmp/8-bpp-rle-move-to-x.bmp", 95},
		{"test-images/bmp/8-bpp-rle-single-color.bmp", 95},
		{"test-images/bmp/8-bpp-rle-small.bmp", 95},
		{"test-images/bmp/8-bpp-rle.bmp", 95},
		{"test-images/bmp/8-bpp.bmp", 95},

		// PNG files
		{"test-images/png/1-bpp.png", 31},
		{"test-images/png/16-bpp-grayscale.png", 31},
		{"test-images/png/16-bpp-linear.png", 95},
		{"test-images/png/16-bpp.png", 95},
		{"test-images/png/8-bpp-grayscale.png", 31},
		{"test-images/png/8-bpp.png", 95},
		{"test-images/png/png.png", 95},

		// HEIF files
		{"test-images/heif/8-bpp.heif", 95},
		{"test-images/heif/heif.heif", 95},

		// GIF files
		{"test-images/gif/gif.gif", 95},

		// TIFF files
		{"test-images/tiff/16-bpp-grayscale.tiff", 31},
		{"test-images/tiff/16-bpp.tiff", 95},
		{"test-images/tiff/32-bpp-linear.tiff", 95},
		{"test-images/tiff/4-bpp-grayscale.tiff", 31},
		{"test-images/tiff/8-bpp-grayscale.tiff", 31},
		{"test-images/tiff/8-bpp.tiff", 95},
		{"test-images/tiff/tiff.tiff", 95},

		// SVG files
		{"test-images/svg/svg.svg", 95},
	}

	for _, tc := range testCases {
		s.Run(tc.filename, func() {
			// Calculate DCT hash
			hash, err := testutil.NewImageHashFromPath(s.TestData.Path(tc.filename), testutil.HashTypeDct)
			s.Require().NoError(err)
			s.Require().NotNil(hash)

			// Dump hash to buffer to check length
			var buf bytes.Buffer
			err = hash.Dump(&buf)
			s.Require().NoError(err)

			// Calculate actual hash length from dumped data
			// Format: 1 byte (type) + 4 bytes (uint32 length) + N*8 bytes (float64 values)
			// So: N = (totalBytes - 5) / 8
			dumpedBytes := buf.Len()
			actualHashLength := (dumpedBytes - 5) / 8

			s.Require().Equal(tc.expectedHashLength, actualHashLength)
		})
	}
}

func (s *DCT2HashTestSuite) TestDCT2HashDifference() {
	// Load original image hash
	originalHash, err := testutil.NewImageHashFromPath(s.TestData.Path("dct2/original.png"), testutil.HashTypeDct)
	s.Require().NoError(err)

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
			testHash, err := testutil.NewImageHashFromPath(s.TestData.Path(tc.filename), testutil.HashTypeDct)
			s.Require().NoError(err)
			s.Require().NotNil(testHash)

			distance, err := originalHash.Distance(testHash)
			s.Require().NoError(err)

			// Allow 1% tolerance, with a minimum absolute delta of 0.001
			delta := tc.expectedDistance*0.01 + 0.001
			s.InDelta(tc.expectedDistance, distance, float64(delta), "unexpected distance for %s", tc.filename)
		})
	}
}

func TestDCT2Hash(t *testing.T) {
	suite.Run(t, new(DCT2HashTestSuite))
}
