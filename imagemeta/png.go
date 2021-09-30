package imagemeta

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var pngMagick = []byte("\x89PNG\r\n\x1a\n")

type PngFormatError string

func (e PngFormatError) Error() string { return "invalid PNG format: " + string(e) }

func DecodePngMeta(r io.Reader) (Meta, error) {
	var tmp [16]byte

	if _, err := io.ReadFull(r, tmp[:8]); err != nil {
		return nil, err
	}

	if !bytes.Equal(pngMagick, tmp[:8]) {
		return nil, PngFormatError("not a PNG image")
	}

	if _, err := io.ReadFull(r, tmp[:]); err != nil {
		return nil, err
	}

	return &meta{
		format: imagetype.PNG,
		width:  int(binary.BigEndian.Uint32(tmp[8:12])),
		height: int(binary.BigEndian.Uint32(tmp[12:16])),
	}, nil
}

func init() {
	RegisterFormat(string(pngMagick), DecodePngMeta)
}
