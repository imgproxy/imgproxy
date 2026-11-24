package otel_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/monitoring/otel"
)

func TestLoadConfigFromEnvMetricsInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval string
		expected time.Duration
		valid    bool
	}{
		{
			name:     "Default",
			expected: 10 * time.Second,
			valid:    true,
		},
		{
			name:     "Set",
			interval: "60000",
			expected: 60 * time.Second,
			valid:    true,
		},
		{
			name:     "Zero",
			interval: "0",
			expected: 0,
			valid:    false,
		},
		{
			name:     "Negative",
			interval: "-1000",
			expected: -1 * time.Second,
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("IMGPROXY_OPEN_TELEMETRY_ENABLE", "true")
			if tt.interval != "" {
				t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", tt.interval)
			}

			c := otel.NewDefaultConfig()
			_, err := otel.LoadConfigFromEnv(&c)
			require.NoError(t, err)

			require.Equal(t, tt.expected, c.MetricsInterval)

			if tt.valid {
				require.NoError(t, c.Validate())
			} else {
				require.Error(t, c.Validate())
			}
		})
	}
}

func TestLoadConfigFromEnvMetricsIntervalInvalid(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "abc")

	c := otel.NewDefaultConfig()
	_, err := otel.LoadConfigFromEnv(&c)
	require.Error(t, err)
}
