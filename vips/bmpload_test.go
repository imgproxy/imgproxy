package vips_test

import (
	"image/png"
	"testing"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/options"
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
		// bottom-up, height 1000001: over the vips_sequential tile_height maximum
		"bmp-height-bottom-up-too-tall.bmp",
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

// Regression test for a heap-buffer-overflow caused by an unvalidated bpp.
// For uncompressed (BI_RGB) BMP input, any bpp >= 24 used to be routed into the
// 24/32-bit row generator, which read a row sized from the unvalidated bpp into
// a row buffer allocated for a 4-byte-per-pixel maximum.
// bmp-invalid-bpp.bmp declares bpp=40 at width=512, which used to read a
// 2560-byte row into a 2052-byte allocation (a 508-byte overflow).
func TestBMPLoadUnsupportedBitDepth(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })
	id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read("bmp-invalid-bpp.bmp"))
	defer id.Close()

	img := new(vips.Image)
	defer img.Clear()

	err := img.Load(id, 1.0, 0, 1)
	if err == nil {
		// The loader is lazy: force pixel decoding if the header was accepted
		err = img.CopyMemory()
	}

	require.ErrorContains(t, err, "unsupported bit depth")
}

// 16-bit BMPs: BI_RGB defaults to RGB555 and BITFIELDS may declare 5-5-5 masks.
// Both used to decode incorrectly (black / halved channel range). The source
// pixels are 0x7FFF (white in RGB555), so the result must be near-white.
func TestBMPLoad16bitRGB555(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })

	testCases := []string{
		"bmp-16bpp-rgb555-bi_rgb.bmp",
		"bmp-16bpp-rgb555-bitfields.bmp",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read(filename))
			defer id.Close()

			img := new(vips.Image)
			defer img.Clear()
			require.NoError(t, img.Load(id, 1.0, 0, 1))

			out, err := img.Save(imagetype.PNG, 0, options.New())
			require.NoError(t, err)
			defer out.Close()

			r := out.Reader()

			decoded, err := png.Decode(r)
			require.NoError(t, err)

			bounds := decoded.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					cr, cg, cb, _ := decoded.At(x, y).RGBA()
					require.GreaterOrEqual(t, cr>>8, uint32(248))
					require.GreaterOrEqual(t, cg>>8, uint32(248))
					require.GreaterOrEqual(t, cb>>8, uint32(248))
				}
			}
		})
	}
}

// RLE8 MOVE_TO with dy > 0 must move exactly dy lines down from the current one.
// Stream: run(idx1); EOL; MOVE_TO(0, 1); run(idx2); EOL; EOF.
// The image is bottom-up (4 rows), so stream rows 0 and 2 are image rows 3 and 1.
func TestBMPLoadRLE8MoveToDy(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })
	id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read("bmp-rle8-move-to.bmp"))
	defer id.Close()

	img := new(vips.Image)
	defer img.Clear()
	require.NoError(t, img.Load(id, 1.0, 0, 1))

	out, err := img.Save(imagetype.PNG, 0, options.New())
	require.NoError(t, err)
	defer out.Close()

	decoded, err := png.Decode(out.Reader())
	require.NoError(t, err)

	// Expected red channel (8-bit) per image row, top to bottom
	expected := []uint32{0x00, 0x50, 0x00, 0x38}

	for y, want := range expected {
		cr, _, _, _ := decoded.At(0, y).RGBA()
		require.Equal(t, want, cr>>8, "row %d", y)
	}
}
