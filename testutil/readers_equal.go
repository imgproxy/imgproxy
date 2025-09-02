package testutil

import (
	"io"

	"github.com/stretchr/testify/require"
)

const bufSize = 4096

// RequireReadersEqual compares two io.Reader contents in a streaming manner.
// It fails the test if contents differ or if reading fails.
func ReadersEqual(t require.TestingT, expected, actual io.Reader) bool {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

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
