package honeybadger

import (
	"net/http"
	"strings"

	"github.com/honeybadger-io/honeybadger-go"
	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	enabled bool

	headersReplacer = strings.NewReplacer("-", "_")
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

func Report(err error, req *http.Request) {
	if enabled {
		headers := make(honeybadger.CGIData)

		for k, v := range req.Header {
			key := "HTTP_" + headersReplacer.Replace(strings.ToUpper(k))
			headers[key] = v[0]
		}

		honeybadger.Notify(err, req.URL, headers)
	}
}
