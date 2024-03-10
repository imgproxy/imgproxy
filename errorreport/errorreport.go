package errorreport

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errorreport/airbrake"
	"github.com/imgproxy/imgproxy/v3/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v3/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v3/errorreport/sentry"
)

type metaCtxKey struct{}

func Init() {
	bugsnag.Init()
	honeybadger.Init()
	sentry.Init()
	airbrake.Init()
}

func StartRequest(req *http.Request) context.Context {
	meta := make(map[string]any)
	return context.WithValue(req.Context(), metaCtxKey{}, meta)
}

func SetMetadata(req *http.Request, key string, value any) {
	meta, ok := req.Context().Value(metaCtxKey{}).(map[string]any)
	if !ok || meta == nil {
		return
	}

	meta[key] = value
}

func Report(err error, req *http.Request) {
	meta, ok := req.Context().Value(metaCtxKey{}).(map[string]any)
	if !ok {
		meta = nil
	}

	bugsnag.Report(err, req, meta)
	honeybadger.Report(err, req, meta)
	sentry.Report(err, req, meta)
	airbrake.Report(err, req, meta)
}

func Close() {
	airbrake.Close()
}
