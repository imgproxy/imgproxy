package newrelic

import "net/http"

// instrumentation.go contains helpers built on the lower level api.

// WrapHandle facilitates instrumentation of handlers registered with an
// http.ServeMux.  For example, to instrument this code:
//
//    http.Handle("/foo", fooHandler)
//
// Perform this replacement:
//
//    http.Handle(newrelic.WrapHandle(app, "/foo", fooHandler))
//
// The Transaction is passed to the handler in place of the original
// http.ResponseWriter, so it can be accessed using type assertion.
// For example, to rename the transaction:
//
//	// 'w' is the variable name of the http.ResponseWriter.
//	if txn, ok := w.(newrelic.Transaction); ok {
//		txn.SetName("other-name")
//	}
//
// The Transaction is added to the request's context, so it may be alternatively
// accessed like this:
//
//	// 'req' is the variable name of the *http.Request.
//	txn := newrelic.FromContext(req.Context())
//
// This function is safe to call if 'app' is nil.
func WrapHandle(app Application, pattern string, handler http.Handler) (string, http.Handler) {
	if app == nil {
		return pattern, handler
	}
	return pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		txn := app.StartTransaction(pattern, w, r)
		defer txn.End()

		r = RequestWithTransactionContext(r, txn)

		handler.ServeHTTP(txn, r)
	})
}

// WrapHandleFunc serves the same purpose as WrapHandle for functions registered
// with ServeMux.HandleFunc.
func WrapHandleFunc(app Application, pattern string, handler func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	p, h := WrapHandle(app, pattern, http.HandlerFunc(handler))
	return p, func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) }
}

// NewRoundTripper creates an http.RoundTripper to instrument external requests.
// The http.RoundTripper returned will create an external segment before
// delegating to the original RoundTripper provided (or http.DefaultTransport if
// none is provided).  If the Transaction parameter is nil, the RoundTripper
// will look for a Transaction in the request's context (using FromContext).
// This is STRONGLY recommended because it allows you to reuse the same client
// for multiple transactions.  Example use:
//
//   client := &http.Client{}
//   client.Transport = newrelic.NewRoundTripper(nil, client.Transport)
//   request, _ := http.NewRequest("GET", "http://example.com", nil)
//   request = newrelic.RequestWithTransactionContext(request, txn)
//   resp, err := client.Do(request)
//
func NewRoundTripper(txn Transaction, original http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		segment := StartExternalSegment(txn, request)

		if nil == original {
			original = http.DefaultTransport
		}
		response, err := original.RoundTrip(request)

		segment.Response = response
		segment.End()

		return response, err
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
