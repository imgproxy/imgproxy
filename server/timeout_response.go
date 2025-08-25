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
	timeout    time.Duration
}

// newTimeoutResponse creates a new timeoutResponse
func newTimeoutResponse(rw http.ResponseWriter, timeout time.Duration) http.ResponseWriter {
	return &timeoutResponse{
		ResponseWriter: rw,
		controller:     http.NewResponseController(rw),
		timeout:        timeout,
	}
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

// withWriteDeadline executes a Write* function with a deadline
func (rw *timeoutResponse) withWriteDeadline(f func()) {
	deadline := time.Now().Add(rw.timeout)

	// Set write deadline
	rw.controller.SetWriteDeadline(deadline)

	// Reset write deadline after method has finished
	defer rw.controller.SetWriteDeadline(time.Time{})
	f()
}
