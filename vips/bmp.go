// Based on https://cs.opensource.google/go/x/image/+/6944b10b:bmp/reader.go
// and https://cs.opensource.google/go/x/image/+/6944b10b:bmp/writer.go
package vips

/*
#include "vips.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"unsafe"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type bmpHeader struct {
	sigBM           [2]byte
	fileSize        uint32
	resverved       [2]uint16
	pixOffset       uint32
	dibHeaderSize   uint32
	width           uint32
	height          uint32
	colorPlane      uint16
	bpp             uint16
	compression     uint32
	imageSize       uint32
	xPixelsPerMeter uint32
	yPixelsPerMeter uint32
	colorUse        uint32
	colorImportant  uint32
}

// errBmpUnsupported means that the input BMP image uses a valid but unsupported
// feature.
var errBmpUnsupported = errors.New("unsupported BMP image")

func readUint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func readUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func prepareBmpCanvas(width, height, bands int) (*C.VipsImage, []byte, error) {
	var tmp *C.VipsImage

	if C.vips_black_go(&tmp, C.int(width), C.int(height), C.int(bands)) != 0 {
		return nil, nil, Error()
	}

	data := unsafe.Pointer(C.vips_image_get_data(tmp))
	datalen := int(tmp.Bands) * int(tmp.Xsize) * int(tmp.Ysize)

	return tmp, ptrToBytes(data, datalen), nil
}

// decodeBmpPaletted reads an 8 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func (img *Image) decodeBmpPaletted(r io.Reader, width, height, bpp int, palette []Color, topDown bool) error {
	tmp, imgData, err := prepareBmpCanvas(width, height, 3)
	if err != nil {
		return err
	}

	defer func() {
		if rerr := recover(); rerr != nil {
			C.clear_image(&tmp)
			panic(rerr)
		}
	}()

	// Each row is 4-byte aligned.
	cap := 8 / bpp
	b := make([]byte, ((width+cap-1)/cap+3)&^3)

	y0, y1, yDelta := height-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, height, +1
	}

	stride := width * 3

	for y := y0; y != y1; y += yDelta {
		if _, err = io.ReadFull(r, b); err != nil {
			C.clear_image(&tmp)
			return err
		}

		p := imgData[y*stride : (y+1)*stride]

		j, bit := 0, 8-bpp
		for i := 0; i < len(p); i += 3 {
			pind := (b[j] >> bit) & (1<<bpp - 1)

			if bit == 0 {
				bit = 8 - bpp
				j++
			} else {
				bit -= bpp
			}

			c := palette[pind]

			p[i+0] = c.R
			p[i+1] = c.G
			p[i+2] = c.B
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	var bitdepth int
	colors := len(palette)

	switch {
	case colors > 16:
		bitdepth = 8
	case colors > 4:
		bitdepth = 4
	case colors > 2:
		bitdepth = 2
	}

	img.SetInt("palette-bit-depth", bitdepth)

	return nil
}

// decodeBmpRGB reads a 24/32 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func (img *Image) decodeBmpRGB(r io.Reader, width, height, bands int, topDown, noAlpha bool) error {
	if bands != 3 && bands != 4 {
		return errBmpUnsupported
	}

	imgBands := 3
	if bands == 4 && !noAlpha {
		// Create RGBA image only when source has 4 bands and the last one is alpha
		imgBands = 4
	}

	tmp, imgData, err := prepareBmpCanvas(width, height, imgBands)
	if err != nil {
		return err
	}

	defer func() {
		if rerr := recover(); rerr != nil {
			C.clear_image(&tmp)
			panic(rerr)
		}
	}()

	// Each row is 4-byte aligned.
	b := make([]byte, (bands*width+3)&^3)

	y0, y1, yDelta := height-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, height, +1
	}

	stride := width * imgBands

	for y := y0; y != y1; y += yDelta {
		if _, err = io.ReadFull(r, b); err != nil {
			C.clear_image(&tmp)
			return err
		}

		p := imgData[y*stride : (y+1)*stride]
		for i, j := 0, 0; i < len(p); i, j = i+imgBands, j+bands {
			// BMP images are stored in BGR order rather than RGB order.
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]

			if imgBands == 4 {
				p[i+3] = b[j+3]
			}
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) loadBmp(data []byte) error {
	// We only support those BMP images that are a BITMAPFILEHEADER
	// immediately followed by a BITMAPINFOHEADER.
	const (
		fileHeaderLen   = 14
		infoHeaderLen   = 40
		v4InfoHeaderLen = 108
		v5InfoHeaderLen = 124
	)

	r := bytes.NewReader(data)

	var b [1024]byte
	if _, err := io.ReadFull(r, b[:fileHeaderLen+4]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	if string(b[:2]) != "BM" {
		return errors.New("not a BMP image")
	}

	offset := readUint32(b[10:14])
	infoLen := readUint32(b[14:18])
	if infoLen != infoHeaderLen && infoLen != v4InfoHeaderLen && infoLen != v5InfoHeaderLen {
		return errBmpUnsupported
	}

	if _, err := io.ReadFull(r, b[fileHeaderLen+4:fileHeaderLen+infoLen]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	width := int(int32(readUint32(b[18:22])))
	height := int(int32(readUint32(b[22:26])))
	topDown := false

	if height < 0 {
		height, topDown = -height, true
	}
	if width <= 0 || height <= 0 {
		return errBmpUnsupported
	}

	// We only support 1 plane and 8, 24 or 32 bits per pixel and no
	// compression.
	planes, bpp, compression := readUint16(b[26:28]), readUint16(b[28:30]), readUint32(b[30:34])

	// if compression is set to BITFIELDS, but the bitmask is set to the default bitmask
	// that would be used if compression was set to 0, we can continue as if compression was 0
	if compression == 3 && infoLen > infoHeaderLen &&
		readUint32(b[54:58]) == 0xff0000 && readUint32(b[58:62]) == 0xff00 &&
		readUint32(b[62:66]) == 0xff && readUint32(b[66:70]) == 0xff000000 {
		compression = 0
	}

	if planes != 1 || compression != 0 {
		return errBmpUnsupported
	}

	switch bpp {
	case 1, 2, 4, 8:
		palColors := readUint32(b[46:50])

		_, err := io.ReadFull(r, b[:palColors*4])
		if err != nil {
			return err
		}

		palette := make([]Color, palColors)
		for i := range palette {
			// BMP images are stored in BGR order rather than RGB order.
			// Every 4th byte is padding.
			palette[i] = Color{b[4*i+2], b[4*i+1], b[4*i+0]}
		}

		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return err
		}

		return img.decodeBmpPaletted(r, width, height, int(bpp), palette, topDown)
	case 24:
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return err
		}
		return img.decodeBmpRGB(r, width, height, 3, topDown, true)
	case 32:
		noAlpha := true
		if infoLen >= 70 {
			// Alpha mask is empty, so no alpha here
			noAlpha = readUint32(b[66:70]) == 0
		}

		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return err
		}

		return img.decodeBmpRGB(r, width, height, 4, topDown, noAlpha)
	}

	return errBmpUnsupported
}

func (img *Image) saveAsBmp() (*imagedata.ImageData, error) {
	width, height := img.Width(), img.Height()

	h := &bmpHeader{
		sigBM:           [2]byte{'B', 'M'},
		fileSize:        14 + 40,
		resverved:       [2]uint16{0, 0},
		pixOffset:       14 + 40,
		dibHeaderSize:   40,
		width:           uint32(width),
		height:          uint32(height),
		colorPlane:      1,
		bpp:             24,
		compression:     0,
		xPixelsPerMeter: 2835,
		yPixelsPerMeter: 2835,
		colorUse:        0,
		colorImportant:  0,
	}

	lineSize := (width*3 + 3) &^ 3

	h.imageSize = uint32(height * lineSize)
	h.fileSize += h.imageSize

	buf := new(bytes.Buffer)
	buf.Grow(int(h.fileSize))

	if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
		return nil, err
	}

	if err := img.CopyMemory(); err != nil {
		return nil, err
	}

	data := unsafe.Pointer(C.vips_image_get_data(img.VipsImage))
	datalen := int(img.VipsImage.Bands) * int(img.VipsImage.Xsize) * int(img.VipsImage.Ysize)
	imgData := ptrToBytes(data, datalen)

	bands := int(img.VipsImage.Bands)
	stride := width * bands

	line := make([]byte, lineSize)

	for y := height - 1; y >= 0; y-- {
		min := y * stride
		max := min + stride

		for i, j := min, 0; i < max; i, j = i+bands, j+3 {
			line[j+0] = imgData[i+2]
			line[j+1] = imgData[i+1]
			line[j+2] = imgData[i+0]

			if bands == 4 && imgData[i+3] < 255 {
				line[j+0] = byte(int(line[j+0]) * int(imgData[i+3]) / 255)
				line[j+1] = byte(int(line[j+1]) * int(imgData[i+3]) / 255)
				line[j+2] = byte(int(line[j+2]) * int(imgData[i+3]) / 255)
			}
		}

		if _, err := buf.Write(line); err != nil {
			return nil, err
		}
	}

	return &imagedata.ImageData{
		Type: imagetype.BMP,
		Data: buf.Bytes(),
	}, nil
}
