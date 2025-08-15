package handlers

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/server"
)

// HeadHandler is a simple handler that responds with a 200 OK status
func HeadHandler(reqID string, rw http.ResponseWriter, r *http.Request) error {
	server.LogResponse(reqID, r, 200, nil)
	rw.WriteHeader(200)
	return nil
}
