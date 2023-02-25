package svg

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"

	"github.com/imgproxy/imgproxy/v3/imagedata"
)

var feDropShadowName = []byte("feDropShadow")

const feDropShadowTemplate = `
	<feMerge result="dsin-%[1]s"><feMergeNode %[3]s /></feMerge>
	<feGaussianBlur %[4]s />
	<feOffset %[5]s result="dsof-%[2]s" />
	<feFlood %[6]s />
	<feComposite in2="dsof-%[2]s" operator="in" />
	<feMerge %[7]s>
		<feMergeNode />
		<feMergeNode in="dsin-%[1]s" />
	</feMerge>
`

func Satitize(data *imagedata.ImageData) (*imagedata.ImageData, error) {
	r := bytes.NewReader(data.Data)
	l := xml.NewLexer(parse.NewInput(r))

	buf, cancel := imagedata.BorrowBuffer()

	ignoreTag := 0

	var curTagName string

	for {
		tt, tdata := l.Next()

		if ignoreTag > 0 {
			switch tt {
			case xml.ErrorToken:
				cancel()
				return nil, l.Err()
			case xml.EndTagToken, xml.StartTagCloseVoidToken:
				ignoreTag--
			case xml.StartTagToken:
				ignoreTag++
			}

			continue
		}

		switch tt {
		case xml.ErrorToken:
			if l.Err() != io.EOF {
				cancel()
				return nil, l.Err()
			}

			newData := imagedata.ImageData{
				Data: buf.Bytes(),
				Type: data.Type,
			}
			newData.SetCancel(cancel)

			return &newData, nil
		case xml.StartTagToken:
			curTagName = strings.ToLower(string(l.Text()))

			if curTagName == "script" {
				ignoreTag++
				continue
			}

			buf.Write(tdata)
		case xml.AttributeToken:
			attrName := strings.ToLower(string(l.Text()))

			if _, unsafe := unsafeAttrs[attrName]; unsafe {
				continue
			}

			if curTagName == "use" && (attrName == "href" || attrName == "xlink:href") {
				val := strings.TrimSpace(strings.Trim(string(l.AttrVal()), `"'`))
				if len(val) > 0 && val[0] != '#' {
					continue
				}
			}

			buf.Write(tdata)
		default:
			buf.Write(tdata)
		}
	}
}

func replaceDropShadowNode(l *xml.Lexer, buf *bytes.Buffer) error {
	var (
		inAttrs     strings.Builder
		blurAttrs   strings.Builder
		offsetAttrs strings.Builder
		floodAttrs  strings.Builder
		finalAttrs  strings.Builder
	)

	inID, _ := nanoid.New(8)
	offsetID, _ := nanoid.New(8)

	hasStdDeviation := false
	hasDx := false
	hasDy := false

TOKEN_LOOP:
	for {
		tt, tdata := l.Next()

		switch tt {
		case xml.ErrorToken:
			if l.Err() != io.EOF {
				return l.Err()
			}
			break TOKEN_LOOP
		case xml.EndTagToken, xml.StartTagCloseVoidToken:
			break TOKEN_LOOP
		case xml.AttributeToken:
			switch strings.ToLower(string(l.Text())) {
			case "in":
				inAttrs.Write(tdata)
			case "stddeviation":
				blurAttrs.Write(tdata)
				hasStdDeviation = true
			case "dx":
				offsetAttrs.Write(tdata)
				hasDx = true
			case "dy":
				offsetAttrs.Write(tdata)
				hasDy = true
			case "flood-color", "flood-opacity":
				floodAttrs.Write(tdata)
			default:
				finalAttrs.Write(tdata)
			}
		}
	}

	if !hasStdDeviation {
		blurAttrs.WriteString(` stdDeviation="2"`)
	}

	if !hasDx {
		offsetAttrs.WriteString(` dx="2"`)
	}

	if !hasDy {
		offsetAttrs.WriteString(` dy="2"`)
	}

	fmt.Fprintf(
		buf, feDropShadowTemplate,
		inID, offsetID,
		inAttrs.String(),
		blurAttrs.String(),
		offsetAttrs.String(),
		floodAttrs.String(),
		finalAttrs.String(),
	)

	return nil
}

func FixUnsupported(data *imagedata.ImageData) (*imagedata.ImageData, bool, error) {
	if !bytes.Contains(data.Data, feDropShadowName) {
		return data, false, nil
	}

	r := bytes.NewReader(data.Data)
	l := xml.NewLexer(parse.NewInput(r))

	buf, cancel := imagedata.BorrowBuffer()

	for {
		tt, tdata := l.Next()

		switch tt {
		case xml.ErrorToken:
			if l.Err() != io.EOF {
				cancel()
				return nil, false, l.Err()
			}

			newData := imagedata.ImageData{
				Data: buf.Bytes(),
				Type: data.Type,
			}
			newData.SetCancel(cancel)

			return &newData, true, nil
		case xml.StartTagToken:
			if bytes.Equal(l.Text(), feDropShadowName) {
				if err := replaceDropShadowNode(l, buf); err != nil {
					cancel()
					return nil, false, err
				}
				continue
			}
			buf.Write(tdata)
		default:
			buf.Write(tdata)
		}
	}
}
