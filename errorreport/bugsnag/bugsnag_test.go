package bugsnag_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/bugsnag"
	"github.com/stretchr/testify/require"
)

func TestReportWithNilRequest(t *testing.T) {
	r, err := bugsnag.New(&bugsnag.Config{Key: "test", Stage: "test"})
	require.NoError(t, err)
	require.NotNil(t, r)

	testErr := errctx.NewTextError("boom", 0)

	require.NotPanics(t, func() {
		r.Report(testErr, nil, map[string]any{"Request ID": "abc123"})
	})
}
