package logger

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type FormatterStructuredTestSuite struct {
	testutil.LazySuite

	buf     testutil.LazyObj[*bytes.Buffer]
	config  testutil.LazyObj[*Config]
	handler testutil.LazyObj[*Handler]
	logger  testutil.LazyObj[*slog.Logger]
}

func (s *FormatterStructuredTestSuite) SetupTest() {
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
			cfg.Format = FormatStructured
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

func (s *FormatterStructuredTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *FormatterStructuredTestSuite) TestLevel() {
	type testEntry struct {
		level     slog.Level
		levelName string
		message   string
	}

	testEntries := []testEntry{
		{level: slog.LevelDebug, levelName: "DEBUG", message: "Debug message"},
		{level: slog.LevelInfo, levelName: "INFO", message: "Info message"},
		{level: slog.LevelWarn, levelName: "WARNING", message: "Warning message"},
		{level: slog.LevelError, levelName: "ERROR", message: "Error message"},
		{level: LevelCritical, levelName: "CRITICAL", message: "Critical message"},
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
				s.checkNextEntry(entry.levelName, fmt.Sprintf(`msg="%s"`, entry.message))
			}
		})
	}
}

func (s *FormatterStructuredTestSuite) TestAttributes() {
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
		"INFO",
		strings.Join([]string{
			`msg="Test message"`,
			`string="value"`,
			`int=-100`,
			`uint64=200`,
			`float64=3.14`,
			`bool=true`,
			`time="1984-01-02 03:04:05"`,
			`duration="1m0s"`,
			`err="error value"`,
			`any="{Field1:value Field2:42}"`,
		}, " "),
	)
}

func (s *FormatterStructuredTestSuite) TestGroups() {
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
			"INFO",
			strings.Join([]string{
				`msg="Test message"`,
				`string="value"`,
				`int=-100`,
				`group1.uint64=200`,
				`group1.float64=3.14`,
				`group1.group2.group3.bool=true`,
				`group1.group2.group3.time="1984-01-02 03:04:05"`,
				`group1.group2.group4.duration="1m0s"`,
				`group1.group2.group4.any="{Field1:value Field2:42}"`,
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
			"INFO",
			strings.Join([]string{
				`msg="Test message"`,
				`string="value"`,
				`int=-100`,
				`group1.uint64=200`,
				`group1.float64=3.14`,
				`group1.group2.group3.bool=true`,
				`group1.group2.group3.time="1984-01-02 03:04:05"`,
			}, " "),
		)
	})
}

func (s *FormatterStructuredTestSuite) TestQuoting() {
	s.logger().Info(
		"Test message",
		"key", "value",
		"key with spaces", "value with spaces",
		`key"with"quotes`, `value"with"quotes`,
		"key\nwith\nnewlines", "value\nwith\nnewlines",
		slog.Group("group name", slog.String("key", "value")),
	)

	s.checkNextEntry(
		"INFO",
		strings.Join([]string{
			`msg="Test message"`,
			`key="value"`,
			`"key with spaces"="value with spaces"`,
			`"key\"with\"quotes"="value\"with\"quotes"`,
			`"key\nwith\nnewlines"="value\nwith\nnewlines"`,
			`"group name.key"="value"`,
		}, " "),
	)
}

func (s *FormatterStructuredTestSuite) TestSpecialFields() {
	s.logger().Info(
		"Test message",
		"stack", "stack value\nwith new lines",
		"key1", "value1",
		"error", errors.New("error value"),
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
		"INFO",
		strings.Join([]string{
			`msg="Test message"`,
			`key1="value1"`,
			`key2="value2"`,
			`key3="value3"`,
			`group.stack="stack in group"`,
			`group.error="error in group"`,
			`group.source="source in group"`,
			`error="error value"`,
			`source="source value"`,
			`stack="stack value\nwith new lines"`,
		}, " "),
	)
}

func (s *FormatterStructuredTestSuite) checkNextEntry(lvl, msg string) {
	str, err := s.buf().ReadString('\n')
	s.Require().NoError(err)

	const timeLen = 19 + 7  // +7 for key, separator, and quotes
	lvlLen := len(lvl) + 10 // +10 for key, separator, quotes, and spaces
	prefixLen := timeLen + lvlLen

	s.Require().GreaterOrEqual(len(str), prefixLen)

	timePart := str[:timeLen]
	levelPart := str[timeLen:prefixLen]

	now := time.Now()
	t, err := time.ParseInLocation(time.DateTime, timePart[6:timeLen-1], now.Location())
	s.Require().NoError(err)
	s.Require().WithinDuration(time.Now(), t, time.Minute)

	s.Equal(` level="`+lvl+`" `, levelPart)

	// Check the message
	s.Equal(msg+"\n", str[prefixLen:])
}

func TestFormatterStructured(t *testing.T) {
	suite.Run(t, new(FormatterStructuredTestSuite))
}
