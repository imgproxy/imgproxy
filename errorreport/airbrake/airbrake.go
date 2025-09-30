package airbrake

import (
	"errors"
	"net/http"
	"strings"

	"github.com/airbrake/gobrake/v5"

	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_AIRBRAKE_PROJECT_ID  = env.Describe("IMGPROXY_AIRBRAKE_PROJECT_ID", "integer")
	IMGPROXY_AIRBRAKE_PROJECT_KEY = env.Describe("IMGPROXY_AIRBRAKE_PROJECT_KEY", "string")
	IMGPROXY_AIRBRAKE_ENV         = env.Describe("IMGPROXY_AIRBRAKE_ENV", "string")

	metaReplacer = strings.NewReplacer(" ", "-")
)

type reporter struct {
	notifier *gobrake.Notifier
}

func New() (*reporter, error) {
	var projectID int
	var projectKey string

	projectEnv := "production"

	err := errors.Join(
		env.Int(&projectID, IMGPROXY_AIRBRAKE_PROJECT_ID),
		env.String(&projectKey, IMGPROXY_AIRBRAKE_PROJECT_KEY),
		env.String(&projectEnv, IMGPROXY_AIRBRAKE_ENV),
	)

	if err != nil {
		return nil, err
	}

	if len(projectKey) == 0 {
		return nil, nil
	}

	notifier := gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:   int64(projectID),
		ProjectKey:  projectKey,
		Environment: projectEnv,
	})

	return &reporter{notifier}, nil
}

func (r *reporter) Report(err error, req *http.Request, meta map[string]any) {
	notice := r.notifier.Notice(err, req, 2)

	for k, v := range meta {
		key := metaReplacer.Replace(strings.ToLower(k))
		notice.Context[key] = v
	}

	r.notifier.SendNoticeAsync(notice)
}

func (r *reporter) Close() {
	r.notifier.Close()
}
