package errorreport

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/airbrake"
	"github.com/imgproxy/imgproxy/v4/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v4/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v4/errorreport/sentry"
	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/server/meta"
)

// reporter is an interface that all error reporters must implement.
// most of our reporters are singletons, so in most cases close is noop.
type reporter interface {
	Report(err errctx.Error, req *http.Request, meta map[string]any)
	Close()
}

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

// Report reports an error to all configured reporters with the request and its metadata.
// The request is read from ctx via meta.RequestFromContext and may be nil for errors that
// don't originate from an HTTP request (e.g. background work); in that case no HTTP-specific
// data (headers, URL) is attached, but ctx-scoped metadata still is.
func (r *Reporter) Report(ctx context.Context, err errctx.Error) {
	req := meta.RequestFromContext(ctx)

	extra := make(map[string]any)

	if m := meta.FromContext(ctx); m != nil {
		extra = m.Map(func(key string, value any) (string, any) {
			switch {
			case key == meta.KeyOptions:
				if o, ok := value.(*options.Options); ok {
					return key, o.NestedMap()
				}

			case meta.IsSimpleValue(value):
				return key, value
			}

			return "", nil
		})
	}

	if url := err.DocsURL(); url != "" {
		extra["Documentation URL"] = url
	}

	for _, reporter := range r.reporters {
		reporter.Report(err, req, extra)
	}
}

// Close closes all reporters
func (r *Reporter) Close() {
	for _, reporter := range r.reporters {
		reporter.Close()
	}
}
