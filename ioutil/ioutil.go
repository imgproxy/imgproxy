package ioutil

import "io"

// TryReadFull acts like io.ReadFull with a couple of differences:
//  1. It doesn't return io.ErrUnexpectedEOF if the reader returns less data than requested.
//     Instead, it returns the number of bytes read and the error from the last read operation.
//  2. It always returns the number of bytes read regardless of the error.
func TryReadFull(r io.Reader, b []byte) (n int, err error) {
	var nn int
	toRead := len(b)

	for n < toRead && err == nil {
		nn, err = r.Read(b[n:])
		n += nn
	}

	return n, err
}
