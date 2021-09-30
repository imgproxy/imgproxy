package imagemeta

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

func DecodeGifMeta(r io.Reader) (Meta, error) {
	var tmp [10]byte

	_, err := io.ReadFull(r, tmp[:])
	if err != nil {
		return nil, err
	}

	return &meta{
		format: imagetype.GIF,
		width:  int(tmp[6]) + int(tmp[7])<<8,
		height: int(tmp[8]) + int(tmp[9])<<8,
	}, nil
}

func init() {
	RegisterFormat("GIF8?a", DecodeGifMeta)
}
