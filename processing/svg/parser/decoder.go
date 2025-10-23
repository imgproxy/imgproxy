package svgparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sync"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

var bufreaderPool = sync.Pool{
	New: func() any {
		return bufio.NewReader(nil)
	},
}

var (
	endTagStart   = []byte("</")
	procInstStart = []byte("<?")
	procInstEnd   = []byte("?>")
	commentStart  = []byte("<!--")
	commentEnd    = []byte("-->")
	cdataStart    = []byte("<![CDATA[")
	cdataEnd      = []byte("]]>")
	doctypeStart  = []byte("<!DOCTYPE")

	targetXML = []byte("xml")

	nameSep = []byte{':'}
)

var encodingRE = regexp.MustCompile(`(?s)^(.*encoding=)("[^"]+?"|'[^']+?')(.*)$`)

type Decoder struct {
	r   *bufio.Reader
	buf *buffer
	err error

	line int
}

func NewDecoder(r io.Reader) *Decoder {
	dec := Decoder{buf: newBuffer(), line: 1}

	dec.setReader(r)
	dec.checkBOM()

	return &dec
}

func (d *Decoder) Close() error {
	d.r.Reset(nil)
	bufreaderPool.Put(d.r)
	d.r = nil

	d.buf.Free()
	d.buf = nil

	return d.err
}

// setReader sets a new reader for the decoder, wrapping it in a bufio.Reader from the pool.
func (d *Decoder) setReader(r io.Reader) {
	d.r = bufreaderPool.Get().(*bufio.Reader)
	d.r.Reset(r)
}

// setEncoding recreates the reader with the specified encoding.
func (d *Decoder) setEncoding(encoding string) bool {
	// Recreate the reader with the specified encoding.
	// We are going to wrap bufio.Reader with bufio.Reader again,
	// but non-UTF-8 encodings are rare, so this should be fine.
	newr, err := charset.NewReaderLabel(encoding, d.r)
	if err != nil {
		d.err = fmt.Errorf("can't create reader for encoding %q: %w", encoding, err)
		return false
	}

	d.setReader(newr)

	return true
}

// checkBOM checks for a Byte Order Mark (BOM) at the start of the stream
// and adjusts the reader accordingly.
func (d *Decoder) checkBOM() {
	b, err := d.r.Peek(4)
	if err != nil {
		return
	}

	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		// It's UTF-8 BOM, nothing to do but skip it.
		d.discard(3)
	case bytes.HasPrefix(b, []byte{0x00, 0x00, 0xFE, 0xFF}):
		// UTF-32 BE
		d.discard(4)
		d.setEncoding("utf-32be")
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE, 0x00, 0x00}):
		// UTF-32 LE
		d.discard(4)
		d.setEncoding("utf-32le")
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		// UTF-16 BE
		d.discard(2)
		d.setEncoding("utf-16be")
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		// UTF-16 LE
		d.discard(2)
		d.setEncoding("utf-16le")
	}
}

func (d *Decoder) Token() (any, error) {
	if d.err != nil {
		return nil, d.err
	}

	b, ok := d.peek(1)
	if !ok {
		return nil, d.err
	}

	// If the next byte is not '<', this is plain text.
	if b[0] != '<' {
		text := d.readText()
		if d.err != nil && d.err != io.EOF {
			return nil, d.err
		}
		return Text(text), nil
	}

	// Read the next 3 byte to determine the type of the tag.
	// Any possible valid tag has at least 2 bytes after '<'.
	b, ok = d.mustPeek(3)
	if !ok {
		return nil, d.err
	}

	switch {
	case bytes.HasPrefix(b, endTagStart):
		// End element
		name, nameOk := d.readEndTag()
		if !nameOk {
			return nil, d.err
		}
		return EndElement{Name: name}, nil

	case bytes.HasPrefix(b, procInstStart):
		// Processing instruction
		target, data, procInstOk := d.readProcInst()
		if !procInstOk {
			return nil, d.err
		}
		return ProcInst{
			Target: target,
			Inst:   data,
		}, nil

	// Check only the first 3 bytes for comment,
	// we will check the full sequence in readComment.
	case bytes.HasPrefix(b, commentStart[:3]):
		data, commentOk := d.readComment()
		if !commentOk {
			return nil, d.err
		}
		return Comment(data), nil

	// Check only the first 3 bytes for CDATA,
	// we will check the full sequence in readCData.
	case bytes.HasPrefix(b, cdataStart[:3]):
		data, cdataOk := d.readCData()
		if !cdataOk {
			return nil, d.err
		}
		return CData(data), nil

	// Check only the first 2 bytes for doctype,
	// we will check the full sequence in readDoctype.
	// Check for doctype only after comment and cdata,
	// as they also start with `<!`.
	case bytes.HasPrefix(b, doctypeStart[:2]):
		data, doctypeOk := d.readDoctype()
		if !doctypeOk {
			return nil, d.err
		}
		return Directive(data), nil

	// If none of other cases matched, this should be a start element.
	default:
		startEl, ok := d.readStartTag()
		if !ok {
			return nil, d.err
		}

		return startEl, nil
	}
}

