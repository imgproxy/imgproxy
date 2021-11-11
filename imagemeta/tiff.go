package imagemeta

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	tiffLeHeader = []byte("II\x2A\x00")
	tiffBeHeader = []byte("MM\x00\x2A")
)

const (
	tiffDtByte  = 1
	tiffDtShort = 3
	tiffDtLong  = 4

	tiffImageWidth  = 256
	tiffImageLength = 257
)

type tiffReader interface {
	io.Reader
	Discard(n int) (discarded int, err error)
}

func asTiffReader(r io.Reader) tiffReader {
	if rr, ok := r.(tiffReader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

type TiffFormatError string

func (e TiffFormatError) Error() string { return "invalid TIFF format: " + string(e) }

func DecodeTiffMeta(rr io.Reader) (Meta, error) {
	var (
		tmp       [12]byte
		byteOrder binary.ByteOrder
	)

	r := asTiffReader(rr)

	if _, err := io.ReadFull(r, tmp[:8]); err != nil {
		return nil, err
	}

	switch {
	case bytes.Equal(tiffLeHeader, tmp[0:4]):
		byteOrder = binary.LittleEndian
	case bytes.Equal(tiffBeHeader, tmp[0:4]):
		byteOrder = binary.BigEndian
	default:
		return nil, TiffFormatError("malformed header")
	}

	ifdOffset := int(byteOrder.Uint32(tmp[4:8]))

	if _, err := r.Discard(ifdOffset - 8); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(r, tmp[0:2]); err != nil {
		return nil, err
	}
	numItems := int(byteOrder.Uint16(tmp[0:2]))

	var width, height int

	for i := 0; i < numItems; i++ {
		if _, err := io.ReadFull(r, tmp[:]); err != nil {
			return nil, err
		}

		tag := byteOrder.Uint16(tmp[0:2])

		if tag != tiffImageWidth && tag != tiffImageLength {
			continue
		}

		datatype := byteOrder.Uint16(tmp[2:4])

		var value int

		switch datatype {
		case tiffDtByte:
			value = int(tmp[9])
		case tiffDtShort:
			value = int(byteOrder.Uint16(tmp[8:10]))
		case tiffDtLong:
			value = int(byteOrder.Uint32(tmp[8:12]))
		default:
			return nil, TiffFormatError("unsupported IFD entry datatype")
		}

		if tag == tiffImageWidth {
			width = value
		} else {
			height = value
		}

		if width > 0 && height > 0 {
			return &meta{
				format: imagetype.TIFF,
				width:  width,
				height: height,
			}, nil
		}
	}

	return nil, TiffFormatError("image dimensions are not specified")
}

func init() {
	RegisterFormat(string(tiffLeHeader), DecodeTiffMeta)
	RegisterFormat(string(tiffBeHeader), DecodeTiffMeta)
}
