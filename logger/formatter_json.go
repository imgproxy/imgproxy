package logger

import (
	"encoding/json"
	"log/slog"
	"time"
	"unicode/utf8"
)

const (
	jsonGroupOpenToken  = '{'
	jsonGroupCloseToken = '}'
)

var jsonAttributeSep = []byte(",")

// formatterJSON is a JSON log formatter.
type formatterJSON struct {
	formatterCommon

	levelKey   string
	messageKey string

	sep          []byte
	groupsOpened int
}

// newFormatterJSON creates a new formatterJSON instance.
func newFormatterJSON(groups []attrGroup, buf *buffer, gcpStyle bool) *formatterJSON {
	f := &formatterJSON{
		formatterCommon: newFormatterCommon(groups, buf),
	}

	// Set the level and message keys based on the style.
	if gcpStyle {
		f.levelKey = "severity"
		f.messageKey = "message"
	} else {
		f.levelKey = slog.LevelKey
		f.messageKey = slog.MessageKey
	}

	return f
}

// format formats a log record.
func (s *formatterJSON) format(r slog.Record) {
	// Open the JSON object and defer closing it.
	s.buf.append(jsonGroupOpenToken)
	defer func() {
		s.buf.append(jsonGroupCloseToken)
	}()

	// Append timestamp
	s.appendKey(slog.TimeKey)
	s.appendTime(r.Time)

	// Append log level
	s.appendKey(s.levelKey)
	s.appendString(s.levelName(r.Level))

	// Append message
	s.appendKey(s.messageKey)
	s.appendString(r.Message)

	// Append groups added with [Handler.WithAttrs] and [Handler.WithGroup]
	for _, g := range s.groups {
		if g.name != "" {
			s.openGroup(g.name)
		}

		s.appendAttributes(g.attrs)
	}

	// Append attributes from the record
	r.Attrs(func(attr slog.Attr) bool {
		s.appendAttribute(attr)
		return true
	})

	// Close all opened groups.
	for s.groupsOpened > 0 {
		s.closeGroup()
	}

	// Append error, source, and stack if present
	if s.error.Key != "" {
		s.appendKey(s.error.Key)
		s.appendValue(s.error.Value)
	}
	if s.source.Key != "" {
		s.appendKey(s.source.Key)
		s.appendValue(s.source.Value)
	}
	if s.stack.Key != "" {
		s.appendKey(s.stack.Key)
		s.appendValue(s.stack.Value)
	}
}

// appendAttributes appends a list of attributes to the buffer.
func (s *formatterJSON) appendAttributes(attrs []slog.Attr) {
	for _, attr := range attrs {
		s.appendAttribute(attr)
	}
}

// appendAttribute appends a single attribute to the buffer.
func (s *formatterJSON) appendAttribute(attr slog.Attr) {
	// Resolve [slog.LogValuer] values
	attr.Value = attr.Value.Resolve()

	// If there are no groups opened, save special attributes for later
	if s.groupsOpened == 0 && s.saveSpecialAttr(attr) {
		return
	}

	// Groups need special handling
	if attr.Value.Kind() == slog.KindGroup {
		s.appendGroup(attr.Key, attr.Value.Group())
		return
	}

	s.appendKey(attr.Key)
	s.appendValue(attr.Value)
}

// appendKey appends an attribute key to the buffer.
func (s *formatterJSON) appendKey(key string) {
	s.buf.append(s.sep...)
	s.sep = jsonAttributeSep

	s.appendString(key)
	s.buf.append(':')
}

// appendValue appends a value to the buffer, applying quoting rules as necessary.
func (s *formatterJSON) appendValue(val slog.Value) {
	switch val.Kind() {
	case slog.KindString:
		s.appendString(val.String())
	case slog.KindInt64:
		s.buf.appendInt(val.Int64())
	case slog.KindUint64:
		s.buf.appendUint(val.Uint64())
	case slog.KindFloat64:
		// strconv.FormatFloat result sometimes differs from json.Marshal,
		// so we use json.Marshal for consistency.
		s.appendJSONMarshal(val.Float64())
	case slog.KindBool:
		s.buf.appendBool(val.Bool())
	case slog.KindDuration:
		s.buf.appendInt(int64(val.Duration()))
	case slog.KindTime:
		s.appendTime(val.Time())
	default:
		s.appendJSONMarshal(val.Any())
	}
}

// appendString appends a string value to the buffer.
// If the string does not require escaping, it is appended directly.
// Otherwise, it is JSON marshaled.
func (s *formatterJSON) appendString(val string) {
	if !s.isStringSafe(val) {
		s.appendJSONMarshal(val)
		return
	}

	s.buf.append('"')
	s.buf.appendStringRaw(val)
	s.buf.append('"')
}

// isStringSafe checks if a string is safe to append without escaping.
func (s *formatterJSON) isStringSafe(val string) bool {
	for i := 0; i < len(val); i++ {
		if b := val[i]; b >= utf8.RuneSelf || !jsonSafeSet[b] {
			return false
		}
	}
	return true
}

// appendTime appends a time value to the buffer.
func (s *formatterJSON) appendTime(val time.Time) {
	s.buf.append('"')
	s.buf.appendStringRaw(val.Format(time.RFC3339))
	s.buf.append('"')
}

// appendJSONMarshal appends a JSON marshaled value to the buffer.
func (s *formatterJSON) appendJSONMarshal(val any) {
	if err, ok := val.(error); ok && err != nil {
		s.appendString(err.Error())
		return
	}

	buf := newBuffer()
	defer func() {
		buf.free()
	}()

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(val); err != nil {
		// This should be a very unlikely situation, but just in case...
		s.buf.appendStringRaw(`"<json marshal error>"`)
		return
	}

	buf.removeNewline()
	s.buf.append(*buf...)
}

// appendGroup appends a group of attributes to the buffer.
func (s *formatterJSON) appendGroup(name string, attrs []slog.Attr) {
	if len(attrs) == 0 {
		return
	}

	if len(name) > 0 {
		// If the group has a name, open it and defer closing it.
		// Unnamed groups should be treated as sets of regular attributes.
		s.openGroup(name)
		defer s.closeGroup()
	}

	s.appendAttributes(attrs)
}

// openGroup opens a new group in the buffer.
func (s *formatterJSON) openGroup(name string) {
	s.groupsOpened++

	s.appendKey(name)
	s.buf.append(jsonGroupOpenToken)
	s.sep = nil
}

// closeGroup closes the most recently opened group in the buffer.
func (s *formatterJSON) closeGroup() {
	s.groupsOpened--

	s.buf.append(jsonGroupCloseToken)
	s.sep = jsonAttributeSep
}

// jsonSafeSet is a set of runes that are safe to include in JSON strings without escaping.
// Some runes here are explicitly marked as unsafe for clarity.
// The unlisted runes are considered unsafe by default.
// Shamesly stolen from https://github.com/golang/go/blob/master/src/encoding/json/tables.go.
var jsonSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
