package router

import (
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

type timeoutResponse struct {
	http.ResponseWriter
	controller *http.ResponseController
}

func newTimeoutResponse(rw http.ResponseWriter) http.ResponseWriter {
	return &timeoutResponse{
		ResponseWriter: rw,
		controller:     http.NewResponseController(rw),
	}
}

func (rw *timeoutResponse) WriteHeader(statusCode int) {
	rw.withWriteDeadline(func() {
		rw.ResponseWriter.WriteHeader(statusCode)
	})
}

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

func (rw *timeoutResponse) withWriteDeadline(f func()) {
	rw.controller.SetWriteDeadline(time.Now().Add(time.Duration(config.WriteResponseTimeout) * time.Second))
	defer rw.controller.SetWriteDeadline(time.Time{})
	f()
}
