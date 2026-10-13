package xmlparser_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/xmlparser"
)

func TestDocumentParsing(t *testing.T) {
	testImagesPath, err := filepath.Abs("../testdata/test-images/svg-test-suite")
	require.NoError(t, err)

	err = filepath.Walk(testImagesPath, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-SVG files
		if filepath.Ext(path) != ".svg" {
			return nil
		}

		// Open the SVG file
		file, err := os.Open(path)
		require.NoError(t, err)
		defer file.Close()

		// Parse the document
		doc1, err := xmlparser.NewDocument(file)
		require.NoError(t, err, "Failed to parse SVG: %s", path)

		// Write the document back to a buffer
		buf := new(bytes.Buffer)
		err = doc1.WriteTo(buf)
		require.NoError(t, err, "Failed to write SVG: %s", path)

		// Parse the document again from the written buffer
		doc2, err := xmlparser.NewDocument(buf)
		require.NoError(t, err, "Failed to re-parse SVG: %s", path)

		// Ensure that the two documents are equivalent
		require.Equal(t, doc1, doc2, "Documents do not match after re-parsing: %s", path)

		return nil
	})

	require.NoError(t, err)
}

func TestEntityReplacement(t *testing.T) {
	svgData := []byte(`<?xml version="1.0"?>
<!DOCTYPE svg [
	<!ENTITY myEntity1 "EntityValue1">
	<!ENTITY myEntity2 'EntityValue2'>
	<!ENTITY myEntity3 '<g>EntityValue3</g>'>
]>
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
	<text x="10" y="20" arg1="Value with &myEntity1;" arg2='&myEntity2;'>
		&myEntity1; and &myEntity2;
	</text>
	&myEntity3;
	<style>
		<![CDATA[
			.textClass { content: "&myEntity1;"; }
		]]>
	</style>
</svg>`)

	expectedData := []byte(`<?xml version="1.0"?>
<!DOCTYPE svg [
	<!ENTITY myEntity1 "EntityValue1">
	<!ENTITY myEntity2 'EntityValue2'>
	<!ENTITY myEntity3 '<g>EntityValue3</g>'>
]>
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
	<text x="10" y="20" arg1="Value with EntityValue1" arg2="EntityValue2">
		EntityValue1 and EntityValue2
	</text>
	<g>EntityValue3</g>
	<style>
		<![CDATA[
			.textClass { content: "&myEntity1;"; }
		]]>
	</style>
</svg>`)

	doc, err := xmlparser.NewDocument(bytes.NewReader(svgData))
	require.NoError(t, err)

	doc.ReplaceEntities()

	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	require.NoError(t, err)

	require.Equal(t, string(expectedData), buf.String())

	// Check that the <g> element from the entity replacement is re-parsed as a proper node
	var svgNode, gNode *xmlparser.Node

	for n := range doc.ChildNodes() {
		if n.Name == "svg" {
			svgNode = n
			break
		}
	}
	require.NotNil(t, svgNode, "SVG node not found")

	for n := range svgNode.ChildNodes() {
		if n.Name == "g" {
			gNode = n
			break
		}
	}
	require.NotNil(t, gNode, "G node not found after entity replacement")

	textNode, ok := gNode.Children[0].(*xmlparser.Text)
	require.True(t, ok, "First child of <g> node is not a text node")
	require.Equal(t, "EntityValue3", string(textNode.Data),
		"Text node inside <g> does not have the expected content")
}

func BenchmarkDocumentParsing(b *testing.B) {
	testImagesPath, err := filepath.Abs("../testdata/test-images/svg-test-suite")
	if err != nil {
		b.Fatal(err)
	}

	samples := [][]byte{}

	err = filepath.Walk(testImagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			b.Fatal(err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-SVG files
		if filepath.Ext(path) != ".svg" {
			return nil
		}

		// Read SVG file
		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}

		samples = append(samples, data)
		return nil
	})

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		for _, sample := range samples {
			d, err := xmlparser.NewDocument(bytes.NewReader(sample))
			if err != nil {
				b.Fatal(err)
			}
			if err = d.ReplaceEntities(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func TestDocumentReplaceEntitiesRejectsNested(t *testing.T) {
	doc, err := xmlparser.NewDocument(strings.NewReader(
		`<!DOCTYPE svg [<!ENTITY a "x"><!ENTITY b "&a;y">]><svg a="&b;"/>`,
	))
	require.NoError(t, err)

	require.Error(t, doc.ReplaceEntities())
}

// billionLaughs builds the classic "Billion Laughs" payload: every entity
// references the previous one ten times.
func billionLaughs() string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE svg [<!ENTITY lol "lol">`)

	prev := "lol"
	for i := 1; i <= 9; i++ {
		name := fmt.Sprintf("lol%d", i)
		fmt.Fprintf(&sb, `<!ENTITY %s "%s">`, name, strings.Repeat("&"+prev+";", 10))
		prev = name
	}

	fmt.Fprintf(&sb, `]><svg><text>&%s;</text></svg>`, prev)

	return sb.String()
}

func TestDocumentReplaceEntitiesBillionLaughs(t *testing.T) {
	doc, err := xmlparser.NewDocument(strings.NewReader(billionLaughs()))
	require.NoError(t, err)

	require.Error(t, doc.ReplaceEntities())

	// The document must not be expanded
	var buf bytes.Buffer
	require.NoError(t, doc.WriteTo(&buf))
	require.Less(t, buf.Len(), 4096)
}
