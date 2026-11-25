package xmlparser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
)

// Node represents an XML node with a name, attributes, and child tokens.
type Node struct {
	Name Name
	// NSURI is the resolved namespace URI of the element ("" if none or unresolved)
	NSURI    string
	Attrs    *Attributes
	Children []Token

	// bindings is the xmlns bindings in effect for this node's children
	// (the parent's bindings plus this node's own xmlns declarations)
	bindings nsBindings
}

// FilterChildren removes all child nodes that do not satisfy the given predicate function.
func (n *Node) FilterChildren(pred func(child Token) bool) {
	n.Children = slices.DeleteFunc(n.Children, func(child Token) bool {
		return !pred(child)
	})
}

// ChildNodes returns an iterator over children of type *Node.
func (n *Node) ChildNodes() iter.Seq[*Node] {
	return func(yield func(*Node) bool) {
		for _, child := range n.Children {
			cn, ok := child.(*Node)
			if !ok {
				continue
			}
			if !yield(cn) {
				return
			}
		}
	}
}

// FilterChildNodes removes all child nodes of type *Node
// that do not satisfy the given predicate function.
func (n *Node) FilterChildNodes(pred func(child *Node) bool) {
	n.Children = slices.DeleteFunc(n.Children, func(child Token) bool {
		cn, ok := child.(*Node)
		if !ok {
			return false
		}
		return !pred(cn)
	})
}

// WriteTo writes the XML representation of the node and its children to the provided writer.
func (n *Node) WriteTo(w TokenWriter) error {
	if len(n.Name) == 0 {
		// Document node or an unnamed node, write only children
		return n.writeChildrenTo(w)
	}

	selfClosing := len(n.Children) == 0

	se := StartElement{
		Name:        n.Name,
		Attrs:       n.Attrs,
		SelfClosing: selfClosing,
	}

	if err := se.WriteTo(w); err != nil {
		return err
	}

	if selfClosing {
		return nil
	}

	if err := n.writeChildrenTo(w); err != nil {
		return err
	}

	ee := EndElement{
		Name: n.Name,
	}

	if err := ee.WriteTo(w); err != nil {
		return err
	}

	return nil
}

// readFrom reads XML data from the provided reader
// and populates the node and its children accordingly.
func (n *Node) readFrom(r io.Reader) error {
	dec := NewDecoder(r)
	defer dec.Close()

	return n.readFromDecoder(dec)
}

// readFromDecoder reads XML data from the provided decoder
// and populates the node and its children accordingly.
func (n *Node) readFromDecoder(dec *Decoder) error {
	for {
		// Read raw token so decoder doesn't mess with attributes and namespaces.
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case *StartElement:
			// An element is opened, create a node for it
			bindings := n.bindings.with(t.Attrs)

			for attr := range t.Attrs.Iter() {
				attr.NSURI = bindings.attrNS(attr.Name)
			}

			nn := &Node{
				Name:     t.Name,
				NSURI:    bindings.elementNS(t.Name),
				Attrs:    t.Attrs,
				bindings: bindings,
			}
			// Append the node to the current node's children and make it current
			n.Children = append(n.Children, nn)

			if !t.SelfClosing {
				if err := nn.readFromDecoder(dec); err != nil {
					return err
				}
			}

		case *EndElement:
			// If the current node has no name, then we are at the root,
			// which can't be closed.
			if len(n.Name) == 0 {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> while no elements are opened",
					t.Name,
				)
			}
			// Closing tag name should match opened node name (which is current)
			if n.Name != t.Name {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> for opened <%s> element",
					t.Name,
					n.Name,
				)
			}
			// The node is closed, return to its parent
			return nil

		case *Text:
			n.Children = append(n.Children, t.Clone())

		case *Directive:
			n.Children = append(n.Children, t.Clone())

		case *Comment:
			n.Children = append(n.Children, t.Clone())

		case *ProcInst:
			n.Children = append(n.Children, t.Clone())
		}
	}

	return nil
}

// writeChildrenTo writes all child tokens of the node to the provided writer.
func (n *Node) writeChildrenTo(w TokenWriter) error {
	for _, child := range n.Children {
		if err := child.WriteTo(w); err != nil {
			return err
		}
	}

	return nil
}

// replaceEntities replaces XML entities in the node
// according to the provided entity map.
func (n *Node) replaceEntities(em map[string][]byte) error {
	// Replace in attributes
	for attr := range n.Attrs.Iter() {
		attr.Value = ReplaceEntitiesString(attr.Value, em)
	}

	// newChildren contains the updated list of child tokens after entity replacement
	// and re-parsing of any embedded XML.
	// If newChldren is nil, it means the original children slice is still valid
	// and no re-parsing has occurred.
	var newChldren []Token

	// Replace in children.
	// Only Text nodes are processed, other Node children
	// are processed recursively.
	for i, child := range n.Children {
		switch c := child.(type) {
		case *Node:
			if err := c.replaceEntities(em); err != nil {
				return err
			}

		case *Text:
			// Skip entity replacement for CDATA sections
			if c.CData {
				break
			}

			c.Data = ReplaceEntitiesBytes(c.Data, em)

			// Check if the new text contains any XML markup
			if !slices.Contains(c.Data, '<') {
				break
			}

			// It seems like the text contains XML markup, so we parse it into a new node.
			// The new nodes inherit the xmlns bindings in effect at this node
			nn := &Node{bindings: n.bindings}
			if err := nn.readFrom(bytes.NewReader(c.Data)); err != nil {
				return err
			}

			// If we have no newChildren yet, initialize the slice and copy previous children into it
			if newChldren == nil {
				newChldren = make([]Token, 0, len(n.Children)+len(nn.Children))
				newChldren = append(newChldren, n.Children[:i]...)
			}

			// Append the newly parsed node's children to the newChldren slice
			newChldren = append(newChldren, nn.Children...)

			// Skip adding the original text node since it has been replaced by parsed children.
			continue
		}

		// If we have started building a new children slice, append the current child to it.
		if newChldren != nil {
			newChldren = append(newChldren, child)
		}
	}

	// If we have built a new children slice, replace the node's children with it.
	if newChldren != nil {
		n.Children = newChldren
	}

	return nil
}
