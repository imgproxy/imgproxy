package testutil

import (
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/transport"
	"github.com/stretchr/testify/require"
)

// NewDefaultFetcher creates a new fetcher with default config and transport
func NewDefaultFetcher(t require.TestingT) *fetcher.Fetcher {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	transportConfig := transport.NewDefaultConfig()
	transport, err := transport.New(transportConfig)
	require.NoError(t, err)

	config, err := fetcher.LoadFromEnv(fetcher.NewDefaultConfig())
	require.NoError(t, err)

	fetcher, err := fetcher.New(transport, config)
	require.NoError(t, err)

	return fetcher
}
