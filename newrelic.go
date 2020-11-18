package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var (
	newRelicEnabled = false

	newRelicApp *newrelic.Application

	newRelicTransactionCtxKey = ctxKey("newRelicTransaction")
)

func initNewrelic() error {
	if len(conf.NewRelicKey) == 0 {
		return nil
	}

	name := conf.NewRelicAppName
	if len(name) == 0 {
		name = "imgproxy"
	}

	var err error

	newRelicApp, err = newrelic.NewApplication(
		newrelic.ConfigAppName(name),
		newrelic.ConfigLicense(conf.NewRelicKey),
	)

	if err != nil {
		return fmt.Errorf("Can't init New Relic agent: %s", err)
	}

	newRelicEnabled = true

	return nil
}

func startNewRelicTransaction(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc, http.ResponseWriter) {
	txn := newRelicApp.StartTransaction("request")
	txn.SetWebRequestHTTP(r)
	newRw := txn.SetWebResponse(rw)
	cancel := func() { txn.End() }
	return context.WithValue(ctx, newRelicTransactionCtxKey, txn), cancel, newRw
}

func startNewRelicSegment(ctx context.Context, name string) context.CancelFunc {
	txn := ctx.Value(newRelicTransactionCtxKey).(*newrelic.Transaction)
	segment := txn.StartSegment(name)
	return func() { segment.End() }
}

func sendErrorToNewRelic(ctx context.Context, err error) {
	txn := ctx.Value(newRelicTransactionCtxKey).(*newrelic.Transaction)
	txn.NoticeError(err)
}

func sendTimeoutToNewRelic(ctx context.Context, d time.Duration) {
	txn := ctx.Value(newRelicTransactionCtxKey).(*newrelic.Transaction)
	txn.NoticeError(newrelic.Error{
		Message: "Timeout",
		Class:   "Timeout",
		Attributes: map[string]interface{}{
			"time": d.Seconds(),
		},
	})
}
