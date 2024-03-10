package honeybadger

import (
	"net/http"
	"strings"

	"github.com/honeybadger-io/honeybadger-go"

	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	enabled bool

	metaReplacer = strings.NewReplacer("-", "_", " ", "_")
)

func Init() {
	if len(config.HoneybadgerKey) > 0 {
		honeybadger.Configure(honeybadger.Configuration{
			APIKey: config.HoneybadgerKey,
			Env:    config.HoneybadgerEnv,
		})
		enabled = true
	}
}

func Report(err error, req *http.Request, meta map[string]any) {
	if !enabled {
		return
	}

	extra := make(honeybadger.CGIData, len(req.Header)+len(meta))

	for k, v := range req.Header {
		key := "HTTP_" + metaReplacer.Replace(strings.ToUpper(k))
		extra[key] = v[0]
	}

	for k, v := range meta {
		key := metaReplacer.Replace(strings.ToUpper(k))
		extra[key] = v
	}

	honeybadger.Notify(err, req.URL, extra)
}
