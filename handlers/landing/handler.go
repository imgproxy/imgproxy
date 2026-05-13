package landing

import (
	_ "embed"
	"net/http"

	"github.com/imgproxy/imgproxy/v4/httpheaders"
	"github.com/imgproxy/imgproxy/v4/server"
)

//go:embed body.html
var landingBody []byte

// Handler handles landing requests
type Handler struct{}

// New creates new handler object
func New() *Handler {
	return &Handler{}
}

// Execute handles the landing request
func (h *Handler) Execute(
	reqID string,
	rw server.ResponseWriter,
	req *http.Request,
) *server.Error {
	rw.Header().Set(httpheaders.ContentType, "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write(landingBody)
	return nil
}
