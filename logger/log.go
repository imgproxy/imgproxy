package logger

import (
	"fmt"
	"os"

	logrus "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
)

func Init() error {
	logrus.SetOutput(os.Stdout)

	logFormat := "pretty"
	logLevel := "info"

	configurators.String(&logFormat, "IMGPROXY_LOG_FORMAT")
	configurators.String(&logLevel, "IMGPROXY_LOG_LEVEL")

	switch logFormat {
	case "structured":
		logrus.SetFormatter(&structuredFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(newPrettyFormatter())
	}

	levelLogLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		levelLogLevel = logrus.InfoLevel
	}

	logrus.SetLevel(levelLogLevel)

	if isSyslogEnabled() {
		slHook, err := newSyslogHook()
		if err != nil {
			return fmt.Errorf("Unable to connect to syslog daemon: %s", err)
		}

		logrus.AddHook(slHook)
	}

	return nil
}
