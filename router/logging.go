package router

import (
	"net"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	log "github.com/sirupsen/logrus"
)

func LogRequest(reqID string, r *http.Request) {
	path := r.RequestURI

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	log.WithFields(log.Fields{
		"request_id": reqID,
		"method":     r.Method,
		"client_ip":  clientIP,
	}).Infof("Started %s", path)
}

func LogResponse(reqID string, r *http.Request, status int, err *ierrors.Error, additional ...log.Fields) {
	var level log.Level

	switch {
	case status >= 500:
		level = log.ErrorLevel
	case status >= 400:
		level = log.WarnLevel
	default:
		level = log.InfoLevel
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	fields := log.Fields{
		"request_id": reqID,
		"method":     r.Method,
		"status":     status,
		"client_ip":  clientIP,
	}

	if err != nil {
		fields["error"] = err

		if stack := err.FormatStack(); len(stack) > 0 {
			fields["stack"] = stack
		}
	}

	for _, f := range additional {
		for k, v := range f {
			fields[k] = v
		}
	}

	log.WithFields(fields).Logf(
		level,
		"Completed in %s %s", ctxTime(r.Context()), r.RequestURI,
	)
}
