package imagemeta

import (
	"io"
)

func init() {
	// Register fake svg decoder. Since we need this only for type detecting, we can
	// return fake image sizes
	decodeMeta := func(io.Reader) (Meta, error) {
		return &meta{format: "svg", width: 1, height: 1}, nil
	}
	RegisterFormat("<?xml ", decodeMeta)
	RegisterFormat("<svg", decodeMeta)
}
