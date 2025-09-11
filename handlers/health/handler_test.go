package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/server/responsewriter"
)

func TestHealthHandler(t *testing.T) {
	// Create responsewriter.Factory
	rwConf := responsewriter.NewDefaultConfig()
	rwf, err := responsewriter.NewFactory(&rwConf)
	require.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a new health handler
	h := New()

	// Call the handler function directly (no need for actual HTTP request)
	h.Execute("test-req-id", rwf.NewWriter(rr), nil)

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
