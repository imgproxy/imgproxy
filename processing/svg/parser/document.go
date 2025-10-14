package svgparser

import (
	"bufio"
	"io"
)

type Document struct {
	Node
}

func NewDocument(r io.ReadSeeker) (*Document, error) {
	doc := &Document{}
	if err := doc.readFrom(r); err != nil {
		return nil, err
	}

	return doc, nil
}

func (doc *Document) WriteTo(w io.Writer) (int64, error) {
	wc := writeCounter{Writer: w}
	bw := bufio.NewWriter(&wc)
	if err := doc.writeChildrenTo(bw); err != nil {
		return 0, err
	}
	if err := bw.Flush(); err != nil {
		return 0, err
	}
	return wc.Count, nil
}
