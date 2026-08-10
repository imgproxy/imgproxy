package clientfeatures_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/clientfeatures"
)

func TestEnableClientHintsParsedTrue(t *testing.T) {
	t.Setenv("IMGPROXY_ENABLE_CLIENT_HINTS", "true")

	cfg, err := clientfeatures.LoadConfigFromEnv(nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.True(t, cfg.EnableClientHints)
}

func TestEnableClientHintsParsedFalseByDefault(t *testing.T) {
	cfg, err := clientfeatures.LoadConfigFromEnv(nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.False(t, cfg.EnableClientHints)
}

func TestEnableClientHintsParsedInvalidValue(t *testing.T) {
	t.Setenv("IMGPROXY_ENABLE_CLIENT_HINTS", "not-a-bool")

	_, err := clientfeatures.LoadConfigFromEnv(nil)

	require.Error(t, err)
}
