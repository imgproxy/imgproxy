package newrelic

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type transactionCtxKey struct{}

var (
	enabled = false

	newRelicApp *newrelic.Application
)

func Init() error {
	if len(config.NewRelicKey) == 0 {
		return nil
	}

	name := config.NewRelicAppName
	if len(name) == 0 {
		name = "imgproxy"
	}

	var err error

	newRelicApp, err = newrelic.NewApplication(
		newrelic.ConfigAppName(name),
		newrelic.ConfigLicense(config.NewRelicKey),
	)

	if err != nil {
		return fmt.Errorf("Can't init New Relic agent: %s", err)
	}

	enabled = true

	return nil
}

func Enabled() bool {
	return enabled
}

func StartTransaction(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	if !enabled {
		return ctx, func() {}, rw
	}

	txn := newRelicApp.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return context.WithValue(ctx, transactionCtxKey{}, txn), cancel, newRw
}

func StartSegment(ctx context.Context, name string) context.CancelFunc {
	if !enabled {
		return func() {}
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		segment := txn.StartSegment(name)
		return func() { segment.End() }
	}

	return func() {}
}

func SendError(ctx context.Context, err error) {
	if !enabled {
		return
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		txn.NoticeError(err)
	}
}

func SendTimeout(ctx context.Context, d time.Duration) {
	if !enabled {
		return
	}

	if txn, ok := ctx.Value(transactionCtxKey{}).(*newrelic.Transaction); ok {
		txn.NoticeError(newrelic.Error{
			Message: "Timeout",
			Class:   "Timeout",
			Attributes: map[string]interface{}{
				"time": d.Seconds(),
			},
		})
	}
}
