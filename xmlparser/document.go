package xmlparser

import (
	"bytes"
	"io"
)

type Document struct {
	Node
}

func NewDocument(r io.Reader) (*Document, error) {
	doc := &Document{
		// Attributes for document are always empty, but they are exposed,
		// so we need to initialize them to avoid nil pointer dereference.
		Attrs: NewAttributes(),
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
// "Billion Laughs" XML entity expansion attack. Since consumers expand
// nested entities recursively, a declaration that references another declared
// entity can't be represented by a single pass, so it is rejected with an error.
func (doc *Document) ReplaceEntities() error {
	em := ParseEntityMap(doc.Children)
	if em == nil {
		return nil
	}

	for name, value := range em {
		if !bytes.Equal(ReplaceEntitiesBytes(value, em), value) {
			return newSyntaxError("entity %q references another entity", name)
		}
	}

	return doc.replaceEntities(em)
}
