package router

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v2/ierrors"
	log "github.com/sirupsen/logrus"
)

func LogRequest(reqID string, r *http.Request) {
	path := r.RequestURI

	log.WithFields(log.Fields{
		"request_id": reqID,
		"method":     r.Method,
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

	fields := log.Fields{
		"request_id": reqID,
		"method":     r.Method,
		"status":     status,
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
