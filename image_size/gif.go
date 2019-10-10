package imagesize

import (
	"io"
)

func DecodeGifMeta(r io.Reader) (*Meta, error) {
	var tmp [10]byte

	_, err := io.ReadFull(r, tmp[:])
	if err != nil {
		return nil, err
	}

	return &Meta{
		Format: "gif",
		Width:  int(tmp[6]) + int(tmp[7])<<8,
		Height: int(tmp[8]) + int(tmp[9])<<8,
	}, nil
}

func init() {
	RegisterFormat("GIF8?a", DecodeGifMeta)
}