// readText reads text until the `<` character or EOF.
func (d *Decoder) readText() []byte {
	d.buf.Reset()

	for {
		b, ok := d.peekBuffered()
		if !ok {
			break
		}

		ind := bytes.IndexByte(b, '<')
		if ind >= 0 {
			// Found the start of a tag.
			b = b[:ind]
		}

		d.buf.Write(b)

		// Discard the bytes we've read.
		if !d.discard(len(b)) {
			break
		}

		if ind >= 0 {
			// We've read up to the start of a tag, break the loop.
			break
		}
	}

	// Return what we've read.
	return d.buf.Bytes()
}

// readStartTag reads a start tag.
func (d *Decoder) readStartTag() (StartElement, bool) {
	// Discard '<'
	if !d.discard(1) {
		return StartElement{}, false
	}

	name, ok := d.readNSName()
	if !ok {
		if d.err == nil {
			d.setSyntaxError("expected name after <")
		}
		return StartElement{}, false
	}

	tag := StartElement{Name: name}

	// Read attributes
	for {
		if !d.skipSpaces() {
			return tag, false
		}

		b, ok := d.mustReadByte()
		if !ok {
			return tag, false
		}

		if b == '/' {
			// Self-closing tag
			b, ok = d.mustReadByte()
			if !ok {
				return tag, false
			}
			if b != '>' {
				d.setSyntaxError("expected '>' at the end of self-closing tag, got %q", b)
				return tag, false
			}

			tag.SelfClosing = true

			break
		}

		if b == '>' {
			// End of start tag
			break
		}

		// Unread the byte for further processing
		d.unreadByte(b)

		// Read attribute name
		attrName, ok := d.readNSName()
		if !ok {
			if d.err == nil {
				d.setSyntaxError("expected attribute name")
			}
			return tag, false
		}

		if !d.skipSpaces() {
			return tag, false
		}

		b, ok = d.mustReadByte()
		if !ok {
			return tag, false
		}

		var attrValue string

		// If the next byte is not '=', this is an attribute without value.
		if b != '=' {
			d.unreadByte(b)
		} else {
			if !d.skipSpaces() {
				return tag, false
			}

			// Read attribute value
			val, ok := d.readAttrValue()
			if !ok {
				if d.err == nil {
					d.setSyntaxError("expected value for attribute %q", attrName.Local)
				}
				return tag, false
			}
			attrValue = string(val)
		}

		tag.Attr = append(tag.Attr, Attr{
			Name:  attrName,
			Value: attrValue,
		})
	}

	return tag, true
}

// readAttrValue reads an attribute value.
func (d *Decoder) readAttrValue() ([]byte, bool) {
	d.buf.Reset()

	b, ok := d.mustReadByte()
	if !ok {
		return nil, false
	}

	if b == '"' || b == '\'' {
		// Quoted attribute value
		// We can just read until the closing quote.
		if !d.mustReadUntil(b) {
			return nil, false
		}
		// Remove the trailing quote from the buffer
		d.buf.Remove(1)
	} else {
		// Unquoted attribute value.
		// Unread the byte for further processing
		d.unreadByte(b)
		// Read until we meet a byte that is not valid in an unquoted attribute value.
		if !d.mustReadWhileFn(isValueByte) {
			return nil, false
		}
	}

	return d.buf.Bytes(), true
}

// isValueByte checks if a byte is valid in an unquoted attribute value.
// See: https://www.w3.org/TR/REC-html40/intro/sgmltut.html#h-3.2.2
func isValueByte(c byte) bool {
	return 'A' <= c && c <= 'Z' ||
		'a' <= c && c <= 'z' ||
		'0' <= c && c <= '9' ||
		c == '_' || c == ':' || c == '-'
}

