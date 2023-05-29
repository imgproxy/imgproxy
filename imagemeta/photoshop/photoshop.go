package photoshop

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ps3Header      = []byte("Photoshop 3.0\x00")
	ps3BlockHeader = []byte("8BIM")

	errInvalidPS3Header = errors.New("invalid Photoshop 3.0 header")
)

const (
	IptcKey       = "\x04\x04"
	ResolutionKey = "\x03\xed"
)

type PhotoshopMap map[string][]byte

func Parse(data []byte, m PhotoshopMap) error {
	buf := bytes.NewBuffer(data)

	if !bytes.Equal(buf.Next(14), ps3Header) {
		return errInvalidPS3Header
	}

	// Read blocks
	// Minimal block size is 12 (4 blockHeader + 2 resoureceID + 2 name + 4 blockSize)
	for buf.Len() >= 12 {
		if !bytes.Equal(buf.Bytes()[:4], ps3BlockHeader) {
			buf.Next(1)
			continue
		}

		// Skip block header
		buf.Next(4)

		resoureceID := buf.Next(2)

		// Skip name
		// Name is zero terminated string padded to even
		for buf.Len() > 0 && buf.Next(2)[1] != 0 {
		}

		if buf.Len() < 4 {
			break
		}

		blockSize := int(binary.BigEndian.Uint32(buf.Next(4)))

		if buf.Len() < blockSize {
			break
		}
		blockData := buf.Next(blockSize)

		m[string(resoureceID)] = blockData
	}

	return nil
}

func (m PhotoshopMap) Dump() []byte {
	buf := new(bytes.Buffer)
	buf.Grow(26)

	buf.Write(ps3Header)

	for id, data := range m {
		if len(data) == 0 {
			continue
		}

		buf.Write(ps3BlockHeader)
		buf.WriteString(id)
		// Write empty name
		buf.Write([]byte{0, 0})
		binary.Write(buf, binary.BigEndian, uint32(len(data)))
		buf.Write(data)
	}

	return buf.Bytes()
}
