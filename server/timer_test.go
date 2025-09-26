package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckTimeout(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		fail  bool
	}{
		{
			name:  "WithoutTimeout",
			setup: context.Background,
			fail:  false,
		},
		{
			name: "ActiveTimerContext",
			setup: func() context.Context {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				newReq, _ := startRequestTimer(req, 10*time.Second)
				return newReq.Context()
			},
			fail: false,
		},
		{
			name: "CancelledContext",
			setup: func() context.Context {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				newReq, cancel := startRequestTimer(req, 10*time.Second)
				cancel() // Cancel immediately
				return newReq.Context()
			},
			fail: true,
		},
		{
			name: "DeadlineExceeded",
			setup: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				time.Sleep(time.Millisecond * 10) // Ensure timeout
				return ctx
			},
			fail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			err := CheckTimeout(ctx)

			if tt.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
