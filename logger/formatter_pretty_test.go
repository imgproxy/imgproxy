package logger

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type FormatterPrettyTestSuite struct {
	testutil.LazySuite

	buf     testutil.LazyObj[*bytes.Buffer]
	config  testutil.LazyObj[*Config]
	handler testutil.LazyObj[*Handler]
	logger  testutil.LazyObj[*slog.Logger]
}

func (s *FormatterPrettyTestSuite) SetupTest() {
	s.buf, _ = testutil.NewLazySuiteObj(
		s,
		func() (*bytes.Buffer, error) {
			return new(bytes.Buffer), nil
		},
	)

	s.config, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Config, error) {
			cfg := NewDefaultConfig()
			cfg.Format = FormatPretty
			return &cfg, nil
		},
	)

	s.handler, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Handler, error) {
			return NewHandler(s.buf(), s.config()), nil
		},
	)

	s.logger, _ = testutil.NewLazySuiteObj(
		s,
		func() (*slog.Logger, error) {
			return slog.New(s.handler()), nil
		},
	)
}

func (s *FormatterPrettyTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *FormatterPrettyTestSuite) TestLevel() {
	type testEntry struct {
		level     slog.Level
		levelName string
		message   string
	}

	testEntries := []testEntry{
		{level: slog.LevelDebug, levelName: "DBG", message: "Debug message"},
		{level: slog.LevelInfo, levelName: "INF", message: "Info message"},
		{level: slog.LevelWarn, levelName: "WRN", message: "Warning message"},
		{level: slog.LevelError, levelName: "ERR", message: "Error message"},
		{level: LevelCritical, levelName: "CRT", message: "Critical message"},
	}

	testCases := []struct {
		level   slog.Level
		entries []testEntry
	}{
		{level: slog.LevelDebug, entries: testEntries},
		{level: slog.LevelInfo, entries: testEntries[1:]},
		{level: slog.LevelWarn, entries: testEntries[2:]},
		{level: slog.LevelError, entries: testEntries[3:]},
		{level: LevelCritical, entries: testEntries[4:]},
	}

	for _, tc := range testCases {
		s.Run(tc.level.String(), func() {
			s.config().Level = tc.level

			for _, entry := range testEntries {
				s.logger().Log(s.T().Context(), entry.level, entry.message)
			}

			for _, entry := range tc.entries {
				s.checkNextEntry(entry.levelName, entry.message)
			}
		})
	}
}

