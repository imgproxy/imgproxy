package main

import (
	"context"
	"net/http"
	"time"

	newrelic "github.com/newrelic/go-agent"
)

var (
	newRelicEnabled = false

	newRelicApp newrelic.Application

	newRelicTransactionCtxKey = ctxKey("newRelicTransaction")
)

func initNewrelic() {
	if len(conf.NewRelicKey) == 0 {
		return
	}

	name := conf.NewRelicAppName
	if len(name) == 0 {
		name = "imgproxy"
	}

	var err error

	config := newrelic.NewConfig(name, conf.NewRelicKey)
	newRelicApp, err = newrelic.NewApplication(config)

	if err != nil {
		logFatal("Can't init New Relic agent: %s", err)
	}

	newRelicEnabled = true
}

func startNewRelicTransaction(ctx context.Context, rw http.ResponseWriter, r *http.Request) (context.Context, context.CancelFunc) {
	txn := newRelicApp.StartTransaction("request", rw, r)
	cancel := func() { txn.End() }
	return context.WithValue(ctx, newRelicTransactionCtxKey, txn), cancel
}

func startNewRelicSegment(ctx context.Context, name string) context.CancelFunc {
	txn := ctx.Value(newRelicTransactionCtxKey).(newrelic.Transaction)
	segment := newrelic.StartSegment(txn, name)
	return func() { segment.End() }
}

func sendErrorToNewRelic(ctx context.Context, err error) {
	txn := ctx.Value(newRelicTransactionCtxKey).(newrelic.Transaction)
	txn.NoticeError(err)
}

func sendTimeoutToNewRelic(ctx context.Context, d time.Duration) {
	txn := ctx.Value(newRelicTransactionCtxKey).(newrelic.Transaction)
	txn.NoticeError(newrelic.Error{
		Message: "Timeout",
		Class:   "Timeout",
		Attributes: map[string]interface{}{
			"time": d.Seconds(),
		},
	})
}
