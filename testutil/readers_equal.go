package testutil

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const bufSize = 4096

// RequireReadersEqual compares two io.Reader contents in a streaming manner.
// It fails the test if contents differ or if reading fails.
func ReadersEqual(t *testing.T, expected, actual io.Reader) bool {
	// Marks this function as a test helper so in case failure happens here, location would
	// point to the correct line in the calling test.
	t.Helper()

	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)

	for {
		n1, err1 := expected.Read(buf1)
		n2, err2 := actual.Read(buf2)

		if n1 != n2 {
			return false
		}

		require.Equal(t, buf1[:n1], buf2[:n1])

		if err1 == io.EOF && err2 == io.EOF {
			return true
		}

		require.NoError(t, err1)
		require.NoError(t, err2)
	}
}
