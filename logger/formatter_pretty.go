package logger

import (
	"log/slog"
	"time"
)

var (
	prettyGroupOpenToken  = []byte("{")
	prettyGroupCloseToken = []byte(" }")

	prettyColorDebug     = []byte("\x1b[34m")   // Blue
	prettyColorDebugBold = []byte("\x1b[1;34m") // Bold Blue
	prettyColorInfo      = []byte("\x1b[32m")   // Green
	prettyColorInfoBold  = []byte("\x1b[1;32m") // Bold Green
	prettyColorWarn      = []byte("\x1b[33m")   // Yellow
	prettyColorWarnBold  = []byte("\x1b[1;33m") // Bold Yellow
	prettyColorError     = []byte("\x1b[31m")   // Red
	prettyColorErrorBold = []byte("\x1b[1;31m") // Bold Red
	prettyColorStack     = []byte("\x1b[2m")    // Dimmed default
	prettyColorReset     = []byte("\x1b[0m")
)

// formatterPretty is a pretty printer for log records.
type formatterPretty struct {
	formatterCommon

	colorThin []byte
	colorBold []byte

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
	s.colorThin, s.colorBold = s.getColor(r.Level)

	// Append timestamp
	s.buf.appendStringRaw(r.Time.Format(time.DateTime))
	s.buf.append(' ')

	// Append level marker
	s.buf.append(s.colorBold...)
	s.buf.appendStringRaw(s.levelName(r.Level))
	s.buf.append(prettyColorReset...)
	s.buf.append(' ')

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

		if docsURL := s.errorDocsURL(); docsURL != nil {
			s.appendKey(docsURL.Key)
			s.appendValue(docsURL.Value, false)
		}
	}
	if s.source.Key != "" {
		s.appendKey(s.source.Key)
		s.appendValue(s.source.Value, false)
	}
	if s.stack.Key != "" {
		s.buf.append('\n')
		s.buf.append(prettyColorStack...)
		s.buf.appendStringRaw(s.stack.Value.String())
		s.buf.append(prettyColorReset...)
	}
}

// getColor returns the terminal color sequences for a given log level.
func (s *formatterPretty) getColor(lvl slog.Level) ([]byte, []byte) {
	switch {
	case lvl < slog.LevelInfo:
		return prettyColorDebug, prettyColorDebugBold
	case lvl < slog.LevelWarn:
		return prettyColorInfo, prettyColorInfoBold
	case lvl < slog.LevelError:
		return prettyColorWarn, prettyColorWarnBold
	default:
		return prettyColorError, prettyColorErrorBold
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
	s.buf.append(' ')
	s.buf.append(s.colorThin...)
	s.buf.appendString(key)
	s.buf.append(prettyColorReset...)
	s.buf.append('=')
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
