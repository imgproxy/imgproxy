package testutil

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/imgproxy/imgproxy/v3/ioutil"
	"github.com/stretchr/testify/require"
)

const bufSize = 4096

// ReadersEqual compares two io.Reader contents in a streaming manner.
// It fails the test if contents differ or if reading fails.
func ReadersEqual(t *testing.T, expected, actual io.Reader) bool {
	// Marks this function as a test helper so in case failure happens here, location would
	// point to the correct line in the calling test.
	t.Helper()

	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)

	for {
		n1, err1 := ioutil.TryReadFull(expected, buf1)
		n2, err2 := ioutil.TryReadFull(actual, buf2)

		if n1 != n2 {
			return false
		}

		if !bytes.Equal(buf1[:n1], buf2[:n1]) {
			return false
		}

		if errors.Is(err1, io.EOF) && errors.Is(err2, io.EOF) {
			return true
		}

		require.NoError(t, err1)
		require.NoError(t, err2)
	}
}
