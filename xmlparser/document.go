package xmlparser

import (
	"io"
)

type Document struct {
	Node
}

func NewDocument(r io.Reader) (*Document, error) {
	doc := &Document{
		Node: Node{
			// Attributes for document are always empty, but they are exposed,
			// so we need to initialize them to avoid nil pointer dereference.
			Attrs: NewAttributes(),
		},
	}

	if err := doc.readFrom(r); err != nil {
		return nil, err
	}

	return doc, nil
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
