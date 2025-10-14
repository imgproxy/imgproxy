package svgparser

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

type Attr = xml.Attr

type Node struct {
	Parent   *Node
	Name     xml.Name
	Attrs    []Attr
	Children []any
}

func (n *Node) readFrom(r io.ReadSeeker) error {
	if n.Parent != nil {
		return errors.New("cannot read child node")
	}

	dec := xml.NewDecoder(r)
	dec.Strict = false
	dec.CharsetReader = charset.NewReaderLabel

	curNode := n

	for {
		// Save the current position to know where to read raw CData from.
		pos := dec.InputOffset()

		// Read raw token so decoder doesn't mess with attributes and namespaces.
		tok, err := dec.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			// An element is opened, create a node for it
			el := &Node{
				Parent: curNode,
				Name:   t.Name,
				Attrs:  t.Attr,
			}
			// Append the node to the current node's children and make it current
			curNode.Children = append(curNode.Children, el)
			curNode = el

		case xml.EndElement:
			// If the current node has no parent, then we are at the root,
			// which can't be closed.
			if curNode.Parent == nil {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> while no elements are opened",
					fullName(t.Name),
				)
			}
			// Closing tag name should match opened node name (which is current)
			if curNode.Name.Local != t.Name.Local || curNode.Name.Space != t.Name.Space {
				return fmt.Errorf(
					"malformed XML: unexpected closing tag </%s> for opened <%s> element",
					fullName(t.Name),
					fullName(curNode.Name),
				)
			}
			// The node is closed, return to its parent
			curNode = curNode.Parent

		case xml.CharData:
			// We want CData as is, so read it raw
			cdata, err := readRawCData(r, pos, dec.InputOffset()-pos)
			if err != nil {
				return err
			}

			curNode.Children = append(curNode.Children, cdata)

		case xml.Directive:
			curNode.Children = append(curNode.Children, t.Copy())

		case xml.Comment:
			curNode.Children = append(curNode.Children, t.Copy())

		case xml.ProcInst:
			curNode.Children = append(curNode.Children, t.Copy())
		}
	}

	return nil
}

func (n *Node) writeTo(w *bufio.Writer) error {
	if err := w.WriteByte('<'); err != nil {
		return err
	}
	if err := writeFullName(w, n.Name); err != nil {
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
	if err := writeFullName(w, n.Name); err != nil {
		return err
	}
	if err := w.WriteByte('>'); err != nil {
		return err
	}

	return nil
}

func (n *Node) writeAttrsTo(w *bufio.Writer) error {
	for _, attr := range n.Attrs {
		if err := w.WriteByte(' '); err != nil {
			return err
		}
		if err := writeFullName(w, attr.Name); err != nil {
			return err
		}
		if _, err := w.WriteString(`="`); err != nil {
			return err
		}
		if len(attr.Value) > 2 && attr.Value[0] == '&' && attr.Value[len(attr.Value)-1] == ';' {
			// Attribute value is an entity, write it as is
			if _, err := w.WriteString(attr.Value); err != nil {
				return err
			}
		} else {
			// Escape the attribute value
			if err := escapeString(w, attr.Value); err != nil {
				return err
			}
		}
		if err := w.WriteByte('"'); err != nil {
			return err
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

		case CData:
			if _, err := w.Write([]byte(c)); err != nil {
				return err
			}

		case Comment:
			if _, err := w.WriteString("<!--"); err != nil {
				return err
			}
			if _, err := w.Write([]byte(c)); err != nil {
				return err
			}
			if _, err := w.WriteString("-->"); err != nil {
				return err
			}

		case Directive:
			if _, err := w.WriteString("<!"); err != nil {
				return err
			}
			if _, err := w.Write([]byte(c)); err != nil {
				return err
			}
			if err := w.WriteByte('>'); err != nil {
				return err
			}

		case ProcInst:
			if _, err := w.WriteString("<?"); err != nil {
				return err
			}
			if _, err := w.WriteString(c.Target); err != nil {
				return err
			}
			if len(c.Inst) > 0 {
				if err := w.WriteByte(' '); err != nil {
					return err
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

func fullName(name xml.Name) string {
	if len(name.Space) == 0 {
		return name.Local
	}
	return name.Space + ":" + name.Local
}

func writeFullName(w *bufio.Writer, name xml.Name) error {
	if len(name.Space) > 0 {
		if _, err := w.WriteString(name.Space); err != nil {
			return err
		}
		if err := w.WriteByte(':'); err != nil {
			return err
		}
	}

	if _, err := w.WriteString(name.Local); err != nil {
		return err
	}

	return nil
}
