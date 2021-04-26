package logger

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	logrus "github.com/sirupsen/logrus"
)

var (
	logKeysPriorities = map[string]int{
		"request_id": 3,
		"method":     2,
		"status":     1,
		"error":      -1,
		"stack":      -2,
	}

	logQuotingRe = regexp.MustCompile(`^[a-zA-Z0-9\-._/@^+]+$`)
)

type logKeys []string

func (p logKeys) Len() int           { return len(p) }
func (p logKeys) Less(i, j int) bool { return logKeysPriorities[p[i]] > logKeysPriorities[p[j]] }
func (p logKeys) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type prettyFormatter struct {
	levelFormat string
}

func newPrettyFormatter() *prettyFormatter {
	f := new(prettyFormatter)

	levelLenMax := 0
	for _, level := range logrus.AllLevels {
		levelLen := utf8.RuneCount([]byte(level.String()))
		if levelLen > levelLenMax {
			levelLenMax = levelLen
		}
	}

	f.levelFormat = fmt.Sprintf("%%-%ds", levelLenMax)

	return f
}

func (f *prettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if k != "stack" {
			keys = append(keys, k)
		}
	}

	sort.Sort(logKeys(keys))

	levelColor := 36
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 37
	case logrus.WarnLevel:
		levelColor = 33
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 31
	}

	levelText := fmt.Sprintf(f.levelFormat, strings.ToUpper(entry.Level.String()))
	msg := strings.TrimSuffix(entry.Message, "\n")

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = new(bytes.Buffer)
	}

	fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m [%s] %s ", levelColor, levelText, entry.Time.Format(time.RFC3339), msg)

	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " \x1b[1m%s\x1b[0m=", k)
		f.appendValue(b, v)
	}

	b.WriteByte('\n')

	if stack, ok := entry.Data["stack"]; ok {
		fmt.Fprintln(b, stack)
	}

	return b.Bytes(), nil
}

func (f *prettyFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	strValue, ok := value.(string)
	if !ok {
		strValue = fmt.Sprint(value)
	}

	if logQuotingRe.MatchString(strValue) {
		b.WriteString(strValue)
	} else {
		fmt.Fprintf(b, "%q", strValue)
	}
}

type structuredFormatter struct{}

func (f *structuredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	sort.Sort(logKeys(keys))

	msg := strings.TrimSuffix(entry.Message, "\n")

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = new(bytes.Buffer)
	}

	f.appendKeyValue(b, "time", entry.Time.Format(time.RFC3339))
	f.appendKeyValue(b, "level", entry.Level.String())
	f.appendKeyValue(b, "message", msg)

	for _, k := range keys {
		f.appendKeyValue(b, k, entry.Data[k])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *structuredFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() != 0 {
		b.WriteByte(' ')
	}

	strValue, ok := value.(string)
	if !ok {
		strValue = fmt.Sprint(value)
	}

	fmt.Fprintf(b, "%s=%q", key, strValue)
}
