package logger

import (
	"fmt"
	"os"
	"strings"

	logrus "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
)

func init() {
	// Configure logrus so it can be used before Init().
	// Structured formatter is a compromise between JSON and pretty formatters.
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&structuredFormatter{})
}

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
	case "gcp":
		logrus.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				"level": "severity",
				"msg":   "message",
			},
		})
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

// Deprecated prints a deprecation warning message.
// If the IMGPROXY_FAIL_ON_DEPRECATION environment variable is truthy,
// it prints an error message and exits the program.
func Deprecated(deprecation, replacement string, additional ...string) {
	msg := fmt.Sprintf("%s is deprecated, use %s instead", deprecation, replacement)

	if len(additional) > 0 {
		msg += ". " + strings.Join(additional, ". ")
	}

	shouldFail := false
	configurators.Bool(&shouldFail, "IMGPROXY_FAIL_ON_DEPRECATION")

	if shouldFail {
		logrus.Fatal(msg)
	} else {
		logrus.Warning(msg)
	}
}
