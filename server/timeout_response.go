package server

import (
	"net/http"
	"time"
)

// timeoutResponse manages response writer with timeout. It has
// timeout on all write methods.
type timeoutResponse struct {
	http.ResponseWriter
	controller *http.ResponseController
	timeout    int
}

// newTimeoutResponse creates a new timeoutResponse
func newTimeoutResponse(rw http.ResponseWriter, timeout int) http.ResponseWriter {
	return &timeoutResponse{
		ResponseWriter: rw,
		controller:     http.NewResponseController(rw),
		timeout:        timeout,
	}
}

// WriteHeader implements http.ResponseWriter.WriteHeader
func (rw *timeoutResponse) WriteHeader(statusCode int) {
	rw.withWriteDeadline(func() {
		rw.ResponseWriter.WriteHeader(statusCode)
	})
}

// Write implements http.ResponseWriter.Write
func (rw *timeoutResponse) Write(b []byte) (int, error) {
	var (
		n   int
		err error
	)
	rw.withWriteDeadline(func() {
		n, err = rw.ResponseWriter.Write(b)
	})
	return n, err
}

// Header returns current HTTP headers
func (rw *timeoutResponse) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// withWriteDeadline executes a Write* function with a deadline
func (rw *timeoutResponse) withWriteDeadline(f func()) {
	deadline := time.Now().Add(time.Duration(rw.timeout) * time.Second)

	// Set write deadline
	rw.controller.SetWriteDeadline(deadline)

	// Reset write deadline after method has finished
	defer rw.controller.SetWriteDeadline(time.Time{})
	f()
}