func (s *FormatterPrettyTestSuite) TestAttributes() {
	s.logger().Info(
		"Test message",
		slog.String("string", "value"),
		slog.Int("int", -100),
		slog.Uint64("uint64", 200),
		slog.Float64("float64", 3.14),
		slog.Bool("bool", true),
		slog.Time("time", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
		slog.Duration("duration", time.Minute),
		slog.Any("err", errors.New("error value")),
		slog.Any("any", struct {
			Field1 string
			Field2 int
		}{"value", 42}),
	)

	s.checkNextEntry(
		"INF",
		strings.Join([]string{
			"Test message",
			"string=value",
			"int=-100",
			"uint64=200",
			"float64=3.14",
			"bool=true",
			`time="1984-01-02 03:04:05"`,
			"duration=1m0s",
			`err="error value"`,
			`any="{Field1:value Field2:42}"`,
		}, " "),
	)
}

func (s *FormatterPrettyTestSuite) TestGroups() {
	s.Run("LastGroupNotEmpty", func() {
		s.logger().
			With(
				slog.String("string", "value"),
				slog.Int("int", -100),
			).
			WithGroup("group1").
			With(
				slog.Uint64("uint64", 200),
				slog.Float64("float64", 3.14),
			).
			WithGroup("group2").
			With(slog.Group("group3",
				slog.Bool("bool", true),
				slog.Time("time", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
			)).
			With(slog.Group("empty_group")).
			WithGroup("group4").
			Info(
				"Test message",
				slog.Duration("duration", time.Minute),
				slog.Any("any", struct {
					Field1 string
					Field2 int
				}{"value", 42}),
			)

		s.checkNextEntry(
			"INF",
			strings.Join([]string{
				"Test message",
				"string=value",
				"int=-100",
				"group1={",
				"uint64=200",
				"float64=3.14",
				"group2={",
				"group3={",
				"bool=true",
				`time="1984-01-02 03:04:05"`,
				"}",
				"group4={",
				"duration=1m0s",
				`any="{Field1:value Field2:42}"`,
				"}",
				"}",
				"}",
			}, " "),
		)
	})

	s.Run("LastGroupsEmpty", func() {
		s.logger().
			With(
				slog.String("string", "value"),
				slog.Int("int", -100),
			).
			WithGroup("group1").
			With(
				slog.Uint64("uint64", 200),
				slog.Float64("float64", 3.14),
			).
			WithGroup("group2").
			With(slog.Group("group3",
				slog.Bool("bool", true),
				slog.Time("time", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
			)).
			With(slog.Group("empty_group")).
			WithGroup("group4").
			WithGroup("group5").
			Info("Test message")

		s.checkNextEntry(
			"INF",
			strings.Join([]string{
				"Test message",
				"string=value",
				"int=-100",
				"group1={",
				"uint64=200",
				"float64=3.14",
				"group2={",
				"group3={",
				"bool=true",
				`time="1984-01-02 03:04:05"`,
				"}",
				"}",
				"}",
			}, " "),
		)
	})
}

func (s *FormatterPrettyTestSuite) TestQuoting() {
	s.logger().Info(
		"Test message",
		"key", "value",
		"key with spaces", "value with spaces",
		`key"with"quotes`, `value"with"quotes`,
		"key\nwith\nnewlines", "value\nwith\nnewlines",
		slog.Group("group name", slog.String("key", "value")),
	)

	s.checkNextEntry(
		"INF",
		strings.Join([]string{
			"Test message",
			"key=value",
			`"key with spaces"="value with spaces"`,
			`"key\"with\"quotes"="value\"with\"quotes"`,
			`"key\nwith\nnewlines"="value\nwith\nnewlines"`,
			`"group name"={ key=value }`,
		}, " "),
	)
}

func (s *FormatterPrettyTestSuite) TestSpecialFields() {
	s.logger().Info(
		"Test message",
		"stack", "stack value\nwith new lines",
		"key1", "value1",
		"error", errctx.NewTextError(
			"test", 0, errctx.WithDocsURL("http://example.com"),
		),
		"key2", "value2",
		"source", "source value",
		"key3", "value3",
		slog.Group(
			"group",
			"stack", "stack in group",
			"error", "error in group",
			"source", "source in group",
		),
	)

	s.checkNextEntry(
		"INF",
		strings.Join([]string{
			"Test message",
			"key1=value1",
			"key2=value2",
			"key3=value3",
			"group={",
			`stack="stack in group"`,
			`error="error in group"`,
			`source="source in group"`,
			"}",
			`error=test`,
			`error_docs_url="http://example.com"`,
			`source="source value"`,
		}, " "),
	)

	s.removeColorCodes()

	s.Require().Equal("stack value\nwith new lines\n", s.buf().String())
}

func (s *FormatterPrettyTestSuite) removeColorCodes() {
	p := s.buf().Bytes()
	q := p[:0]

	inEscape := false

	for _, b := range p {
		switch {
		case b == '\x1b':
			// Skip ANSI escape codes
			inEscape = true
		case inEscape && b == 'm':
			inEscape = false
		case !inEscape:
			q = append(q, b)
		}
	}

	s.buf().Truncate(len(q))
}

func (s *FormatterPrettyTestSuite) checkNextEntry(lvl, msg string) {
	// Remove color codes from the log entry,
	// we're not going to test coloring
	s.removeColorCodes()

	// Pretty level names are always 3 characters long
	s.Require().Len(lvl, 3)

	str, err := s.buf().ReadString('\n')
	s.Require().NoError(err)

	const timeLen = 19
	const lvlLen = 3 + 4 // +4 for the space and brackets
	const prefixLen = timeLen + lvlLen

	s.Require().GreaterOrEqual(len(str), prefixLen)

	timePart := str[:timeLen]
	levelPart := str[timeLen:prefixLen]

	now := time.Now()
	t, err := time.ParseInLocation(time.DateTime, timePart, now.Location())
	s.Require().NoError(err)
	s.Require().WithinDuration(time.Now(), t, time.Minute)

	s.Equal(" ["+lvl+"] ", levelPart)

	// Check the message
	s.Equal(msg+"\n", str[prefixLen:])
}

func TestFormatterPretty(t *testing.T) {
	suite.Run(t, new(FormatterPrettyTestSuite))
}
