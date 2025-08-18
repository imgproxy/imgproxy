package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

func TestHealthHandler(t *testing.T) {
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function directly (no need for actual HTTP request)
	HealthHandler("test-req-id", rr, nil)

	// Check that we get a valid response (either 200 or 500 depending on vips state)
	assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusInternalServerError)

	// Check headers are set correctly
	assert.Equal(t, "text/plain", rr.Header().Get(httpheaders.ContentType))
	assert.Equal(t, "no-cache", rr.Header().Get(httpheaders.CacheControl))

	// Verify response format and content
	body := rr.Body.String()
	assert.NotEmpty(t, body)

	assert.Equal(t, "imgproxy is running", body)
}
