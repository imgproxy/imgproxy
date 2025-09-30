package honeybadger

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/honeybadger-io/honeybadger-go"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var (
	IMGPROXY_HONEYBADGER_KEY = env.Describe("IMGPROXY_HONEYBADGER_KEY", "string")
	IMGPROXY_HONEYBADGER_ENV = env.Describe("IMGPROXY_HONEYBADGER_ENV", "string")

	metaReplacer = strings.NewReplacer("-", "_", " ", "_")
)

type reporter struct{}

func New() (*reporter, error) {
	key := ""
	envir := "production"

	err := errors.Join(
		env.String(&key, IMGPROXY_HONEYBADGER_KEY),
		env.String(&envir, IMGPROXY_HONEYBADGER_ENV),
	)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, nil
	}

	honeybadger.Configure(honeybadger.Configuration{
		APIKey: key,
		Env:    envir,
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
