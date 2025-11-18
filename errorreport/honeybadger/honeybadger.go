package honeybadger

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/honeybadger-io/honeybadger-go"
	"github.com/imgproxy/imgproxy/v3/errctx"
)

var (
	metaReplacer = strings.NewReplacer("-", "_", " ", "_")
)

type reporter struct {
	client *honeybadger.Client
}

func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.Key) == 0 {
		return nil, nil
	}

	client := honeybadger.New(honeybadger.Configuration{
		APIKey: config.Key,
		Env:    config.Env,
	})

	return &reporter{client: client}, nil
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
	extra := make(honeybadger.CGIData, len(req.Header)+len(meta))

	for k, v := range req.Header {
		key := "HTTP_" + metaReplacer.Replace(strings.ToUpper(k))
		extra[key] = v[0]
	}

	for k, v := range meta {
		key := metaReplacer.Replace(strings.ToUpper(k))
		extra[key] = v
	}

	// imgproxy may wrap errors using errctx.WrappedError to add context, so Honeybadger
	// would report the error type as *errctx.WrappedError.
	//
	// To avoid this, we provide error class information explicitly.
	errClass := honeybadger.ErrorClass{Name: errctx.ErrorType(err)}

	if _, repErr := r.client.Notify(err, errClass, req.URL, extra); repErr != nil {
		slog.Warn("Failed to report error to Honeybadger", "error", repErr)
	}
}

func (r *reporter) Close() {
	r.client.Flush()
}
