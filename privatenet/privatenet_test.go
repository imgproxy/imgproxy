package privatenet_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/privatenet"
)

func TestIsPrivate(t *testing.T) {
	testCases := []struct {
		name       string
		ip         string
		expectPriv bool
	}{
		// RFC 1918 - private addresses
		{name: "RFC1918 10/8", ip: "10.0.0.1", expectPriv: true},
		{name: "RFC1918 172.16/12", ip: "172.16.0.1", expectPriv: true},
		{name: "RFC1918 192.168/16", ip: "192.168.1.1", expectPriv: true},

		// RFC 6598 - CGNAT shared address space (the missing range)
		{name: "CGNAT start", ip: "100.64.0.0", expectPriv: true},
		{name: "CGNAT middle", ip: "100.64.0.1", expectPriv: true},
		{name: "CGNAT end", ip: "100.127.255.255", expectPriv: true},
		{name: "Just before CGNAT", ip: "100.63.255.255", expectPriv: false},
		{name: "Just after CGNAT", ip: "100.128.0.0", expectPriv: false},

		// Public addresses
		{name: "Public 8.8.8.8", ip: "8.8.8.8", expectPriv: false},
		{name: "Public 1.1.1.1", ip: "1.1.1.1", expectPriv: false},

		// IPv6 RFC 4193 ULA (private)
		{name: "IPv6 ULA fc00", ip: "fc00::1", expectPriv: true},
		{name: "IPv6 ULA fd00", ip: "fd00::1", expectPriv: true},

		// IPv6 public
		{name: "IPv6 public 2001:4860::", ip: "2001:4860::1", expectPriv: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP: %s", tc.ip)
			result := privatenet.IsPrivate(ip)
			require.Equal(t, tc.expectPriv, result, "IsPrivate(%s) = %v, expected %v", tc.ip, result, tc.expectPriv)
		})
	}
}

