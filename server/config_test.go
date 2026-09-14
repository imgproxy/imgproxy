package server_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/server"
)

func TestLoadConfigFromEnvGracefulStopTimeout(t *testing.T) {
	tests := []struct {
		name                string
		timeout             string
		gracefulStopTimeout string
		expected            time.Duration
	}{
		{
			name:     "Defaults",
			expected: 20 * time.Second,
		},
		{
			name:     "TimeoutChangesDefault",
			timeout:  "30",
			expected: 60 * time.Second,
		},
		{
			name:                "ExplicitGracefulStopTimeoutOverridesDefault",
			timeout:             "30",
			gracefulStopTimeout: "5",
			expected:            5 * time.Second,
		},
		{
			name:                "ExplicitGracefulStopTimeoutWithoutTimeout",
			gracefulStopTimeout: "5",
			expected:            5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout != "" {
				t.Setenv("IMGPROXY_TIMEOUT", tt.timeout)
			}
			if tt.gracefulStopTimeout != "" {
				t.Setenv("IMGPROXY_GRACEFUL_STOP_TIMEOUT", tt.gracefulStopTimeout)
			}

			c := server.NewDefaultConfig()
			_, err := server.LoadConfigFromEnv(&c)
			require.NoError(t, err)

			require.Equal(t, tt.expected, c.GracefulStopTimeout)
		})
	}
}
