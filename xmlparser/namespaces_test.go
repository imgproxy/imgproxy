package xmlparser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/xmlparser"
)

const (
	testSVGNS   = "http://www.w3.org/2000/svg"
	testXHTMLNS = "http://www.w3.org/1999/xhtml"
	testXlinkNS = "http://www.w3.org/1999/xlink"
)

func parseNS(t *testing.T, src string) *xmlparser.Document {
	t.Helper()

	doc, err := xmlparser.NewDocument(strings.NewReader(src))
	require.NoError(t, err)

	return doc
}

func firstChild(n *xmlparser.Node) *xmlparser.Node {
	for c := range n.ChildNodes() {
		return c
	}

	return nil
}

func TestNamespaceInheritedAndShadowed(t *testing.T) {
	doc := parseNS(t, `<svg xmlns="`+testSVGNS+`"><g><div xmlns="`+testXHTMLNS+`"><p/></div><rect/></g></svg>`)

	svg := firstChild(&doc.Node)
	require.Equal(t, testSVGNS, svg.NSURI)

	g := firstChild(svg)
	require.Equal(t, testSVGNS, g.NSURI)

	var div, rect *xmlparser.Node
	for c := range g.ChildNodes() {
		switch c.Name {
		case "div":
			div = c
		case "rect":
			rect = c
		}
	}

	require.Equal(t, testXHTMLNS, div.NSURI)
	require.Equal(t, testXHTMLNS, firstChild(div).NSURI) // inherited from div
	require.Equal(t, testSVGNS, rect.NSURI)              // sibling is not affected
}

func TestNamespacePrefixes(t *testing.T) {
	doc := parseNS(t, `<svg xmlns:x="`+testXlinkNS+`" xmlns:s="`+testSVGNS+`">`+
		`<s:a x:href="#a" xml:space="preserve" plain="1" u:attr="2"/></svg>`)

	a := firstChild(firstChild(&doc.Node))
	require.Equal(t, testSVGNS, a.NSURI)

	ns := map[xmlparser.Name]string{}
	for attr := range a.Attrs.Iter() {
		ns[attr.Name] = attr.NSURI
	}

	require.Equal(t, testXlinkNS, ns["x:href"])
	require.Equal(t, xmlparser.NamespaceXML, ns["xml:space"])
	require.Empty(t, ns["plain"])
	require.Empty(t, ns["u:attr"]) // unresolved prefix
}

func TestNamespaceInjectedByEntityInheritsBindings(t *testing.T) {
	doc := parseNS(t, `<!DOCTYPE svg [<!ENTITY e "<rect/>">]>`+
		`<svg xmlns="`+testSVGNS+`"><g xmlns="`+testXHTMLNS+`">&e;</g></svg>`)
	require.NoError(t, doc.ReplaceEntities())

	g := firstChild(firstChild(&doc.Node))
	rect := firstChild(g)

	require.Equal(t, "rect", rect.Name.String())
	require.Equal(t, testXHTMLNS, rect.NSURI)
}
