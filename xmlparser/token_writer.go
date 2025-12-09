package xmlparser

import "io"

// TokenWriter defines an interface with methods used to write XML tokens.
type TokenWriter interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}
