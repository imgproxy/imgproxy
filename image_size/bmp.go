package imageSize

import (
	"bytes"
	"encoding/binary"
	"io"
)

var bmpMagick = []byte("BM")

type BmpFormatError string

func (e BmpFormatError) Error() string { return "invalid BMP format: " + string(e) }

func DecodeBmpMeta(r io.Reader) (*Meta, error) {
	var tmp [26]byte

	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}

	if !bytes.Equal(tmp[:2], bmpMagick) {
		return nil, BmpFormatError("malformed header")
	}

	infoSize := binary.LittleEndian.Uint32(tmp[14:18])

	var width, height int

	if infoSize >= 40 {
		width = int(binary.LittleEndian.Uint32(tmp[18:22]))
		height = int(binary.LittleEndian.Uint32(tmp[22:26]))
	} else {
		// CORE
		width = int(binary.LittleEndian.Uint16(tmp[18:20]))
		height = int(binary.LittleEndian.Uint16(tmp[20:22]))
	}

	return &Meta{
		Format: "bmp",
		Width:  width,
		Height: height,
	}, nil
}

func init() {
	RegisterFormat(string(bmpMagick), DecodeBmpMeta)
}
