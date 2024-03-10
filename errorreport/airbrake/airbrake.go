package airbrake

import (
	"net/http"
	"strings"

	"github.com/airbrake/gobrake/v5"

	"github.com/imgproxy/imgproxy/v3/config"
)

var (
	notifier *gobrake.Notifier

	metaReplacer = strings.NewReplacer(" ", "-")
)

func Init() {
	if len(config.AirbrakeProjecKey) > 0 {
		notifier = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
			ProjectId:   int64(config.AirbrakeProjecID),
			ProjectKey:  config.AirbrakeProjecKey,
			Environment: config.AirbrakeEnv,
		})
	}
}

func Report(err error, req *http.Request, meta map[string]any) {
	if notifier == nil {
		return
	}

	notice := notifier.Notice(err, req, 2)

	for k, v := range meta {
		key := metaReplacer.Replace(strings.ToLower(k))
		notice.Context[key] = v
	}

	notifier.SendNoticeAsync(notice)
}

func Close() {
	if notifier != nil {
		notifier.Close()
	}
}
