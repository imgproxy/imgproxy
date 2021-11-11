package imagemeta

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var bmpMagick = []byte("BM")

type BmpFormatError string

func (e BmpFormatError) Error() string { return "invalid BMP format: " + string(e) }

func DecodeBmpMeta(r io.Reader) (Meta, error) {
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
		height = int(int32(binary.LittleEndian.Uint32(tmp[22:26])))
	} else {
		// CORE
		width = int(binary.LittleEndian.Uint16(tmp[18:20]))
		height = int(int16(binary.LittleEndian.Uint16(tmp[20:22])))
	}

	// height can be negative in Windows bitmaps
	if height < 0 {
		height = -height
	}

	return &meta{
		format: imagetype.BMP,
		width:  width,
		height: height,
	}, nil
}

func init() {
	RegisterFormat(string(bmpMagick), DecodeBmpMeta)
}
