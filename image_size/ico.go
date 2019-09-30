package imageSize

import (
	"encoding/binary"
	"io"
)

func DecodeIcoMeta(r io.Reader) (*Meta, error) {
	var tmp [16]byte

	if _, err := io.ReadFull(r, tmp[:6]); err != nil {
		return nil, err
	}

	count := binary.LittleEndian.Uint16(tmp[4:6])

	width, height := byte(0), byte(0)

	for i := uint16(0); i < count; i++ {
		if _, err := io.ReadFull(r, tmp[:]); err != nil {
			return nil, err
		}

		if tmp[0] > width || tmp[1] > height {
			width = tmp[0]
			height = tmp[1]
		}
	}

	return &Meta{
		Format: "ico",
		Width:  int(width),
		Height: int(height),
	}, nil
}

func init() {
	RegisterFormat("\x00\x00\x01\x00", DecodeIcoMeta)
}
