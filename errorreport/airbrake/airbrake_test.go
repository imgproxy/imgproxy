package airbrake_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/airbrake"
	"github.com/stretchr/testify/require"
)

func TestReportWithNilRequest(t *testing.T) {
	r, err := airbrake.New(&airbrake.Config{ProjectID: 1, ProjectKey: "test", Env: "test"})
	require.NoError(t, err)
	require.NotNil(t, r)

	testErr := errctx.NewTextError("boom", 0)

	require.NotPanics(t, func() {
		r.Report(testErr, nil, map[string]any{"Request ID": "abc123"})
	})
}
