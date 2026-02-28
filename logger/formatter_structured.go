package logger

import (
	"log/slog"
)

// formatterStructured is a flat structured log formatter.
type formatterStructured struct {
	formatterCommon

	// Current group prefix
	prefix *buffer
}

// newFormatterStructured creates a new formatterStructured instance.
func newFormatterStructured(groups []attrGroup, buf *buffer) *formatterStructured {
	return &formatterStructured{
		formatterCommon: newFormatterCommon(groups, buf),
	}
}

// format formats a log record.
func (s *formatterStructured) format(r slog.Record) {
	s.prefix = newBuffer()
	defer func() {
		s.prefix.free()
	}()

	// Append timestamp
	s.appendKey(slog.TimeKey)
	s.appendTime(r.Time)

	// Append log level
	s.appendKey(slog.LevelKey)
	s.appendString(s.levelName(r.Level), true)

	// Append message
	s.appendKey(slog.MessageKey)
	s.appendString(r.Message, true)

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

	// Append error, source, and stack if present
	if s.error.Key != "" {
		s.appendKey(s.error.Key)
		s.appendValue(s.error.Value)

		if docsURL := s.errorDocsURL(); docsURL != nil {
			s.appendKey(docsURL.Key)
			s.appendValue(docsURL.Value)
		}
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
func (s *formatterStructured) appendAttributes(attrs []slog.Attr) {
	for _, attr := range attrs {
		s.appendAttribute(attr)
	}
}

// appendAttribute appends a single attribute to the buffer.
func (s *formatterStructured) appendAttribute(attr slog.Attr) {
	// Resolve [slog.LogValuer] values
	attr.Value = attr.Value.Resolve()

	// If there are no groups opened, save special attributes for later
	if s.prefix.len() == 0 && s.saveSpecialAttr(attr) {
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
func (s *formatterStructured) appendKey(key string) {
	if len(*s.buf) > 0 {
		s.buf.append(' ')
	}

	s.buf.appendString(s.prefix.String() + key)
	s.buf.append('=')
}

// appendValue appends an attribute value to the buffer, applying quoting.
func (s *formatterStructured) appendValue(val slog.Value) {
	s.formatterCommon.appendValue(val, true)
}

// appendGroup appends a group of attributes to the buffer.
func (s *formatterStructured) appendGroup(name string, attrs []slog.Attr) {
	if len(attrs) == 0 {
		return
	}

	if len(name) > 0 {
		// If the group has a name, open it and defer closing it.
		// Unnamed groups should be treated as sets of regular attributes.
		s.openGroup(name)
		defer s.closeGroup(name)
	}

	s.appendAttributes(attrs)
}

// openGroup opens a new group of attributes.
func (s *formatterStructured) openGroup(name string) {
	s.prefix.appendStringRaw(name)
	s.prefix.append('.')
}

// closeGroup closes the most recently opened group of attributes.
func (s *formatterStructured) closeGroup(name string) {
	s.prefix.remove(len(name) + 1) // +1 for the dot
}
