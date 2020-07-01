package main

import (
	"fmt"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	syslogLevels = map[string]logrus.Level{
		"crit":    logrus.FatalLevel,
		"error":   logrus.ErrorLevel,
		"warning": logrus.WarnLevel,
		"info":    logrus.InfoLevel,
	}
)

type syslogHook struct {
	writer    *syslog.Writer
	levels    []logrus.Level
	formatter logrus.Formatter
}

func isSyslogEnabled() (enabled bool) {
	boolEnvConfig(&enabled, "IMGPROXY_SYSLOG_ENABLE")
	return
}

func newSyslogHook() (*syslogHook, error) {
	var (
		network, addr string
		level         logrus.Level

		tag      = "imgproxy"
		levelStr = "notice"
	)

	strEnvConfig(&network, "IMGPROXY_SYSLOG_NETWORK")
	strEnvConfig(&addr, "IMGPROXY_SYSLOG_ADDRESS")
	strEnvConfig(&tag, "IMGPROXY_SYSLOG_TAG")
	strEnvConfig(&levelStr, "IMGPROXY_SYSLOG_LEVEL")

	if l, ok := syslogLevels[levelStr]; ok {
		level = l
	} else {
		level = logrus.InfoLevel
		logWarning("Syslog level '%s' is invalid, 'info' is used", levelStr)
	}

	w, err := syslog.Dial(network, addr, syslog.LOG_NOTICE, tag)

	return &syslogHook{
		writer:    w,
		levels:    logrus.AllLevels[:int(level)+1],
		formatter: &logStructuredFormatter{},
	}, err
}

func (hook *syslogHook) Fire(entry *logrus.Entry) error {
	line, err := hook.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v\n", err)
		return err
	}

	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel:
		return hook.writer.Crit(string(line))
	case logrus.ErrorLevel:
		return hook.writer.Err(string(line))
	case logrus.WarnLevel:
		return hook.writer.Warning(string(line))
	case logrus.InfoLevel:
		return hook.writer.Info(string(line))
	default:
		return nil
	}
}

func (hook *syslogHook) Levels() []logrus.Level {
	return hook.levels
}
