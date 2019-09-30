package imageSize

import (
	"encoding/binary"
	"io"
)

func icoBestSize(r io.Reader) (width, height byte, page uint16, err error) {
	var tmp [16]byte

	if _, err = io.ReadFull(r, tmp[:6]); err != nil {
		return
	}

	count := binary.LittleEndian.Uint16(tmp[4:6])

	for i := uint16(0); i < count; i++ {
		if _, err = io.ReadFull(r, tmp[:]); err != nil {
			return
		}

		if tmp[0] > width || tmp[1] > height || tmp[0] == 0 || tmp[1] == 0 {
			width = tmp[0]
			height = tmp[1]
			page = i
		}
	}

	return
}

func BestIcoPage(r io.Reader) (int, error) {
	_, _, page, err := icoBestSize(r)
	return int(page), err
}

func DecodeIcoMeta(r io.Reader) (*Meta, error) {
	bwidth, bheight, _, err := icoBestSize(r)
	if err != nil {
		return nil, err
	}

	width := int(bwidth)
	height := int(bheight)

	if width == 0 {
		width = 256
	}

	if height == 0 {
		height = 256
	}

	return &Meta{
		Format: "ico",
		Width:  width,
		Height: height,
	}, nil
}

func init() {
	RegisterFormat("\x00\x00\x01\x00", DecodeIcoMeta)
}
