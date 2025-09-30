package monitoring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetaFilter(t *testing.T) {
	// Create a Meta with some test data
	meta := Meta{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": 42,
	}

	// Test filtering with existing keys
	filtered := meta.Filter("key1", "key3")

	// Check that filtered meta has the correct keys
	require.Len(t, filtered, 2)
	require.Equal(t, "value1", filtered["key1"])
	require.Equal(t, "value3", filtered["key3"])

	// Check that non-requested keys are not present
	require.NotContains(t, filtered, "key2")
	require.NotContains(t, filtered, "key4")

	// Test filtering with non-existing keys
	filtered2 := meta.Filter("nonexistent")
	require.Empty(t, filtered2)

	// Test filtering with empty parameters
	filtered3 := meta.Filter()
	require.Empty(t, filtered3)
}
