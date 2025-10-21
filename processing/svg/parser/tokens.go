package svgparser

type Name struct {
	Space string
	Local string
}

type Attr struct {
	Name  Name
	Value string
}

type StartElement struct {
	Name Name
	Attr []Attr
}

type EndElement struct {
	Name Name
}

type Directive []byte

func (d Directive) Clone() Directive {
	return Directive(append([]byte(nil), []byte(d)...))
}

type Comment []byte

func (c Comment) Clone() Comment {
	return Comment(append([]byte(nil), []byte(c)...))
}

type ProcInst struct {
	Target string
	Inst   []byte
}

func (p ProcInst) Clone() ProcInst {
	return ProcInst{
		Target: p.Target,
		Inst:   append([]byte(nil), p.Inst...),
	}
}

type Text []byte

func (t Text) Clone() Text {
	return Text(append([]byte(nil), []byte(t)...))
}

type CData []byte

func (c CData) Clone() CData {
	return CData(append([]byte(nil), []byte(c)...))
}
