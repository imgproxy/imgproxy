package vips_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips"
	"github.com/stretchr/testify/require"
)

func TestBMPRLE8AbsoluteRunExceedsWidth(t *testing.T) {
	// A 1x1 image whose absolute run contains 255 pixels plus a padding byte,
	// so the loader has to read 256 bytes into the row buffer.
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })
	id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read("bmp-rle8-overlong-run.bmp"))
	defer id.Close()

	img := new(vips.Image)
	defer img.Clear()
	require.NoError(t, img.Load(id, 1.0, 0, 1))

	// The loader is lazy; copying to memory forces the RLE run to be decoded.
	require.NoError(t, img.CopyMemory())
}

// Crafted BMP headers with dimensions that used to overflow the row buffer size
// computation (GHSA-c4f2-mxv4-6fp5). The loader must reject them cleanly.
func TestBMPLoadInvalidDimensions(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })

	testCases := []string{
		"bmp-width-overflow.bmp",  // width 536870911, 8 bpp
		"bmp-width-int-max.bmp",   // width 2147483647, 24 bpp
		"bmp-width-too-large.bmp", // width 10000001, 24 bpp
		"bmp-height-int-min.bmp",  // height INT32_MIN
		"bmp-area-too-large.bmp",  // 100000x100000, 24 bpp
		"bmp-32bpp-overflow.bmp",  // width 0x20000000, 32 bpp
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read(filename))
			defer id.Close()

			img := new(vips.Image)
			defer img.Clear()

			err := img.Load(id, 1.0, 0, 1)
			if err == nil {
				// The loader is lazy: force pixel decoding if the header was accepted
				err = img.CopyMemory()
			}

			require.ErrorContains(t, err, "unsupported BMP image dimensions")
		})
	}
}
