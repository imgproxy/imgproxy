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
	"github.com/imgproxy/imgproxy/v3/imath"
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

func bmpClearOnPanic(img **C.VipsImage) {
	if rerr := recover(); rerr != nil {
		C.clear_image(img)
		panic(rerr)
	}
}

func bmpSetBitDepth(img *Image, colors int) {
	var bitdepth int

	switch {
	case colors > 16:
		bitdepth = 8
	case colors > 4:
		bitdepth = 4
	case colors > 2:
		bitdepth = 2
	}

	img.SetInt("palette-bit-depth", bitdepth)
}

// decodeBmpPaletted reads an 8/4/2/1 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func (img *Image) decodeBmpPaletted(r io.Reader, width, height, bpp int, palette []Color, topDown bool) error {
	tmp, imgData, err := prepareBmpCanvas(width, height, 3)
	if err != nil {
		return err
	}

	defer bmpClearOnPanic(&tmp)

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

	bmpSetBitDepth(img, len(palette))

	return nil
}

// decodeBmpRLE reads an 8/4 bit-per-pixel RLE-encoded BMP image from r.
func (img *Image) decodeBmpRLE(r io.Reader, width, height, bpp int, palette []Color) error {
	tmp, imgData, err := prepareBmpCanvas(width, height, 3)
	if err != nil {
		return err
	}

	defer bmpClearOnPanic(&tmp)

	b := make([]byte, 256)

	readPair := func() (byte, byte, error) {
		_, err := io.ReadFull(r, b[:2])
		return b[0], b[1], err
	}

	x, y := 0, height-1
	cap := 8 / bpp

Loop:
	for {
		b1, b2, err := readPair()
		if err != nil {
			C.clear_image(&tmp)
			return err
		}

		if b1 == 0 {
			switch b2 {
			case 0: // End of line
				x, y = 0, y-1
				if y < 0 {
					// We should probably return an error here,
					// but it's safier to just stop decoding
					break Loop
				}
			case 1: // End of file
				break Loop
			case 2:
				dx, dy, err := readPair()
				if err != nil {
					C.clear_image(&tmp)
					return err
				}

				x = imath.Min(x+int(dx), width)
				y -= int(dy)
				if y < 0 {
					break Loop
				}
			default:
				pixelsCount := int(b2)

				n := ((pixelsCount+cap-1)/cap + 1) &^ 1
				if _, err := io.ReadFull(r, b[:n]); err != nil {
					C.clear_image(&tmp)
					return err
				}

				pixelsCount = imath.Min(pixelsCount, width-x)

				if pixelsCount > 0 {
					start := (y*width + x) * 3
					p := imgData[start : start+pixelsCount*3]

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

					x += pixelsCount
				}
			}
		} else {
			pixelsCount := imath.Min(int(b1), width-x)

			if pixelsCount > 0 {
				start := (y*width + x) * 3
				p := imgData[start : start+pixelsCount*3]

				bit := 8 - bpp
				for i := 0; i < len(p); i += 3 {
					pind := (b2 >> bit) & (1<<bpp - 1)

					if bit == 0 {
						bit = 8 - bpp
					} else {
						bit -= bpp
					}

					c := palette[pind]

					p[i+0] = c.R
					p[i+1] = c.G
					p[i+2] = c.B
				}

				x += pixelsCount
			}
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	bmpSetBitDepth(img, len(palette))

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

	defer bmpClearOnPanic(&tmp)

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

// decodeBmpRGB16 reads a 16 bit-per-pixel BMP image from r.
// If topDown is false, the image rows will be read bottom-up.
func (img *Image) decodeBmpRGB16(r io.Reader, width, height int, topDown, bmp565 bool) error {
	tmp, imgData, err := prepareBmpCanvas(width, height, 3)
	if err != nil {
		return err
	}

	defer bmpClearOnPanic(&tmp)

	// Each row is 4-byte aligned.
	b := make([]byte, (2*width+3)&^3)

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
		for i, j := 0, 0; i < len(p); i, j = i+3, j+2 {
			pixel := readUint16(b[j:])

			if bmp565 {
				p[i+0] = uint8((pixel&0xF800)>>11) << 3
				p[i+1] = uint8((pixel&0x7E0)>>5) << 2
			} else {
				p[i+0] = uint8((pixel&0x7C00)>>10) << 3
				p[i+1] = uint8((pixel&0x3E0)>>5) << 3
			}
			p[i+2] = uint8(pixel&0x1F) << 3
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) loadBmp(data []byte, noAlpha bool) error {
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

	// We only support 1 plane and 8, 24 or 32 bits per pixel
	planes, bpp, compression := readUint16(b[26:28]), readUint16(b[28:30]), readUint32(b[30:34])

	if planes != 1 {
		return errBmpUnsupported
	}

	rle := false
	bmp565 := false

	switch {
	case compression == 0:
		// Go ahead
	case compression == 1 && bpp == 8 || compression == 2 && bpp == 4:
		rle = true
	case compression == 3 && infoLen >= infoHeaderLen:
		if infoLen == infoHeaderLen {
			// Color mask is stored after the info header
			if _, err := io.ReadFull(r, b[54:66]); err != nil {
				if err == io.EOF {
					err = io.ErrUnexpectedEOF
				}
				return err
			}
		}

		rmask := readUint32(b[54:58])
		gmask := readUint32(b[58:62])
		bmask := readUint32(b[62:66])
		amask := readUint32(b[66:70])

		switch {
		case bpp == 16 && rmask == 0xF800 && gmask == 0x7E0 && bmask == 0x1F:
			bmp565 = true
		case bpp == 16 && rmask == 0x7C00 && gmask == 0x3E0 && bmask == 0x1F:
			// Go ahead, it's a regular 16 bit image
		case bpp == 32 && rmask == 0xff0000 && gmask == 0xff00 && bmask == 0xff && amask == 0xff000000:
			// Go ahead, it's a regular 32-bit image
		default:
			return errBmpUnsupported
		}
	default:
		return errBmpUnsupported
	}

	var palette []Color
	if bpp <= 8 {
		palColors := readUint32(b[46:50])
		if palColors == 0 {
			palColors = 1 << bpp
		}

		_, err := io.ReadFull(r, b[:palColors*4])
		if err != nil {
			return err
		}

		palette = make([]Color, palColors)
		for i := range palette {
			// BMP images are stored in BGR order rather than RGB order.
			// Every 4th byte is padding.
			palette[i] = Color{b[4*i+2], b[4*i+1], b[4*i+0]}
		}
	}

	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return err
	}

	if rle {
		return img.decodeBmpRLE(r, width, height, int(bpp), palette)
	}

	switch bpp {
	case 1, 2, 4, 8:
		return img.decodeBmpPaletted(r, width, height, int(bpp), palette, topDown)
	case 16:
		return img.decodeBmpRGB16(r, width, height, topDown, bmp565)
	case 24:
		return img.decodeBmpRGB(r, width, height, 3, topDown, true)
	case 32:
		if infoLen >= 70 {
			// Alpha mask is empty, so no alpha here
			noAlpha = readUint32(b[66:70]) == 0
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
