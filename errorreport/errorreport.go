package errorreport

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v2/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v2/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v2/errorreport/sentry"
)

func Init() {
	bugsnag.Init()
	honeybadger.Init()
	sentry.Init()
}

func Report(err error, req *http.Request) {
	bugsnag.Report(err, req)
	honeybadger.Report(err, req)
	sentry.Report(err, req)
}
