package logger

import (
	"log/slog"
	"time"
)

var (
	prettyGroupOpenToken  = []byte("{")
	prettyGroupCloseToken = []byte(" }")
)

// formatterPretty is a pretty printer for log records.
type formatterPretty struct {
	formatterCommon

	color        int
	groupsOpened int
}

// newFormatterPretty creates a new instance of formatterPretty.
func newFormatterPretty(groups []attrGroup, buf *buffer) *formatterPretty {
	return &formatterPretty{
		formatterCommon: newFormatterCommon(groups, buf),
	}
}

// format formats a log record as a pretty-printed string.
func (s *formatterPretty) format(r slog.Record) {
	s.color = s.getColor(r.Level)

	// Append timestamp
	s.buf.appendStringRaw(r.Time.Format(time.DateTime))
	s.buf.append(' ')

	// Append level marker
	s.buf.appendf("\x1b[1;%dm%s\x1b[0m ", s.color, s.levelName(r.Level))

	// Append message
	s.buf.appendStringRaw(r.Message)

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
		s.appendValue(s.error.Value, false)
	}
	if s.source.Key != "" {
		s.appendKey(s.source.Key)
		s.appendValue(s.source.Value, false)
	}
	if s.stack.Key != "" {
		s.buf.appendf("\n\x1b[37m%v\x1b[0m", s.stack.Value.String())
	}
}

// getColor returns the color code for a given log level.
func (s *formatterPretty) getColor(lvl slog.Level) int {
	switch {
	case lvl < slog.LevelInfo:
		return 37
	case lvl < slog.LevelWarn:
		return 36
	case lvl < slog.LevelError:
		return 33
	default:
		return 31
	}
}

// levelName returns the string representation of a log level.
func (s *formatterPretty) levelName(lvl slog.Level) string {
	switch {
	case lvl < slog.LevelInfo:
		return "[DBG]"
	case lvl < slog.LevelWarn:
		return "[INF]"
	case lvl < slog.LevelError:
		return "[WRN]"
	case lvl < LevelCritical:
		return "[ERR]"
	default:
		return "[CRT]"
	}
}

// appendAttributes appends a list of attributes to the buffer.
func (s *formatterPretty) appendAttributes(attrs []slog.Attr) {
	for _, attr := range attrs {
		s.appendAttribute(attr)
	}
}

// appendAttribute appends a single attribute to the buffer.
func (s *formatterPretty) appendAttribute(attr slog.Attr) {
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
	s.appendValue(attr.Value, false)
}

// appendKey appends an attribute key to the buffer.
func (s *formatterPretty) appendKey(key string) {
	s.buf.appendf(" \x1b[%dm", s.color)
	s.buf.appendString(key)
	s.buf.appendStringRaw("\x1b[0m=")
}

// appendGroup appends a group of attributes to the buffer.
func (s *formatterPretty) appendGroup(name string, attrs []slog.Attr) {
	// Ignore empty groups
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

// openGroup opens a new group of attributes.
func (s *formatterPretty) openGroup(name string) {
	s.groupsOpened++

	s.appendKey(name)
	s.buf.append(prettyGroupOpenToken...)
}

// closeGroup closes the most recently opened group of attributes.
func (s *formatterPretty) closeGroup() {
	s.groupsOpened--

	s.buf.append(prettyGroupCloseToken...)
}