func TestUnwrap(t *testing.T) {
	testCases := []struct {
		name            string
		ip              net.IP
		expectUnwrapped net.IP
	}{
		// IPv4 inputs — not an embedding scheme
		{
			name:            "IPv4 private",
			ip:              net.ParseIP("10.0.0.1"),
			expectUnwrapped: nil,
		},
		{
			name:            "IPv4 public",
			ip:              net.ParseIP("8.8.8.8"),
			expectUnwrapped: nil,
		},

		// RFC 4291 IPv4-compatible (::a.b.c.d / ::/96)
		{
			name:            "IPv4-compatible private 10.0.0.1",
			ip:              net.ParseIP("::10.0.0.1"),
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name:            "IPv4-compatible public 8.8.8.8",
			ip:              net.ParseIP("::8.8.8.8"),
			expectUnwrapped: net.ParseIP("8.8.8.8"),
		},
		{
			name:            "IPv4-compatible loopback 127.0.0.1",
			ip:              net.ParseIP("::127.0.0.1"),
			expectUnwrapped: net.ParseIP("127.0.0.1"),
		},

		// RFC 3056 6to4 (2002:a.b.c.d::/16)
		{
			name:            "6to4 private 10.0.0.1",
			ip:              net.ParseIP("2002:0a00:0001::"),
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name:            "6to4 public 8.8.8.8",
			ip:              net.ParseIP("2002:0808:0808::"),
			expectUnwrapped: net.ParseIP("8.8.8.8"),
		},
		{
			name:            "6to4 CGNAT 100.64.0.1",
			ip:              net.ParseIP("2002:6440:0001::"),
			expectUnwrapped: net.ParseIP("100.64.0.1"),
		},

		// RFC 6052 NAT64 well-known (64:ff9b::/96)
		{
			name:            "NAT64 well-known private 10.0.0.1",
			ip:              net.ParseIP("64:ff9b::10.0.0.1"),
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name:            "NAT64 well-known public 8.8.8.8",
			ip:              net.ParseIP("64:ff9b::8.8.8.8"),
			expectUnwrapped: net.ParseIP("8.8.8.8"),
		},
		{
			name:            "NAT64 well-known CGNAT 100.64.0.1",
			ip:              net.ParseIP("64:ff9b::100.64.0.1"),
			expectUnwrapped: net.ParseIP("100.64.0.1"),
		},

		// RFC 8215 NAT64 local-use (/48)
		// RFC 6052 §2.2 Figure 1 layout: /48 prefix (6 bytes), then two u bits (2 bytes),
		// then interleaved IPv4: bytes 6–7 (first two octets), byte 8 reserved, bytes 9–10 (last two octets),
		// then padding to 16 bytes total.
		{
			name: "NAT64 local-use embedded 10.0.0.1",
			ip: net.IP{
				0x64, 0xff, 0x9b, 0x01, 0x00, 0x00, // /48 prefix (bytes 0–5)
				0x0a, 0x00, // first 2 bytes of 10.0.0.1 (bytes 6–7)
				0x00,       // reserved (byte 8)
				0x00, 0x01, // last 2 bytes of 10.0.0.1 (bytes 9–10)
				0x00, 0x00, 0x00, 0x00, 0x00, // padding to 16 bytes
			},
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name: "NAT64 local-use embedded 192.168.1.1",
			ip: net.IP{
				0x64, 0xff, 0x9b, 0x01, 0x00, 0x00, // /48 prefix
				0xc0, 0xa8, // 192.168
				0x00,       // reserved
				0x01, 0x01, // 1.1
				0x00, 0x00, 0x00, 0x00, 0x00, // padding to 16 bytes
			},
			expectUnwrapped: net.ParseIP("192.168.1.1"),
		},

		// RFC 4380 Teredo (2001::/32)
		// IPv4 in last 32 bits, each byte XORed with 0xFF.
		{
			name: "Teredo embedded 10.0.0.1 (XORed: f5.ff.ff.fe)",
			ip: net.IP{
				0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xf5, 0xff, 0xff, 0xfe, // 10^0xff=f5, 0^0xff=ff, 0^0xff=ff, 1^0xff=fe
			},
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name: "Teredo embedded 8.8.8.8 (XORed: f7.f7.f7.f7)",
			ip: net.IP{
				0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xf7, 0xf7, 0xf7, 0xf7, // 8^0xff=f7, 8^0xff=f7, 8^0xff=f7, 8^0xff=f7
			},
			expectUnwrapped: net.ParseIP("8.8.8.8"),
		},

		// RFC 5214 ISATAP — interface ID 0000:5EFE:a.b.c.d or 0200:5EFE:a.b.c.d
		// Bytes 8–11 check for the pattern.
		{
			name: "ISATAP 0000:5EFE embedded 10.0.0.1",
			ip: net.IP{
				0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x5e, 0xfe, // 0000:5EFE
				0x0a, 0x00, 0x00, 0x01, // 10.0.0.1
			},
			expectUnwrapped: net.ParseIP("10.0.0.1"),
		},
		{
			name: "ISATAP 0200:5EFE embedded 192.168.1.1",
			ip: net.IP{
				0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x5e, 0xfe, // 0200:5EFE
				0xc0, 0xa8, 0x01, 0x01, // 192.168.1.1
			},
			expectUnwrapped: net.ParseIP("192.168.1.1"),
		},

		// Not an embedding scheme — should not unwrap
		{
			name:            "Regular IPv6 public address",
			ip:              net.ParseIP("2001:4860:4860::8888"),
			expectUnwrapped: nil,
		},
		{
			name:            "IPv6 ULA private",
			ip:              net.ParseIP("fc00::1"),
			expectUnwrapped: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped := privatenet.Unwrap(tc.ip)
			if tc.expectUnwrapped == nil {
				require.Nil(t, unwrapped, "Unwrap(%s) = %v, expected nil", tc.ip, unwrapped)
			} else {
				require.NotNil(t, unwrapped, "Unwrap(%s) returned nil address", tc.ip)
				require.Equal(t, tc.expectUnwrapped.String(), unwrapped.String(),
					"Unwrap(%s) = %s, expected %s", tc.ip, unwrapped, tc.expectUnwrapped)
			}
		})
	}
}

