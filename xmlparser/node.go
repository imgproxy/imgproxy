package xmlparser

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"slices"
)

// Node represents an XML node with a name, attributes, and child tokens.
type Node struct {
	Parent   *Node
	Name     Name
	Attrs    *Attributes
	Children []Token
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

// readFrom reads XML data from the provided reader
// and populates the node and its children accordingly.
func (n *Node) readFrom(r io.Reader) error {
	if n.Parent != nil {
		return errors.New("cannot read child node")
	}

	dec := NewDecoder(r)
	defer dec.Close()

	curNode := n

	for {
		// Read raw token so decoder doesn't mess with attributes and namespaces.
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case *StartElement:
			// An element is opened, create a node for it
			el := &Node{
				Parent: curNode,
				Name:   t.Name,
				Attrs:  t.Attrs,
			}
			// Append the node to the current node's children and make it current
			curNode.Children = append(curNode.Children, el)

			if !t.SelfClosing {
				curNode = el
			}

		case *EndElement:
			// If the current node has no parent, then we are at the root,
			// which can't be closed.
			if curNode.Parent == nil {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> while no elements are opened",
					t.Name,
				)
			}
			// Closing tag name should match opened node name (which is current)
			if curNode.Name != t.Name {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> for opened <%s> element",
					t.Name,
					curNode.Name,
				)
			}
			// The node is closed, return to its parent
			curNode = curNode.Parent

		case *Text:
			curNode.Children = append(curNode.Children, t.Clone())

		case *Directive:
			curNode.Children = append(curNode.Children, t.Clone())

		case *Comment:
			curNode.Children = append(curNode.Children, t.Clone())

		case *ProcInst:
			curNode.Children = append(curNode.Children, t.Clone())
		}
	}

	return nil
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
func (n *Node) replaceEntities(em map[string][]byte) {
	// Replace in attributes
	for attr := range n.Attrs.Iter() {
		attr.Value = replaceEntitiesString(attr.Value, em)
	}

	// Replace in children.
	// Only Text nodes are processed, other Node children
	// are processed recursively.
	for _, child := range n.Children {
		switch c := child.(type) {
		case *Node:
			c.replaceEntities(em)

		case *Text:
			if !c.CData {
				c.Data = replaceEntitiesBytes(c.Data, em)
			}
		}
	}
}
