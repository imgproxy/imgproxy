package semaphores

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSemaphoresQueueDisabled(t *testing.T) {
	s, err := New(&Config{RequestsQueueSize: 0, Workers: 1})
	require.NoError(t, err)

	// Queue acquire should always work when disabled
	release, err := s.AcquireQueue()
	require.NoError(t, err)
	release() // Should not panic

	procRelease, err := s.AcquireProcessing(t.Context())
	require.NoError(t, err)
	procRelease()
}

func TestSemaphoresQueueEnabled(t *testing.T) {
	s, err := New(&Config{RequestsQueueSize: 2, Workers: 1})
	require.NoError(t, err)

	// Should be able to acquire up to queue size
	release1, err := s.AcquireQueue()
	require.NoError(t, err)

	release2, err := s.AcquireQueue()
	require.NoError(t, err)

	// Third should fail (exceeds capacity)
	_, err = s.AcquireQueue()
	require.Error(t, err)

	// Release and try again
	release1()
	release3, err := s.AcquireQueue()
	require.NoError(t, err)

	release2()
	release3()
}

func TestSemaphoresInvalidConfig(t *testing.T) {
	_, err := New(&Config{RequestsQueueSize: 0, Workers: 0})
	require.Error(t, err)
}