// readEndTag reads an end tag.
func (d *Decoder) readEndTag() (Name, bool) {
	// Discard '</'
	if !d.discard(len(endTagStart)) {
		return Name{}, false
	}

	name, ok := d.readNSName()
	if !ok {
		if d.err == nil {
			d.setSyntaxError("expected name after </")
		}
		return name, false
	}

	// Skip spaces before '>'
	if !d.skipSpaces() {
		return name, false
	}

	// Expect '>'
	b, ok := d.mustReadByte()
	if !ok {
		return name, false
	}
	if b != '>' {
		d.setSyntaxError("expected '>' at the end of end element, got %q", b)
		return name, false
	}

	return name, true
}

// readProcInst reads a processing instruction (until `?>`).
//
// If the processing instruction specifies an encoding, it recreates
// the reader with the specified encoding.
func (d *Decoder) readProcInst() ([]byte, []byte, bool) {
	// Discard '<?'
	if !d.discard(len(procInstStart)) {
		return nil, nil, false
	}

	// Target name should follow immediately after `<?`.
	if !d.readName() {
		// If we couldn't read a name but there was no error, it means
		// there was no valid target name after <?.
		// Set an error in this case.
		if d.err == nil {
			d.setSyntaxError("expected target name after <?")
		}
		return nil, nil, false
	}

	target := d.buf.Bytes()

	// Read until '?>'
	// We don't reset the buffer here, as we don't want target name to be overwritten.
	for {
		if !d.mustReadUntil('>') {
			return nil, nil, false
		}

		if d.buf.HasSuffix(procInstEnd) {
			break
		}
	}

	// Trim the trailing '?>'
	d.buf.Remove(len(procInstEnd))

	// Separate the target and data
	data := d.buf.Bytes()[len(target):]

	if bytes.Equal(target, targetXML) {
		// Get the encoding from the processing instruction data
		data = d.handleProcInstEncoding(data)
	}

	return target, data, true
}

// handleProcInstEncoding replaces the encoding declaration in the processing instruction data
// with "UTF-8" and returns the updated data.
// It also recreates the reader with defined encoding.
func (d *Decoder) handleProcInstEncoding(data []byte) []byte {
	matches := encodingRE.FindSubmatch(data)
	if matches == nil {
		// No encoding declaration found, return original data without changes
		return data
	}

	// Get the encoding from the processing instruction data
	encoding := bytes.Trim(matches[2], `"'`)

	if bytes.EqualFold(encoding, []byte("utf-8")) || bytes.EqualFold(encoding, []byte("utf8")) {
		// No need for special handling if encoding is already UTF-8
		return data
	}

	// Recreate the reader with defined encoding.
	// If the encoding is UTF-16/32, we have already handled it in the BOM check.
	if len(encoding) < 3 || !bytes.EqualFold(encoding[:3], []byte("utf")) {
		if !d.setEncoding(string(encoding)) {
			return data
		}
	}

	// Build the updated data with "UTF-8" encoding.
	// We write it to the buffer that already contains the processing instruction data,
	// so we mark the position of the updated data start.
	start := d.buf.Len()
	d.buf.Write(matches[1])        // Up to encoding=
	d.buf.Write([]byte(`"UTF-8"`)) // New encoding
	d.buf.Write(matches[3])        // After encoding declaration
	updated := d.buf.Bytes()[start:]

	return updated
}

// readComment reads a comment (until `-->`).
func (d *Decoder) readComment() ([]byte, bool) {
	if !d.checkAndDiscardPrefix(commentStart) {
		if d.err == nil {
			d.setSyntaxError("invalid sequence <!- not part of <!--")
		}
		return nil, false
	}

	d.buf.Reset()

	for {
		if !d.mustReadUntil('>') {
			return nil, false
		}

		if d.buf.HasSuffix(commentEnd) {
			break
		}
	}

	// Trim the trailing '-->'
	d.buf.Remove(len(commentEnd))

	return d.buf.Bytes(), true
}

// readCData reads a CDATA section (until `]]>`).
func (d *Decoder) readCData() ([]byte, bool) {
	if !d.checkAndDiscardPrefix(cdataStart) {
		if d.err == nil {
			d.setSyntaxError("invalid sequence <![ not part of <![CDATA[")
		}
		return nil, false
	}

	d.buf.Reset()

	// Read until ']]>'
	for {
		if !d.mustReadUntil('>') {
			return nil, false
		}

		if d.buf.HasSuffix(cdataEnd) {
			break
		}
	}

	// Trim the trailing ']]>'
	d.buf.Remove(len(cdataEnd))

	return d.buf.Bytes(), true
}

