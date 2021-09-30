package vips

/*
#include "vips.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

func (img *Image) loadIco(data []byte, shrink int, scale float64, pages int) error {
	icoMeta, err := imagemeta.DecodeIcoMeta(bytes.NewReader(data))
	if err != nil {
		return err
	}

	offset := icoMeta.BestImageOffset()
	size := icoMeta.BestImageSize()

	internalData := data[offset : offset+size]

	var internalType imagetype.Type

	meta, err := imagemeta.DecodeMeta(bytes.NewReader(internalData))
	if err != nil {
		// Looks like it's BMP with an incomplete header
		if d, err := imagemeta.FixBmpHeader(internalData); err == nil {
			internalType = imagetype.BMP
			internalData = d
		} else {
			return err
		}
	} else {
		internalType = meta.Format()
	}

	if internalType == imagetype.ICO || !SupportsLoad(internalType) {
		return fmt.Errorf("Can't load %s from ICO", internalType)
	}

	imgdata := imagedata.ImageData{
		Type: internalType,
		Data: internalData,
	}

	return img.Load(&imgdata, shrink, scale, pages)
}

func (img *Image) saveAsIco() (*imagedata.ImageData, error) {
	if img.Width() > 256 || img.Height() > 256 {
		return nil, errors.New("Image dimensions is too big. Max dimension size for ICO is 256")
	}

	var ptr unsafe.Pointer
	imgsize := C.size_t(0)

	defer func() {
		C.g_free_go(&ptr)
	}()

	if C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize, 0, 0, 256) != 0 {
		return nil, Error()
	}

	b := ptrToBytes(ptr, int(imgsize))

	buf := new(bytes.Buffer)
	buf.Grow(22 + int(imgsize))

	// ICONDIR header
	if _, err := buf.Write([]byte{0, 0, 1, 0, 1, 0}); err != nil {
		return nil, err
	}

	// ICONDIRENTRY
	if _, err := buf.Write([]byte{
		byte(img.Width() % 256),
		byte(img.Height() % 256),
	}); err != nil {
		return nil, err
	}
	// Number of colors. Not supported in our case
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// Reserved
	if err := buf.WriteByte(0); err != nil {
		return nil, err
	}
	// Color planes. Always 1 in our case
	if _, err := buf.Write([]byte{1, 0}); err != nil {
		return nil, err
	}
	// Bits per pixel
	if img.HasAlpha() {
		if _, err := buf.Write([]byte{32, 0}); err != nil {
			return nil, err
		}
	} else {
		if _, err := buf.Write([]byte{24, 0}); err != nil {
			return nil, err
		}
	}
	// Image data size
	if err := binary.Write(buf, binary.LittleEndian, uint32(imgsize)); err != nil {
		return nil, err
	}
	// Image data offset. Always 22 in our case
	if _, err := buf.Write([]byte{22, 0, 0, 0}); err != nil {
		return nil, err
	}

	if _, err := buf.Write(b); err != nil {
		return nil, err
	}

	imgdata := imagedata.ImageData{
		Type: imagetype.ICO,
		Data: buf.Bytes(),
	}

	return &imgdata, nil
}
