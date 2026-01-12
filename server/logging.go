package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

func LogRequest(reqID string, r *http.Request) {
	path := r.RequestURI

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	slog.Info(
		fmt.Sprintf("Started %s", path),
		"request_id", reqID,
		"method", r.Method,
		"client_ip", clientIP,
	)
}

func LogResponse(
	reqID string,
	r *http.Request,
	status int,
	err errctx.Error,
	additional ...slog.Attr,
) {
	var level slog.Level

	switch {
	case status >= 500 || (err != nil && err.StatusCode() >= 500):
		level = slog.LevelError
	case status >= 400:
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	attrs := []slog.Attr{
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.Int("status", status),
		slog.String("client_ip", clientIP),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))

		if level >= slog.LevelError {
			if stack := err.FormatStack(); len(stack) > 0 {
				attrs = append(attrs, slog.String("stack", stack))
			}
		}
	}

	attrs = append(attrs, additional...)

	slog.LogAttrs(
		context.Background(),
		level,
		fmt.Sprintf("Completed in %s %s", requestStartedAt(r.Context()), r.RequestURI),
		attrs...,
	)
}
