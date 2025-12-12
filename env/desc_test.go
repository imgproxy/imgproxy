package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariableReadCorrectly(t *testing.T) {
	t.Setenv("TEST_INT", "42")

	desc := Int("TEST_INT")
	var result int
	err := desc.Parse(&result)

	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestMissingVariable(t *testing.T) {
	desc := Int("TEST_INT_MISSING")
	result := 123 // existing value
	err := desc.Parse(&result)

	require.NoError(t, err)
	assert.Equal(t, 123, result) // value should not change
}

func TestEmptyValue(t *testing.T) {
	t.Setenv("TEST_INT", "")

	desc := Int("TEST_INT")
	result := 123 // existing value
	err := desc.Parse(&result)

	require.NoError(t, err)
	assert.Equal(t, 123, result) // value should not change
}

func TestParseFailure(t *testing.T) {
	t.Setenv("TEST_INT", "not_a_number")

	desc := Int("TEST_INT")
	var result int
	err := desc.Parse(&result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_INT")
	assert.Contains(t, err.Error(), docsUrl+"TEST_INT")
}

func TestCustomFormatAndDocsURL(t *testing.T) {
	desc := Int("TEST_INT").
		WithFormat("custom integer format").
		WithDocsURL("https://custom.docs.url")

	assert.Contains(t, desc.ErrorEmpty().Error(), "custom.docs.url")
}
