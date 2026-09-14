package testutil

import (
	"bytes"
	"io"
	"net/http"
)

type TestTransportHookFunc func(req *http.Request, res *http.Response)

// TestRoundTripper is a custom http.RoundTripper for testing purposes.
// It always returns the configured response, regardless of the request.
// You can register it as the transport for an HTTP client, or register
// it for a specific protocol in [http.Transport] or [fetcher.Fetcher].
type TestRoundTripper struct {
	status int
	data   []byte
	header http.Header
	hook   TestTransportHookFunc
}

// NewTestRoundTripper creates a new TestRoundTripper.
func NewTestRoundTripper() *TestRoundTripper {
	return &TestRoundTripper{
		status: http.StatusOK,
		header: make(http.Header),
	}
}

// RoundTrip implements the http.RoundTripper interface for TestRoundTripper.
// It returns a response with the configured status, headers, and body,
// and invokes the hook if set.
func (rt *TestRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res := &http.Response{
		StatusCode:    rt.status,
		Header:        rt.header.Clone(),
		Body:          io.NopCloser(bytes.NewReader(rt.data)),
		ContentLength: int64(len(rt.data)),
	}

	if rt.hook != nil {
		rt.hook(req, res)
	}

	return res, nil
}

// SetStatusCode sets the HTTP status code for the TestRoundTripper.
func (rt *TestRoundTripper) SetStatusCode(status int) *TestRoundTripper {
	rt.status = status
	return rt
}

// SetBody sets the body that will be returned by the TestRoundTripper.
func (rt *TestRoundTripper) SetBody(data []byte) *TestRoundTripper {
	rt.data = data
	return rt
}

// SetHeaders adds headers that will be returned by the TestRoundTripper.
// Odd arguments are treated as keys, even arguments as values.
func (rt *TestRoundTripper) SetHeaders(kv ...string) *TestRoundTripper {
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		value := kv[i+1]
		rt.header.Set(key, value)
	}

	return rt
}

// SetHook sets a function that will be called on each request for the TestRoundTripper.
func (rt *TestRoundTripper) SetHook(f TestTransportHookFunc) *TestRoundTripper {
	rt.hook = f
	return rt
}
