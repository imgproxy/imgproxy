package airbrake

import (
	"net/http"
	"strings"

	"github.com/airbrake/gobrake/v5"
	"github.com/imgproxy/imgproxy/v3/errctx"
)

var (
	metaReplacer = strings.NewReplacer(" ", "-")
)

type reporter struct {
	notifier *gobrake.Notifier
}

func New(config *Config) (*reporter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if len(config.ProjectKey) == 0 {
		return nil, nil
	}

	notifier := gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:   int64(config.ProjectID),
		ProjectKey:  config.ProjectKey,
		Environment: config.Env,
	})

	return &reporter{notifier}, nil
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
	notice := r.notifier.Notice(err, req, 2)

	// imgproxy may wrap errors using errctx.WrappedError to add context, so Airbrake
	// would report the error type as *errctx.WrappedError.
	//
	// To avoid this, we set the correct error type here.
	notice.Errors[0].Type = errctx.ErrorType(err)

	for k, v := range meta {
		key := metaReplacer.Replace(strings.ToLower(k))
		notice.Context[key] = v
	}

	r.notifier.SendNoticeAsync(notice)
}

func (r *reporter) Close() {
	r.notifier.Close()
}
