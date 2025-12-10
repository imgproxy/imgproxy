package logger

import (
	"strconv"
	"sync"
	"unicode/utf8"
)

var bufPool = sync.Pool{
	New: func() any {
		// Reserve some capacity to not re-allocate on short logs.
		buf := make(buffer, 0, 1024)
		return &buf
	},
}

// buffer is a slice of bytes with some additional convenience methods.
type buffer []byte

// newBuffer creates a new buffer from the pool.
func newBuffer() *buffer {
	//nolint:forcetypeassert
	return bufPool.Get().(*buffer)
}

// Write writes data to the buffer.
func (b *buffer) Write(p []byte) (n int, err error) {
	b.append(p...)
	return len(p), nil
}

// String returns the contents of the buffer as a string.
func (b *buffer) String() string {
	return string(*b)
}

// free truncates the buffer and returns it to the pool.
func (b *buffer) free() {
	// Don't keep large buffers around.
	if len(*b) > 16*1024 {
		return
	}

	*b = (*b)[:0]
	bufPool.Put(b)
}

// len returns the number of bytes written to the buffer.
func (b *buffer) len() int {
	return len(*b)
}

// append appends data to the buffer.
func (b *buffer) append(data ...byte) {
	*b = append(*b, data...)
}

// appendString appends a string value to the buffer.
// If the string does not require escaping, it is appended directly.
// Otherwise, it is escaped and quoted.
func (b *buffer) appendString(data string) {
	if b.isStringQuoteSafe(data) {
		b.appendStringRaw(data)
	} else {
		b.appendStringQuoted(data)
	}
}

// appendStringRaw appends a string value to the buffer without escaping.
func (b *buffer) appendStringRaw(data string) {
	*b = append(*b, data...)
}

// appendStringQuoted appends a string value to the buffer, escaping and quoting it as necessary.
func (b *buffer) appendStringQuoted(data string) {
	*b = strconv.AppendQuote(*b, data)
}

// appendInt appends an integer value to the buffer.
func (b *buffer) appendInt(data int64) {
	*b = strconv.AppendInt(*b, data, 10)
}

// appendUint appends an unsigned integer value to the buffer.
func (b *buffer) appendUint(data uint64) {
	*b = strconv.AppendUint(*b, data, 10)
}

// appendFloat appends a float value to the buffer.
func (b *buffer) appendFloat(data float64) {
	*b = strconv.AppendFloat(*b, data, 'g', -1, 64)
}

// appendBool appends a boolean value to the buffer.
func (b *buffer) appendBool(data bool) {
	*b = strconv.AppendBool(*b, data)
}

// remove removes the last n bytes from the buffer.
func (b *buffer) remove(n int) {
	n = max(0, n)
	trimTo := max(0, len(*b)-n)
	*b = (*b)[:trimTo]
}

// removeNewline removes the trailing newline character from the buffer, if present.
func (b *buffer) removeNewline() {
	if len(*b) > 0 && (*b)[len(*b)-1] == '\n' {
		*b = (*b)[:len(*b)-1]
	}
}

// isStringQuoteSafe checks if a string is safe to append without quoting.
func (b *buffer) isStringQuoteSafe(val string) bool {
	for i := range len(val) {
		if b := val[i]; b >= utf8.RuneSelf || !quoteSafeSet[b] {
			return false
		}
	}
	return true
}

// quoteSafeSet is a set of runes that are safe to append without quoting.
// Some runes here are explicitly marked as unsafe for clarity.
// The unlisted runes are considered unsafe by default.
// Shamesly stolen from https://github.com/golang/go/blob/master/src/encoding/json/tables.go
// and tuned for our needs.
var quoteSafeSet = [utf8.RuneSelf]bool{
	' ':  false,
	'!':  true,
	'"':  false,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': false,
	'(':  true,
	')':  true,
	'*':  true,
	'+':  true,
	',':  true,
	'-':  true,
	'.':  true,
	'/':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	':':  false,
	';':  true,
	'<':  true,
	'=':  false,
	'>':  true,
	'?':  true,
	'@':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'V':  true,
	'W':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'[':  true,
	'\\': false,
	']':  true,
	'^':  true,
	'_':  true,
	'`':  false,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'{':  false,
	'|':  true,
	'}':  false,
	'~':  true,
}
