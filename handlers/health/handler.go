package health

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var imgproxyIsRunningMsg = []byte("imgproxy is running")

// Handler handles health requests
type Handler struct{}

// New creates new handler object
func New() *Handler {
	return &Handler{}
}

// Execute handles the health request
func (h *Handler) Execute(
	reqID string,
	rw server.ResponseWriter,
	req *http.Request,
) error {
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
		server.LogResponse(reqID, req, status, ierr)
	}

	rw.Header().Set(httpheaders.ContentType, "text/plain")
	rw.Header().Set(httpheaders.CacheControl, "no-cache")
	rw.WriteHeader(status)
	rw.Write(msg)

	return nil
}
