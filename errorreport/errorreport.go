package errorreport

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errorreport/airbrake"
	"github.com/imgproxy/imgproxy/v3/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v3/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v3/errorreport/sentry"
)

func Init() {
	bugsnag.Init()
	honeybadger.Init()
	sentry.Init()
	airbrake.Init()
}

func Report(err error, req *http.Request) {
	bugsnag.Report(err, req)
	honeybadger.Report(err, req)
	sentry.Report(err, req)
	airbrake.Report(err, req)
}

func Close() {
	airbrake.Close()
}
