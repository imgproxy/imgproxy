package main

import (
	"fmt"
	"log"
	"log/syslog"
	"net/http"
)

const (
	logRequestFmt        = "[%s] %s: %s"
	logRequestSyslogFmt  = "REQUEST [%s] %s: %s"
	logResponseFmt       = "[%s] |\033[7;%dm %d \033[0m| %s"
	logResponseSyslogFmt = "RESPONSE [%s] | %d | %s"
	logWarningFmt        = "\033[1;33m[WARNING]\033[0m %s"
	logWarningSyslogFmt  = "WARNING %s"
	logFatalSyslogFmt    = "FATAL %s"
)

func logRequest(reqID string, r *http.Request) {
	path := r.URL.RequestURI()

	log.Printf(logRequestFmt, reqID, r.Method, path)

	if syslogWriter != nil {
		syslogWriter.Notice(fmt.Sprintf(logRequestSyslogFmt, reqID, r.Method, path))
	}
}

func logResponse(reqID string, status int, msg string) {
	var color int

	if status >= 500 {
		color = 31
	} else if status >= 400 {
		color = 33
	} else {
		color = 32
	}

	log.Printf(logResponseFmt, reqID, color, status, msg)

	if syslogWriter != nil {
		msg := fmt.Sprintf(logResponseSyslogFmt, reqID, status, msg)

		if status >= 500 {
			if syslogLevel >= syslog.LOG_ERR {
				syslogWriter.Err(msg)
			}
		} else if status >= 400 {
			if syslogLevel >= syslog.LOG_WARNING {
				syslogWriter.Warning(msg)
			}
		} else {
			if syslogLevel >= syslog.LOG_NOTICE {
				syslogWriter.Notice(msg)
			}
		}
	}
}

func logNotice(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)

	log.Print(msg)

	if syslogWriter != nil && syslogLevel >= syslog.LOG_NOTICE {
		syslogWriter.Notice(msg)
	}
}

func logWarning(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)

	log.Printf(logWarningFmt, msg)

	if syslogWriter != nil && syslogLevel >= syslog.LOG_WARNING {
		syslogWriter.Warning(fmt.Sprintf(logWarningSyslogFmt, msg))
	}
}

func logFatal(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)

	if syslogWriter != nil {
		syslogWriter.Crit(fmt.Sprintf(logFatalSyslogFmt, msg))
	}

	log.Fatal(msg)
}
