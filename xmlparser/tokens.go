package xmlparser

import "bytes"

type StartElement struct {
	Name        Name
	Attr        *Attributes
	SelfClosing bool
}

type EndElement struct {
	Name Name
}

type Directive struct {
	Data []byte
}

func (d *Directive) Clone() *Directive {
	return &Directive{bytes.Clone(d.Data)}
}

type Comment struct {
	Data []byte
}

func (c *Comment) Clone() *Comment {
	return &Comment{bytes.Clone(c.Data)}
}

type ProcInst struct {
	Target []byte
	Inst   []byte
}

func (p *ProcInst) Clone() *ProcInst {
	return &ProcInst{
		Target: bytes.Clone(p.Target),
		Inst:   bytes.Clone(p.Inst),
	}
}

type Text struct {
	Data  []byte
	CData bool
}

func (t *Text) Clone() *Text {
	return &Text{bytes.Clone(t.Data), t.CData}
}