// readDoctype reads a directive (until `>`).
func (d *Decoder) readDoctype() ([]byte, bool) {
	if !d.checkAndDiscardPrefix(doctypeStart) {
		if d.err == nil {
			d.setSyntaxError("invalid sequence <! not part of <!DOCTYPE, <!--, or <![CDATA[")
		}
		return nil, false
	}

	d.buf.Reset()

	var (
		inQuote    byte // Quote character of the current quote (' or "), 0 if not in quote
		inBrackets bool // Whether we are inside brackets ([...]
	)

	// Read until '>'
	for {
		b, ok := d.mustReadByte()
		if !ok {
			return nil, false
		}

		d.buf.WriteByte(b)

		switch {
		case b == inQuote:
			// We met the closing quote, exit quote mode.
			inQuote = 0

		case inQuote != 0:
			// Inside a quote, do nothing.

		case b == '"' || b == '\'':
			// We met an opening quote, enter quote mode.
			inQuote = b

		case b == ']':
			// We met a closing bracket.
			// If we are not inside brackets, this is an error.
			if !inBrackets {
				d.setSyntaxError("unexpected ']' in directive")
				return nil, false
			}
			// Otherwise, exit brackets mode.
			inBrackets = false

		case b == '[':
			// We met an opening bracket.
			// If we are already inside brackets, this is an error.
			if inBrackets {
				d.setSyntaxError("nested '[' in directive")
				return nil, false
			}
			// Otherwise, enter brackets mode.
			inBrackets = true

		case inBrackets:
			// Inside brackets, do nothing

		case b == '<':
			// Unexpected '<' outside quotes and brackets.
			d.setSyntaxError("unexpected '<' in directive")
			return nil, false

		case b == '>':
			// End of directive.
			// Trim the trailing '>' from the buffer and return.
			d.buf.Remove(1)
			return d.buf.Bytes(), true
		}
	}
}

// readNSName reads a name with optional namespace prefix (e.g., "svg:svg").
func (d *Decoder) readNSName() (Name, bool) {
	var name Name

	if !d.readName() {
		return name, false
	}

	nameBytes := d.buf.Bytes()

	if space, local, ok := bytes.Cut(nameBytes, nameSep); ok && len(space) > 0 && len(local) > 0 {
		name.Space = string(space)
		name.Local = string(local)
	} else {
		name.Local = string(nameBytes)
	}

	return name, true
}

// readName reads a name (tag or attribute) to the buffer until a non-name byte is encountered.
func (d *Decoder) readName() bool {
	d.buf.Reset()

	if !d.mustReadWhileFn(isNameByte) {
		return false
	}

	return d.buf.Len() > 0
}

func isNameByte(c byte) bool {
	// We allow all non-ASCII bytes as names.
	return c >= utf8.RuneSelf ||
		'A' <= c && c <= 'Z' ||
		'a' <= c && c <= 'z' ||
		'0' <= c && c <= '9' ||
		c == '_' || c == ':' || c == '.' || c == '-'
}

// skipSpaces skips whitespace characters.
func (d *Decoder) skipSpaces() bool {
	for {
		b, ok := d.peekBuffered()
		if !ok {
			return false
		}

		found := false
		for i, c := range b {
			if !isSpace(c) {
				// Found a non-space byte.
				// Trim the bytes up to (but not including) this byte.
				b = b[:i]
				found = true
				break
			}
		}

		// Discard the spaces we've read.
		if !d.discard(len(b)) {
			return false
		}

		if found {
			// We've skipped all spaces, break the loop.
			return true
		}
	}
}

// isSpace checks if a byte is a whitespace character.
func isSpace(b byte) bool {
	return b == ' ' || b == '\r' || b == '\n' || b == '\t'
}

func (d *Decoder) checkAndDiscardPrefix(prefix []byte) bool {
	prefixLen := len(prefix)
	b, ok := d.mustPeek(prefixLen)
	if !ok {
		return false
	}
	if !bytes.Equal(b, prefix) {
		return false
	}
	return d.discard(prefixLen)
}

// readByte reads a single byte from the reader.
// If an error occurs, it sets d.err and returns false.
func (d *Decoder) readByte() (byte, bool) {
	b, err := d.r.ReadByte()
	if err != nil {
		d.err = err
		return 0, false
	}
	if b == '\n' {
		d.line++
	}
	return b, true
}

