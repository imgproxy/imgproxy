package security

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/stretchr/testify/require"
)

func TestVerifySourceNetwork(t *testing.T) {
	testCases := []struct {
		name           string
		addr           string
		allowLoopback  bool
		allowLinkLocal bool
		allowPrivate   bool
		expectedErr    error
	}{
		{
			name:           "Invalid IP address",
			addr:           "not-an-ip",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    ErrInvalidSourceAddress,
		},
		{
			name:           "Loopback local not allowed",
			addr:           "127.0.0.1",
			allowLoopback:  false,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    ErrSourceAddressNotAllowed,
		},
		{
			name:           "Loopback local allowed",
			addr:           "127.0.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    nil,
		},
		{
			name:           "Unspecified (0.0.0.0) not allowed",
			addr:           "0.0.0.0",
			allowLoopback:  false,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    ErrSourceAddressNotAllowed,
		},
		{
			name:           "Link local unicast not allowed",
			addr:           "169.254.0.1",
			allowLoopback:  true,
			allowLinkLocal: false,
			allowPrivate:   true,
			expectedErr:    ErrSourceAddressNotAllowed,
		},
		{
			name:           "Link local unicast allowed",
			addr:           "169.254.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    nil,
		},
		{
			name:           "Private address not allowed",
			addr:           "192.168.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   false,
			expectedErr:    ErrSourceAddressNotAllowed,
		},
		{
			name:           "Private address allowed",
			addr:           "192.168.0.1",
			allowLoopback:  true,
			allowLinkLocal: true,
			allowPrivate:   true,
			expectedErr:    nil,
		},
		{
			name:           "Global unicast should be allowed",
			addr:           "8.8.8.8",
			allowLoopback:  false,
			allowLinkLocal: false,
			allowPrivate:   false,
			expectedErr:    nil,
		},
		{
			name:           "Port in address with global IP",
			addr:           "8.8.8.8:8080",
			allowLoopback:  false,
			allowLinkLocal: false,
			allowPrivate:   false,
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Backup original config
			originalLoopback := config.AllowLoopbackSourceAddresses
			originalLinkLocal := config.AllowLinkLocalSourceAddresses
			originalPrivate := config.AllowPrivateSourceAddresses

			// Restore original config after test
			defer func() {
				config.AllowLoopbackSourceAddresses = originalLoopback
				config.AllowLinkLocalSourceAddresses = originalLinkLocal
				config.AllowPrivateSourceAddresses = originalPrivate
			}()

			// Override config for the test
			config.AllowLoopbackSourceAddresses = tc.allowLoopback
			config.AllowLinkLocalSourceAddresses = tc.allowLinkLocal
			config.AllowPrivateSourceAddresses = tc.allowPrivate

			err := VerifySourceNetwork(tc.addr)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
