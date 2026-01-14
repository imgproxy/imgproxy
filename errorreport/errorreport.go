package errorreport

import (
	"context"
	"maps"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/errorreport/airbrake"
	"github.com/imgproxy/imgproxy/v3/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v3/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v3/errorreport/sentry"
)

// reporter is an interface that all error reporters must implement.
// most of our reporters are singletons, so in most cases close is noop.
type reporter interface {
	Report(err errctx.Error, req *http.Request, meta map[string]any)
	Close()
}

// metaCtxKey is the context.Context key for request metadata
type metaCtxKey struct{}

type Reporter struct {
	// initialized reporters
	reporters []reporter
}

// New initializes all configured error reporters and returns a Reporter instance.
func New(config *Config) (*Reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	reporters := make([]reporter, 0)

	if r, err := bugsnag.New(&config.Bugsnag); err != nil {
		return nil, err
	} else if r != nil {
		reporters = append(reporters, r)
	}

	if r, err := honeybadger.New(&config.Honeybadger); err != nil {
		return nil, err
	} else if r != nil {
		reporters = append(reporters, r)
	}

	if r, err := sentry.New(&config.Sentry); err != nil {
		return nil, err
	} else if r != nil {
		reporters = append(reporters, r)
	}

	if r, err := airbrake.New(&config.Airbrake); err != nil {
		return nil, err
	} else if r != nil {
		reporters = append(reporters, r)
	}

	return &Reporter{
		reporters: reporters,
	}, nil
}

// StartRequest initializes metadata storage in the request context.
func StartRequest(req *http.Request) context.Context {
	meta := make(map[string]any)
	return context.WithValue(req.Context(), metaCtxKey{}, meta)
}

// SetMetadata sets a metadata key-value pair in the request context.
func SetMetadata(req *http.Request, key string, value any) {
	meta, ok := req.Context().Value(metaCtxKey{}).(map[string]any)
	if !ok || meta == nil {
		return
	}

	meta[key] = value
}

// Report reports an error to all configured reporters with the request and its metadata.
func (r *Reporter) Report(err errctx.Error, req *http.Request) {
	meta, ok := req.Context().Value(metaCtxKey{}).(map[string]any)
	if !ok {
		meta = nil
	}

	if url := err.DocsURL(); url != "" {
		meta = maps.Clone(meta)
		meta["Documentation URL"] = url
	}

	for _, reporter := range r.reporters {
		reporter.Report(err, req, meta)
	}
}

// Close closes all reporters
func (r *Reporter) Close() {
	for _, reporter := range r.reporters {
		reporter.Close()
	}
}
