package airbrake

import (
	"net/http"

	"github.com/airbrake/gobrake/v5"
	"github.com/imgproxy/imgproxy/v3/config"
)

var notifier *gobrake.Notifier

func Init() {
	if len(config.AirbrakeProjecKey) > 0 {
		notifier = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
			ProjectId:   int64(config.AirbrakeProjecID),
			ProjectKey:  config.AirbrakeProjecKey,
			Environment: config.AirbrakeEnv,
		})
	}
}

func Report(err error, req *http.Request) {
	if notifier != nil {
		notifier.Notify(err, req)
	}
}

func Close() {
	if notifier != nil {
		notifier.Close()
	}
}