// TestRegressionCombined verifies that combined checks work:
// an embedded private address should be caught by the recursive IsPrivate check.
func TestRegressionCombined(t *testing.T) {
	testCases := []struct {
		name       string
		ip         string
		shouldFail bool
	}{
		// IPv6 wrappings of private/CGNAT addresses should fail IsPrivate
		{name: "IPv4-compatible 10.0.0.1", ip: "::10.0.0.1", shouldFail: true},
		{name: "IPv4-compatible 100.64.0.1 (CGNAT)", ip: "::100.64.0.1", shouldFail: true},
		{name: "6to4 192.168.0.1", ip: "2002:c0a8:0001::", shouldFail: true},
		{name: "NAT64 well-known 172.16.0.1", ip: "64:ff9b::172.16.0.1", shouldFail: true},
		{name: "NAT64 well-known 100.64.0.1 (CGNAT)", ip: "64:ff9b::100.64.0.1", shouldFail: true},

		// IPv6 wrappings of public addresses should pass
		{name: "IPv4-compatible 8.8.8.8", ip: "::8.8.8.8", shouldFail: false},
		{name: "6to4 1.1.1.1", ip: "2002:0101:0101::", shouldFail: false},
		{name: "NAT64 well-known 8.8.8.8", ip: "64:ff9b::8.8.8.8", shouldFail: false},

		// Plain private addresses should fail
		{name: "Direct 10.0.0.1", ip: "10.0.0.1", shouldFail: true},
		{name: "Direct 100.64.0.1 (CGNAT)", ip: "100.64.0.1", shouldFail: true},

		// Plain public addresses should pass
		{name: "Direct 8.8.8.8", ip: "8.8.8.8", shouldFail: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP: %s", tc.ip)
			failed := privatenet.IsPrivate(ip)
			require.Equal(t, tc.shouldFail, failed,
				"IsPrivate(%s) = %v, expected %v", tc.ip, failed, tc.shouldFail)
		})
	}
}

func TestIsLoopback(t *testing.T) {
	testCases := []struct {
		name           string
		ip             string
		expectLoopback bool
	}{
		// IPv4 loopback and unspecified
		{name: "IPv4 loopback 127.0.0.1", ip: "127.0.0.1", expectLoopback: true},
		{name: "IPv4 loopback 127.0.0.254", ip: "127.0.0.254", expectLoopback: true},
		{name: "IPv4 unspecified 0.0.0.0", ip: "0.0.0.0", expectLoopback: true},

		// IPv4 public
		{name: "IPv4 public 8.8.8.8", ip: "8.8.8.8", expectLoopback: false},
		{name: "IPv4 private 10.0.0.1", ip: "10.0.0.1", expectLoopback: false},

		// IPv6 loopback and unspecified
		{name: "IPv6 loopback ::1", ip: "::1", expectLoopback: true},
		{name: "IPv6 unspecified ::", ip: "::", expectLoopback: true},

		// IPv6 public
		{name: "IPv6 public 2001:4860::8888", ip: "2001:4860::8888", expectLoopback: false},

		// IPv6 embedding loopback
		{name: "IPv4-compatible loopback ::127.0.0.1", ip: "::127.0.0.1", expectLoopback: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP: %s", tc.ip)
			result := privatenet.IsLoopback(ip)
			require.Equal(t, tc.expectLoopback, result, "IsLoopback(%s) = %v, expected %v", tc.ip, result, tc.expectLoopback)
		})
	}
}

func TestIsLinkLocal(t *testing.T) {
	testCases := []struct {
		name            string
		ip              string
		expectLinkLocal bool
	}{
		// IPv4 link-local
		{name: "IPv4 link-local 169.254.0.1", ip: "169.254.0.1", expectLinkLocal: true},
		{name: "IPv4 link-local 169.254.169.254", ip: "169.254.169.254", expectLinkLocal: true},

		// IPv4 public
		{name: "IPv4 public 8.8.8.8", ip: "8.8.8.8", expectLinkLocal: false},
		{name: "IPv4 private 10.0.0.1", ip: "10.0.0.1", expectLinkLocal: false},

		// IPv6 link-local
		{name: "IPv6 link-local fe80::1", ip: "fe80::1", expectLinkLocal: true},

		// IPv6 public
		{name: "IPv6 public 2001:4860::8888", ip: "2001:4860::8888", expectLinkLocal: false},

		// IPv6 embedding link-local
		{name: "IPv4-compatible link-local ::169.254.0.1", ip: "::169.254.0.1", expectLinkLocal: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			require.NotNil(t, ip, "failed to parse IP: %s", tc.ip)
			result := privatenet.IsLinkLocal(ip)
			require.Equal(t, tc.expectLinkLocal, result, "IsLinkLocal(%s) = %v, expected %v", tc.ip, result, tc.expectLinkLocal)
		})
	}
}
