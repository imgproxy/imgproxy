package httpheaders

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	from := http.Header{
		"X-Test-1": {"value1", "value2"},
		"X-Test-2": {"value3"},
		"X-Test-3": {"value4"},
		"X-Test-4": nil,
	}

	to := http.Header{
		"X-Test-1": {"oldvalue"},
		"X-Test-4": {"value5"},
		"X-Test-5": {"value6"},
	}

	Copy(from, to, []string{"X-Test-1", "x-test-3", "X-Non-Existent"})

	require.Equal(t, []string{"value1", "value2"}, to.Values("X-Test-1"))
	require.Equal(t, []string{"value4"}, to.Values("X-Test-3"))
	require.Equal(t, []string{"value5"}, to.Values("X-Test-4"))
	require.Equal(t, []string{"value6"}, to.Values("X-Test-5"))
	require.Empty(t, to.Values("X-Test-2"))
}

func TestCopyAll(t *testing.T) {
	from := http.Header{
		"X-Test-1": {"value1", "value2"},
		"X-Test-2": {"value3"},
		"X-Test-3": nil,
	}

	to := http.Header{
		"X-Test-1": {"oldvalue"},
		"X-Test-3": {"value4"},
		"X-Test-4": {"value5"},
	}

	testCases := []struct {
		overwrite bool
		expected  http.Header
	}{
		{
			overwrite: false,
			expected: http.Header{
				"X-Test-1": {"oldvalue"},
				"X-Test-2": {"value3"},
				"X-Test-3": {"value4"},
				"X-Test-4": {"value5"},
			},
		},
		{
			overwrite: true,
			expected: http.Header{
				"X-Test-1": {"value1", "value2"},
				"X-Test-2": {"value3"},
				"X-Test-3": {"value4"},
				"X-Test-4": {"value5"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("overwrite=%v", tc.overwrite), func(t *testing.T) {
			toCopy := to.Clone() // Clone to avoid modifying the original 'to' header
			CopyAll(from, toCopy, tc.overwrite)
			require.Equal(t, tc.expected, toCopy)
		})
	}
}

func TestCopyFromRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	req.Host = "example.com"
	req.Header = http.Header{
		"X-Test-1": {"value1", "value2"},
		"X-Test-2": {"value3"},
		"X-Test-3": nil,
	}

	header := http.Header{
		"X-Test-1": {"oldvalue"},
		"X-Test-3": {"value4"},
		"X-Test-4": {"value5"},
	}

	CopyFromRequest(req, header, []string{"X-Test-1", "x-test-2", "host", "X-Non-Existent"})

	require.Equal(t, []string{"value1", "value2"}, header.Values("X-Test-1"))
	require.Equal(t, []string{"value3"}, header.Values("X-Test-2"))
	require.Equal(t, []string{"value4"}, header.Values("X-Test-3"))
	require.Equal(t, []string{"value5"}, header.Values("X-Test-4"))
	require.Equal(t, []string{"example.com"}, header.Values("Host"))
}

func TestCopyToRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	req.Header = http.Header{
		"X-Test-1": {"oldvalue"},
		"X-Test-3": {"value4"},
		"X-Test-4": {"value5"},
	}

	header := http.Header{
		"X-Test-1": {"value1", "value2"},
		"X-Test-2": {"value3"},
		"X-Test-3": nil,
		"Host":     {"newhost.com"},
	}

	CopyToRequest(header, req)

	require.Equal(t, []string{"value1", "value2"}, req.Header.Values("X-Test-1"))
	require.Equal(t, []string{"value3"}, req.Header.Values("X-Test-2"))
	require.Equal(t, []string{"value4"}, req.Header.Values("X-Test-3"))
	require.Equal(t, []string{"value5"}, req.Header.Values("X-Test-4"))
	require.Equal(t, "newhost.com", req.Host)
}
