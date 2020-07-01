package imagemeta

import (
	"bytes"
	"encoding/binary"
	"io"
)

type IcoMeta struct {
	Meta
	offset int
	size   int
}

func (m *IcoMeta) BestImageOffset() int {
	return m.offset
}

func (m *IcoMeta) BestImageSize() int {
	return m.size
}

func icoBestSize(r io.Reader) (width, height byte, offset uint32, size uint32, err error) {
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
			size = binary.LittleEndian.Uint32(tmp[8:12])
			offset = binary.LittleEndian.Uint32(tmp[12:16])
		}
	}

	return
}

func BestIcoPage(r io.Reader) (int, int, error) {
	_, _, offset, size, err := icoBestSize(r)
	return int(offset), int(size), err
}

func DecodeIcoMeta(r io.Reader) (*IcoMeta, error) {
	bwidth, bheight, offset, size, err := icoBestSize(r)
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

	return &IcoMeta{
		Meta: &meta{
			format: "ico",
			width:  width,
			height: height,
		},
		offset: int(offset),
		size:   int(size),
	}, nil
}

func init() {
	RegisterFormat(
		"\x00\x00\x01\x00",
		func(r io.Reader) (Meta, error) { return DecodeIcoMeta(r) },
	)
}

// FixBmpHeader fixes an incomplete header of BMP stored in ICO
func FixBmpHeader(b []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	fileSize := uint32(14 + len(b))

	buf.Grow(int(fileSize))

	buf.Write(bmpMagick)

	if err := binary.Write(buf, binary.LittleEndian, &fileSize); err != nil {
		return nil, err
	}

	reserved := uint32(0)
	if err := binary.Write(buf, binary.LittleEndian, &reserved); err != nil {
		return nil, err
	}

	colorUsed := binary.LittleEndian.Uint32(b[32:36])
	bitCount := binary.LittleEndian.Uint16(b[14:16])

	var pixOffset uint32
	if colorUsed == 0 && bitCount <= 8 {
		pixOffset = 14 + 40 + 4*(1<<bitCount)
	} else {
		pixOffset = 14 + 40 + 4*colorUsed
	}

	if err := binary.Write(buf, binary.LittleEndian, &pixOffset); err != nil {
		return nil, err
	}

	// Write size and width
	buf.Write(b[:8])

	// For some reason ICO stores double height
	height := binary.LittleEndian.Uint32(b[8:12]) / 2
	if err := binary.Write(buf, binary.LittleEndian, &height); err != nil {
		return nil, err
	}

	// Write the rest
	buf.Write(b[12:])

	return buf.Bytes(), nil
}
