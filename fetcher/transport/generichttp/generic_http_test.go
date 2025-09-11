package generichttp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifySourceNetwork(t *testing.T) {
	testCases := []struct {
		name           string
		addr           string
		allowLoopback  bool
		allowLinkLocal bool
		allowPrivate   bool
		expectErr      bool
	}{
		{
			name:           "Invalid IP address",
			addr:           "not-an-ip",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      true,
		},
		{
			name:           "Loopback local not allowed",
			addr:           "127.0.0.1",
			allowLoopback:  false,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      true,
		},
		{
			name:           "Loopback local allowed",
			addr:           "127.0.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      false,
		},
		{
			name:           "Unspecified (0.0.0.0) not allowed",
			addr:           "0.0.0.0",
			allowLoopback:  false,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      true,
		},
		{
			name:           "Link local unicast not allowed",
			addr:           "169.254.0.1",
			allowLoopback:  true,
			allowLinkLocal: false,
			allowPrivate:   true,
			expectErr:      true,
		},
		{
			name:           "Link local unicast allowed",
			addr:           "169.254.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      false,
		},
		{
			name:           "Private address not allowed",
			addr:           "192.168.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   false,
			expectErr:      true,
		},
		{
			name:           "Private address allowed",
			addr:           "192.168.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectErr:      false,
		},
		{
			name:           "Global unicast should be allowed",
			addr:           "8.8.8.8",
			allowLoopback:  false,
			allowLinkLocal: false,
			allowPrivate:   false,
			expectErr:      false,
		},
		{
			name:           "Port in address with global IP",
			addr:           "8.8.8.8:8080",
			allowLoopback:  false,
			allowLinkLocal: false,
			allowPrivate:   false,
			expectErr:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewDefaultConfig()

			// Override config for the test
			config.AllowLoopbackSourceAddresses = tc.allowLoopback
			config.AllowLinkLocalSourceAddresses = tc.allowLinkLocal
			config.AllowPrivateSourceAddresses = tc.allowPrivate

			err := verifySourceNetwork(tc.addr, &config)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
