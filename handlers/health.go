package handlers

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var imgproxyIsRunningMsg = []byte("imgproxy is running")

// HealthHandler handles the health check requests
func HealthHandler(reqID string, rw http.ResponseWriter, r *http.Request) error {
	var (
		status int
		msg    []byte
		ierr   *ierrors.Error
	)

	if err := vips.Health(); err == nil {
		status = http.StatusOK
		msg = imgproxyIsRunningMsg
	} else {
		status = http.StatusInternalServerError
		msg = []byte("Error")
		ierr = ierrors.Wrap(err, 1)
	}

	if len(msg) == 0 {
		msg = []byte{' '}
	}

	// Log response only if something went wrong
	if ierr != nil {
		server.LogResponse(reqID, r, status, ierr)
	}

	rw.Header().Set(httpheaders.ContentType, "text/plain")
	rw.Header().Set(httpheaders.CacheControl, "no-cache")
	rw.WriteHeader(status)
	rw.Write(msg)

	return nil
}
