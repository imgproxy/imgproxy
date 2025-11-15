package xmlparser

import (
	"bufio"
	"io"
)

type Document struct {
	Node
}

func NewDocument(r io.Reader) (*Document, error) {
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

// ReplaceEntities replaces XML entities in the document
// according to the entity declarations in the DOCTYPE.
// It modifies the document in place.
//
// Entities are replaced only once to avoid attacks like the
// "Billion Laughs" XML entity expansion attack.
func (doc *Document) ReplaceEntities() {
	if em := parseEntityMap(doc.Children); em != nil {
		doc.replaceEntities(em)
	}
}
