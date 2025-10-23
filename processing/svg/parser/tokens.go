package svgparser

import "bytes"

type Name struct {
	Space string
	Local string
}

type Attr struct {
	Name  Name
	Value string
}

type StartElement struct {
	Name        Name
	Attr        []Attr
	SelfClosing bool
}

type EndElement struct {
	Name Name
}

type Directive []byte

func (d Directive) Clone() Directive {
	return Directive(bytes.Clone([]byte(d)))
}

type Comment []byte

func (c Comment) Clone() Comment {
	return Comment(bytes.Clone([]byte(c)))
}

type ProcInst struct {
	Target []byte
	Inst   []byte
}

func (p ProcInst) Clone() ProcInst {
	return ProcInst{
		Target: bytes.Clone(p.Target),
		Inst:   bytes.Clone(p.Inst),
	}
}

type Text []byte

func (t Text) Clone() Text {
	return Text(bytes.Clone([]byte(t)))
}

type CData []byte

func (c CData) Clone() CData {
	return CData(bytes.Clone([]byte(c)))
}
