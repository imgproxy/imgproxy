package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type FormatterJsonTestSuite struct {
	testutil.LazySuite

	buf     testutil.LazyObj[*bytes.Buffer]
	config  testutil.LazyObj[*Config]
	handler testutil.LazyObj[*Handler]
	logger  testutil.LazyObj[*slog.Logger]
}

func (s *FormatterJsonTestSuite) SetupTest() {
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
			cfg.Format = FormatJSON
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

func (s *FormatterJsonTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *FormatterJsonTestSuite) TestLevel() {
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
				s.checkNextEntry(entry.levelName, map[string]any{
					"msg": entry.message,
				})
			}
		})
	}
}

func (s *FormatterJsonTestSuite) TestAttributes() {
	s.logger().Info(
		"Test message",
		slog.String("string", "value"),
		slog.Int("int", -100),
		slog.Uint64("uint64", 200),
		slog.Float64("float64", 3.14),
		slog.Bool("bool", true),
		slog.Time("timearg", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
		slog.Duration("duration", time.Minute),
		slog.Any("err", errors.New("error value")),
		slog.Any("any", struct {
			Field1 string
			Field2 int
		}{"value", 42}),
	)

	s.checkNextEntry(
		"INFO",
		map[string]any{
			"msg":      "Test message",
			"string":   "value",
			"int":      -100.0,
			"uint64":   200.0,
			"float64":  3.14,
			"bool":     true,
			"timearg":  "1984-01-02T03:04:05Z",
			"duration": float64(time.Minute),
			"err":      "error value",
			"any":      map[string]any{"Field1": "value", "Field2": 42.0},
		},
	)
}

func (s *FormatterJsonTestSuite) TestGroups() {
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
				slog.Time("timearg", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
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
			map[string]any{
				"msg":    "Test message",
				"string": "value",
				"int":    -100.0,
				"group1": map[string]any{
					"uint64":  200.0,
					"float64": 3.14,
					"group2": map[string]any{
						"group3": map[string]any{
							"bool":    true,
							"timearg": "1984-01-02T03:04:05Z",
						},
						"group4": map[string]any{
							"duration": float64(time.Minute),
							"any":      map[string]any{"Field1": "value", "Field2": 42.0},
						},
					},
				},
			},
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
				slog.Time("timearg", time.Date(1984, 1, 2, 3, 4, 5, 6, time.UTC)),
			)).
			With(slog.Group("empty_group")).
			WithGroup("group4").
			WithGroup("group5").
			Info("Test message")

		s.checkNextEntry(
			"INFO",
			map[string]any{
				"msg":    "Test message",
				"string": "value",
				"int":    -100.0,
				"group1": map[string]any{
					"uint64":  200.0,
					"float64": 3.14,
					"group2": map[string]any{
						"group3": map[string]any{
							"bool":    true,
							"timearg": "1984-01-02T03:04:05Z",
						},
					},
				},
			},
		)
	})
}

func (s *FormatterJsonTestSuite) TestEscaping() {
	s.logger().Info(
		"Test message",
		"key", "value",
		"key 1", "value 1",
		`"key"`, `"value"`,
		`<key>`, `<value>`,
		"\nkey\n", "\nvalue\n",
		slog.Group("group name", slog.String("key", "value")),
	)

	s.checkNextEntry(
		"INFO",
		map[string]any{
			"msg":        "Test message",
			"key":        "value",
			"key 1":      "value 1",
			`"key"`:      `"value"`,
			`<key>`:      `<value>`,
			"\nkey\n":    "\nvalue\n",
			"group name": map[string]any{"key": "value"},
		},
	)
}

func (s *FormatterJsonTestSuite) TestSpecialFields() {
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

	expectedJSON := strings.Join([]string{
		`"msg":"Test message",`,
		`"key1":"value1",`,
		`"key2":"value2",`,
		`"key3":"value3",`,
		`"group":{`,
		`"stack":"stack in group",`,
		`"error":"error in group",`,
		`"source":"source in group"`,
		`},`,
		`"error":"test",`,
		`"error_docs_url":"http://example.com",`,
		`"source":"source value",`,
		`"stack":"stack value\nwith new lines"`,
		"}\n",
	}, "")

	s.Require().Contains(s.buf().String(), expectedJSON)
}

func (s *FormatterJsonTestSuite) checkNextEntry(lvl string, msg map[string]any) {
	str, err := s.buf().ReadString('\n')
	s.Require().NoError(err)

	var parsed map[string]any
	err = json.Unmarshal([]byte(str), &parsed)
	s.Require().NoError(err)

	s.Require().IsType("", parsed["time"])
	s.Require().IsType("", parsed["level"])

	now := time.Now()
	t, err := time.ParseInLocation(time.RFC3339, parsed["time"].(string), now.Location())
	s.Require().NoError(err)
	s.Require().WithinDuration(time.Now(), t, time.Minute)

	s.Equal(lvl, parsed["level"].(string))

	// Remove time and level as they are not included in `msg`
	delete(parsed, "time")
	delete(parsed, "level")

	// Check the message
	s.Equal(msg, parsed)
}

func TestFormatterJson(t *testing.T) {
	suite.Run(t, new(FormatterJsonTestSuite))
}
