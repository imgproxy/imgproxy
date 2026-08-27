package honeybadger_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/honeybadger"
	"github.com/stretchr/testify/require"
)

func TestReportWithNilRequest(t *testing.T) {
	r, err := honeybadger.New(&honeybadger.Config{Key: "test", Env: "test"})
	require.NoError(t, err)
	require.NotNil(t, r)

	testErr := errctx.NewTextError("boom", 0)

	require.NotPanics(t, func() {
		r.Report(testErr, nil, map[string]any{"Request ID": "abc123"})
	})
}
