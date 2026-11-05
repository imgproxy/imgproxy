package vips_test

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/vips"
)

// Per the BMP spec, an RLE4/RLE8 absolute-mode run header is 0x00 <byteCount>,
// followed by byteCount bytes of packed pixel literals. For RLE4 those bytes
// encode 2*byteCount pixels, for RLE8 - byteCount pixels.
//
// buildRle4AbsRunBmp builds an 8x1 BI_RLE4 bitmap whose entire pixel stream is
// one absolute run of 4 bytes (8 pixels: palette indices 1..8) plus EOL/EOF.
// Palette index i maps to gray (32 + 24*i), so a correct decode yields grays
// 56,80,...,224 at x=0..7. Before the fix, the run consumed only 2 of its 4
// bytes and wrote 4 of its 8 pixels, and the remaining 56 78 bytes were parsed
// as a repeat command (palette indices 7,8 twice).
func buildRle4AbsRunBmp() []byte {
	var b bytes.Buffer

	// BITMAPFILEHEADER: BM + size(4) + reserved(4) + data offset(4)
	dataOff := 14 + 40 + 16*4
	fileSize := dataOff + 10
	b.Write([]byte("BM"))
	b.Write([]byte{byte(fileSize), byte(fileSize >> 8), 0, 0, 0, 0, 0, 0, byte(dataOff), byte(dataOff >> 8), 0, 0})

	// BITMAPINFOHEADER: 8x1, planes=1, bpp=4, BI_RLE4
	w := []int32{40, 8, 1}
	u := []uint32{1 | 4<<16, 2, 10, 0, 0, 16, 0}
	for _, v := range w {
		b.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
	}
	for _, v := range u {
		b.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
	}

	// 16-entry palette: index i -> gray (32 + 24*i)
	for i := 0; i < 16; i++ {
		g := byte(32 + i*24)
		b.Write([]byte{g, g, g, 0})
	}

	// RLE stream: absolute run of 4 bytes (pixels 1..8), EOL, EOF
	b.Write([]byte{0x00, 0x04, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x01})
	return b.Bytes()
}

func TestRle4AbsoluteRun(t *testing.T) {
	imgd := imagedata.NewFromBytesWithFormat(imagetype.BMP, buildRle4AbsRunBmp())
	img := &vips.Image{}
	if err := img.Load(imgd, 1.0, 0, 1); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	defer img.Clear()

	if img.Width() != 8 || img.Height() != 1 {
		t.Fatalf("unexpected size %dx%d", img.Width(), img.Height())
	}

	out, err := img.Save(imagetype.PNG, 0, &options.Options{})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	im, err := png.Decode(out.Reader())
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}

	for x := 0; x < 8; x++ {
		r, g, bl, _ := im.At(x, 0).RGBA()
		want := byte(32 + (x+1)*24) // palette index x+1
		if uint8(r>>8) != want || uint8(g>>8) != want || uint8(bl>>8) != want {
			t.Errorf("pixel %d: got (%d,%d,%d), want gray %d (palette index %d)",
				x, uint8(r>>8), uint8(g>>8), uint8(bl>>8), want, x+1)
		}
	}
}
