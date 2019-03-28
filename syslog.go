package main

import (
	"log"
	"log/syslog"
)

var (
	syslogWriter *syslog.Writer
	syslogLevel  syslog.Priority
)

var syslogLevels = map[string]syslog.Priority{
	"crit":    syslog.LOG_CRIT,
	"error":   syslog.LOG_ERR,
	"warning": syslog.LOG_WARNING,
	"notice":  syslog.LOG_NOTICE,
}

func initSyslog() {
	var (
		err error

		enabled       bool
		network, addr string
	)

	boolEnvConfig(&enabled, "IMGPROXY_SYSLOG_ENABLE")

	if !enabled {
		return
	}

	strEnvConfig(&network, "IMGPROXY_SYSLOG_NETWORK")
	strEnvConfig(&addr, "IMGPROXY_SYSLOG_ADDRESS")

	tag := "imgproxy"
	strEnvConfig(&tag, "IMGPROXY_SYSLOG_TAG")

	syslogWriter, err = syslog.Dial(network, addr, syslog.LOG_NOTICE, tag)

	if err != nil {
		log.Fatalf("Can't connect to syslog: %s", err)
	}

	levelStr := "notice"
	strEnvConfig(&levelStr, "IMGPROXY_SYSLOG_LEVEL")

	if l, ok := syslogLevels[levelStr]; ok {
		syslogLevel = l
	} else {
		syslogLevel = syslog.LOG_NOTICE
		logWarning("Syslog level '%s' is invalid, 'notice' is used", levelStr)
	}
}
