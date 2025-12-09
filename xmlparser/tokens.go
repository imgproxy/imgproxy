package xmlparser

import (
	"bytes"
	"strings"
)

// Token represents a token that can be written to a TokenWriter.
type Token interface {
	WriteTo(w TokenWriter) error
}

// StartElement represents a start element token (e.g., <tag>).
// SelfClosing indicates whether the element is self-closing (e.g., <tag/>).
type StartElement struct {
	Name        Name
	Attrs       *Attributes
	SelfClosing bool
}

// WriteTo writes the XML representation of the start element to the provided writer.
func (e *StartElement) WriteTo(w TokenWriter) error {
	if err := w.WriteByte('<'); err != nil {
		return err
	}
	if _, err := w.WriteString(e.Name.String()); err != nil {
		return err
	}

	e.writeAttrsTo(w)

	if e.SelfClosing {
		if _, err := w.WriteString("/>"); err != nil {
			return err
		}
		return nil
	}

	if err := w.WriteByte('>'); err != nil {
		return err
	}

	return nil
}

// writeAttrsTo writes the attributes of the start element to the provided writer.
func (e *StartElement) writeAttrsTo(w TokenWriter) error {
	for attr := range e.Attrs.Iter() {
		if err := w.WriteByte(' '); err != nil {
			return err
		}

		quote := byte('"')
		if strings.IndexByte(attr.Value, quote) != -1 {
			quote = '\''
		}

		if _, err := w.WriteString(attr.Name.String()); err != nil {
			return err
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

	return nil
}

// EndElement represents an end element token (e.g., </tag>).
type EndElement struct {
	Name Name
}

// WriteTo writes the XML representation of the end element to the provided writer.
func (e *EndElement) WriteTo(w TokenWriter) error {
	if _, err := w.WriteString("</"); err != nil {
		return err
	}
	if _, err := w.WriteString(e.Name.String()); err != nil {
		return err
	}
	if err := w.WriteByte('>'); err != nil {
		return err
	}
	return nil
}

// Directive represents a DOCTYPE directive token.
// When obtained from [Decoder], the bytes in the Data field are actual only
// until the next call to [Decoder.Token]. To retain the data, use the [Directive.Clone] method.
type Directive struct {
	Data []byte
}

// Clone creates a deep copy of the Directive.
func (d *Directive) Clone() *Directive {
	return &Directive{bytes.Clone(d.Data)}
}

// WriteTo writes the XML representation of the directive to the provided writer.
func (d *Directive) WriteTo(w TokenWriter) error {
	if _, err := w.WriteString("<!DOCTYPE"); err != nil {
		return err
	}
	if _, err := w.Write(d.Data); err != nil {
		return err
	}
	if err := w.WriteByte('>'); err != nil {
		return err
	}
	return nil
}

// Comment represents a comment token (e.g., <!-- comment -->).
// When obtained from [Decoder], the bytes in the Data field are actual only
// until the next call to [Decoder.Token]. To retain the data, use the [Comment.Clone] method.
type Comment struct {
	Data []byte
}

// Clone creates a deep copy of the Comment.
func (c *Comment) Clone() *Comment {
	return &Comment{bytes.Clone(c.Data)}
}

// WriteTo writes the XML representation of the comment to the provided writer.
func (c *Comment) WriteTo(w TokenWriter) error {
	if _, err := w.WriteString("<!--"); err != nil {
		return err
	}
	if _, err := w.Write(c.Data); err != nil {
		return err
	}
	if _, err := w.WriteString("-->"); err != nil {
		return err
	}
	return nil
}

// ProcInst represents a processing instruction token (e.g., <?xml version="1.0"?>).
// When obtained from [Decoder], the bytes in the Target and Inst fields are actual only
// until the next call to [Decoder.Token]. To retain the data, use the [ProcInst.Clone] method.
type ProcInst struct {
	Target []byte
	Inst   []byte
}

// Clone creates a deep copy of the ProcInst.
func (p *ProcInst) Clone() *ProcInst {
	return &ProcInst{
		Target: bytes.Clone(p.Target),
		Inst:   bytes.Clone(p.Inst),
	}
}

// WriteTo writes the XML representation of the processing instruction to the provided writer.
func (p *ProcInst) WriteTo(w TokenWriter) error {
	if _, err := w.WriteString("<?"); err != nil {
		return err
	}
	if _, err := w.Write(p.Target); err != nil {
		return err
	}
	if len(p.Inst) > 0 {
		if !isSpace(p.Inst[0]) {
			if err := w.WriteByte(' '); err != nil {
				return err
			}
		}
		if _, err := w.Write(p.Inst); err != nil {
			return err
		}
	}
	if _, err := w.WriteString("?>"); err != nil {
		return err
	}
	return nil
}

// Text represents a text token.
// CData indicates whether the text is a CDATA section.
// When obtained from [Decoder], the bytes in the Data field are actual only
// until the next call to [Decoder.Token]. To retain the data, use the [Text.Clone] method.
type Text struct {
	Data  []byte
	CData bool
}

// Clone creates a deep copy of the Text.
func (t *Text) Clone() *Text {
	return &Text{bytes.Clone(t.Data), t.CData}
}

// WriteTo writes the XML representation of the text to the provided writer.
func (t *Text) WriteTo(w TokenWriter) error {
	if t.CData {
		if _, err := w.WriteString("<![CDATA["); err != nil {
			return err
		}
	}
	if _, err := w.Write(t.Data); err != nil {
		return err
	}
	if t.CData {
		if _, err := w.WriteString("]]>"); err != nil {
			return err
		}
	}
	return nil
}
