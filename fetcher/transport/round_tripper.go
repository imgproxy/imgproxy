package transport

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/storage"
)

// RoundTripper wraps storage with http.RoundTripper
type RoundTripper struct {
	http.RoundTripper

	storage        storage.Reader
	querySeparator string
}

// New creates a new RoundTripper
func NewRoundTripper(storage storage.Reader, querySeparator string) *RoundTripper {
	return &RoundTripper{
		storage:        storage,
		querySeparator: querySeparator,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t RoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Parse container and object name from the URL
	container, key, query := GetBucketAndKey(req.URL, t.querySeparator)

	// Call GetObject
	r, err := t.storage.GetObject(req.Context(), req.Header, container, key, query)
	if err != nil {
		return nil, err
	}

	return r.Response(req), nil
}
