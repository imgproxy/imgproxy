package xmlparser

import "io"

type writeCounter struct {
	Writer io.Writer
	Count  int64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n, err := wc.Writer.Write(p)
	wc.Count += int64(n)
	return n, err
}
