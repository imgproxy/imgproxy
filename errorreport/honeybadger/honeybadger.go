package honeybadger

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/honeybadger-io/honeybadger-go"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var (
	metaReplacer = strings.NewReplacer("-", "_", " ", "_")
)

type reporter struct{}

func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.Key) == 0 {
		return nil, nil
	}

	honeybadger.Configure(honeybadger.Configuration{
		APIKey: config.Key,
		Env:    config.Env,
	})

	return &reporter{}, nil
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

	hbErr := honeybadger.NewError(err)

	if e, ok := err.(*ierrors.Error); ok {
		hbErr.Class = reflect.TypeOf(e.Unwrap()).String()
	}

	honeybadger.Notify(hbErr, req.URL, extra)
}

func (r *reporter) Close() {
	// noop
}
