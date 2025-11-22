package xmlparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Node struct {
	Parent   *Node
	Name     Name
	Attrs    *Attributes
	Children []any
}

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
				Attrs:  t.Attr,
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

func (n *Node) writeTo(w *bufio.Writer) error {
	if err := w.WriteByte('<'); err != nil {
		return err
	}
	if _, err := w.WriteString(n.Name.String()); err != nil {
		return err
	}

	n.writeAttrsTo(w)

	if len(n.Children) == 0 {
		if _, err := w.WriteString("/>"); err != nil {
			return err
		}
		return nil
	}

	if err := w.WriteByte('>'); err != nil {
		return err
	}

	if err := n.writeChildrenTo(w); err != nil {
		return err
	}

	if _, err := w.WriteString("</"); err != nil {
		return err
	}
	if _, err := w.WriteString(n.Name.String()); err != nil {
		return err
	}
	if err := w.WriteByte('>'); err != nil {
		return err
	}

	return nil
}

func (n *Node) writeAttrsTo(w *bufio.Writer) error {
	for attr := range n.Attrs.Iter() {
		if err := w.WriteByte(' '); err != nil {
			return err
		}
		if _, err := w.WriteString(attr.Name.String()); err != nil {
			return err
		}
		if len(attr.Value) > 0 {
			quote := byte('"')
			if strings.IndexByte(attr.Value, quote) != -1 {
				quote = '\''
			}

			if err := w.WriteByte('='); err != nil {
				return err
			}
			if err := w.WriteByte(quote); err != nil {
				return err
			}
			if _, err := w.WriteString(attr.Value); err != nil {
				return err
			}
			if err := w.WriteByte(quote); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) writeChildrenTo(w *bufio.Writer) error {
	for _, child := range n.Children {
		switch c := child.(type) {
		case *Node:
			if err := c.writeTo(w); err != nil {
				return err
			}

		case *Text:
			if c.CData {
				if _, err := w.WriteString("<![CDATA["); err != nil {
					return err
				}
			}
			if _, err := w.Write(c.Data); err != nil {
				return err
			}
			if c.CData {
				if _, err := w.WriteString("]]>"); err != nil {
					return err
				}
			}

		case *Comment:
			if _, err := w.WriteString("<!--"); err != nil {
				return err
			}
			if _, err := w.Write(c.Data); err != nil {
				return err
			}
			if _, err := w.WriteString("-->"); err != nil {
				return err
			}

		case *Directive:
			if _, err := w.WriteString("<!DOCTYPE"); err != nil {
				return err
			}
			if _, err := w.Write(c.Data); err != nil {
				return err
			}
			if err := w.WriteByte('>'); err != nil {
				return err
			}

		case *ProcInst:
			if _, err := w.WriteString("<?"); err != nil {
				return err
			}
			if _, err := w.Write(c.Target); err != nil {
				return err
			}
			if len(c.Inst) > 0 {
				if !isSpace(c.Inst[0]) {
					if err := w.WriteByte(' '); err != nil {
						return err
					}
				}
				if _, err := w.Write([]byte(c.Inst)); err != nil {
					return err
				}
			}
			if _, err := w.WriteString("?>"); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown child type: %T", c)
		}
	}

	return nil
}

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
