package logger

import (
	"encoding"
	"fmt"
	"log/slog"
	"time"
)

// formatterCommon holds the common logic for both pretty and structured formatting.
type formatterCommon struct {
	buf *buffer

	groups []attrGroup

	// Attributes that should be handled specially
	error  slog.Attr
	source slog.Attr
	stack  slog.Attr
}

// newHandlerFormatterCommon creates a new formatterCommon instance.
func newFormatterCommon(groups []attrGroup, buf *buffer) formatterCommon {
	return formatterCommon{
		buf:    buf,
		groups: groups,
	}
}

// levelName returns the name of the log level.
func (s *formatterCommon) levelName(lvl slog.Level) string {
	switch {
	case lvl < slog.LevelInfo:
		return "DEBUG"
	case lvl < slog.LevelWarn:
		return "INFO"
	case lvl < slog.LevelError:
		return "WARNING"
	case lvl < LevelCritical:
		return "ERROR"
	default:
		return "CRITICAL"
	}
}

// saveSpecialAttr saves special attributes for later use.
// It returns true if the attribute was saved (meaning it was a special attribute).
func (s *formatterCommon) saveSpecialAttr(attr slog.Attr) bool {
	switch attr.Key {
	case "error":
		s.error = attr
	case "source":
		s.source = attr
	case "stack":
		s.stack = attr
	default:
		return false
	}

	return true
}

// appendValue appends a value to the buffer, applying quoting rules as necessary.
func (s *formatterCommon) appendValue(val slog.Value, forceQuote bool) {
	switch val.Kind() {
	case slog.KindString:
		s.appendString(val.String(), forceQuote)
	case slog.KindInt64:
		s.buf.appendInt(val.Int64())
	case slog.KindUint64:
		s.buf.appendUint(val.Uint64())
	case slog.KindFloat64:
		s.buf.appendFloat(val.Float64())
	case slog.KindBool:
		s.buf.appendBool(val.Bool())
	case slog.KindDuration:
		s.appendString(val.Duration().String(), forceQuote)
	case slog.KindTime:
		s.appendTime(val.Time())
	default:
		s.appendAny(val.Any(), forceQuote)
	}
}

// appendString appends a string value to the buffer, applying quoting rules as necessary.
func (s *formatterCommon) appendString(val string, forceQuote bool) {
	if forceQuote {
		s.buf.appendStringQuoted(val)
	} else {
		s.buf.appendString(val)
	}
}

// appendTime appends a time value to the buffer, wrapping it in quotes,
// ([time.DateTime] always contains a space)
func (s *formatterCommon) appendTime(val time.Time) {
	s.buf.append('"')
	s.buf.appendStringRaw(val.Format(time.DateTime))
	s.buf.append('"')
}

// appendAny appends a value of any type to the buffer, applying quoting rules as necessary.
func (s *formatterCommon) appendAny(val any, forceQuote bool) {
	switch v := val.(type) {
	case fmt.Stringer:
		s.appendString(v.String(), forceQuote)
		return
	case error:
		s.appendString(v.Error(), forceQuote)
		return
	case encoding.TextMarshaler:
		if data, err := v.MarshalText(); err == nil {
			s.appendString(string(data), forceQuote)
			return
		}
	}
	// Fallback to default string representation
	s.appendString(fmt.Sprintf("%+v", val), forceQuote)
}
