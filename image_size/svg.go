package imagesize

import (
	"io"
)

func init() {
	// Register fake svg decoder. Since we need this only for type detecting, we can
	// return fake image sizes
	decodeMeta := func(io.Reader) (*Meta, error) {
		return &Meta{Format: "svg", Width: 1, Height: 1}, nil
	}
	RegisterFormat("<?xml ", decodeMeta)
	RegisterFormat("<svg", decodeMeta)
}
