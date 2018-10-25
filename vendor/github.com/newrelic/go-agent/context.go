// +build go1.7

package newrelic

import (
	"context"
	"net/http"
)

type contextKeyType struct{}

var contextKey = contextKeyType(struct{}{})

// NewContext returns a new Context that carries the provided transcation.
func NewContext(ctx context.Context, txn Transaction) context.Context {
	return context.WithValue(ctx, contextKey, txn)
}

// FromContext returns the Transaction from the context if present, and nil
// otherwise.
func FromContext(ctx context.Context) Transaction {
	h, _ := ctx.Value(contextKey).(Transaction)
	return h
}

// RequestWithTransactionContext adds the transaction to the request's context.
func RequestWithTransactionContext(req *http.Request, txn Transaction) *http.Request {
	ctx := req.Context()
	ctx = NewContext(ctx, txn)
	return req.WithContext(ctx)
}
