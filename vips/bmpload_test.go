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
