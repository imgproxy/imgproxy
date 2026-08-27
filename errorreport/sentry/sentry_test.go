package sentry_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport/sentry"
	"github.com/stretchr/testify/require"
)

func TestReportWithNilRequest(t *testing.T) {
	r, err := sentry.New(&sentry.Config{
		DSN:         "https://public@example.com/1",
		Release:     "test",
		Environment: "test",
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	testErr := errctx.NewTextError("boom", 0)

	require.NotPanics(t, func() {
		r.Report(testErr, nil, map[string]any{"Request ID": "abc123"})
	})
}
