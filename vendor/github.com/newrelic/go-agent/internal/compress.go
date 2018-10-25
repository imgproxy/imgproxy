package internal

import (
	"bytes"
	"compress/zlib"
)

func compress(b []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(b)
	w.Close()

	if nil != err {
		return nil, err
	}

	return &buf, nil
}