// mustReadByte reads a single byte from the reader.
// If an error occurs, it sets d.err and returns false.
// If io.EOF is encountered, it sets d.err to a more descriptive error.
func (d *Decoder) mustReadByte() (byte, bool) {
	b, ok := d.readByte()
	if !ok {
		if d.err == io.EOF {
			d.setSyntaxError("unexpected EOF")
		}
	}
	return b, ok
}

// unreadByte unreads the last byte read from the reader.
// If an error occurs, it sets d.err and returns false.
func (d *Decoder) unreadByte(b byte) bool {
	if err := d.r.UnreadByte(); err != nil {
		d.err = err
		return false
	}
	if b == '\n' {
		d.line--
	}
	return true
}

// mustReadUntil reads bytes to the buffer until the specified delimiter byte is encountered.
// The delimiter byte is included in the buffer.
func (d *Decoder) mustReadUntil(delim byte) bool {
	for {
		b, ok := d.mustPeekBuffered()
		if !ok {
			return false
		}

		ind := bytes.IndexByte(b, delim)
		if ind >= 0 {
			// Found the delimiter byte.
			// Trim the bytes up to and including the delimiter.
			b = b[:ind+1]
		}

		d.buf.Write(b)

		// Discard the bytes we've read.
		if !d.discard(len(b)) {
			return false
		}

		if ind >= 0 {
			// We've read up to the delimiter byte, break the loop.
			return true
		}
	}
}

// mustReadWhileFn reads bytes to the buffer while the provided function returns true.
// The byte that causes the function to return false is not included in the buffer.
func (d *Decoder) mustReadWhileFn(f func(byte) bool) bool {
	for {
		b, ok := d.mustPeekBuffered()
		if !ok {
			return false
		}

		found := false
		for i, c := range b {
			if !f(c) {
				// Found a byte that does not satisfy the condition.
				// Trim the bytes up to (but not including) this byte.
				b = b[:i]
				found = true
				break
			}
		}

		d.buf.Write(b)

		// Discard the bytes we've read.
		if !d.discard(len(b)) {
			return false
		}

		if found {
			// We've read up to the delimiter byte, break the loop.
			return true
		}
	}
}

// peek peeks at the next n bytes without advancing the reader.
func (d *Decoder) peek(n int) ([]byte, bool) {
	b, err := d.r.Peek(n)
	if err != nil {
		d.err = err
		return nil, false
	}
	return b, true
}

// mustPeek peeks at the next n bytes without advancing the reader.
// If an error occurs, it sets d.err and returns false.
// If io.EOF is encountered, it sets d.err to a more descriptive error.
func (d *Decoder) mustPeek(n int) ([]byte, bool) {
	b, ok := d.peek(n)
	if !ok {
		if d.err == io.EOF {
			d.setSyntaxError("unexpected EOF")
		}
	}
	return b, ok
}

// peekBuffered peeks at all currently buffered bytes without advancing the reader.
// If no bytes are buffered, it peeks at least 1 byte.
// If an error occurs, it sets d.err and returns false.
func (d *Decoder) peekBuffered() ([]byte, bool) {
	toPeek := max(d.r.Buffered(), 1)
	return d.peek(toPeek)
}

// mustPeekBuffered peeks at all currently buffered bytes without advancing the reader.
// If no bytes are buffered, it peeks at least 1 byte.
// If an error occurs, it sets d.err and returns false.
// If io.EOF is encountered, it sets d.err to a more descriptive error.
func (d *Decoder) mustPeekBuffered() ([]byte, bool) {
	toPeek := max(d.r.Buffered(), 1)
	return d.mustPeek(toPeek)
}

// discard discards the next n bytes from the reader.
func (d *Decoder) discard(n int) bool {
	// Peek bytes we want to discard to count new lines.
	if b, err := d.r.Peek(n); err == nil {
		// Somehow this is more efficient than bytes.Count...
		for {
			ind := bytes.IndexByte(b, '\n')
			if ind < 0 {
				break
			}
			d.line++
			b = b[ind+1:]
		}
	}

	_, err := d.r.Discard(n)
	if err != nil {
		d.err = err
		return false
	}
	return true
}

func (d *Decoder) setSyntaxError(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	d.err = newSyntaxError("%s (line %d)", msg, d.line)
}
