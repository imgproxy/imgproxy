package xmlparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEntityMap(t *testing.T) {
	directive := &Directive{
		Data: []byte(`<!DOCTYPE svg [
		<!ENTITY entity1 "Value1">
		<!ENTITY entity2 'Value2'>
		<!ENTITY entity3 "Value with spaces">
	]>`),
	}

	nodes := []any{directive}

	em := parseEntityMap(nodes)

	expected := map[string][]byte{
		"entity1": []byte("Value1"),
		"entity2": []byte("Value2"),
		"entity3": []byte("Value with spaces"),
	}

	require.Equal(t, expected, em)
}

func TestReplaceEntities(t *testing.T) {
	em := map[string][]byte{
		"entity1": []byte("Value1"),
		"entity2": []byte("Value2"),
		"entity3": []byte("Value3"),
	}

	input1 := `
		This is a test string with &entity1;, &entity2;, and &entity3; to be replaced.
		Unknown entity &entity4; should remain unchanged.
		Known entity &entity1; after unknown &entity4; should still replace.
		Incomplete entity &entity1 should also remain unchanged.
	`
	expected1 := `
		This is a test string with Value1, Value2, and Value3 to be replaced.
		Unknown entity &entity4; should remain unchanged.
		Known entity Value1 after unknown &entity4; should still replace.
		Incomplete entity &entity1 should also remain unchanged.
	`

	// Corner case with entity at the very start and end
	input2 := `&entity1;`
	expected2 := `Value1`

	t.Run("string", func(t *testing.T) {
		result := replaceEntitiesString(input1, em)
		require.Equal(t, expected1, result)

		result = replaceEntitiesString(input2, em)
		require.Equal(t, expected2, result)
	})

	t.Run("bytes", func(t *testing.T) {
		result := replaceEntitiesBytes([]byte(input1), em)
		require.Equal(t, expected1, string(result))

		result = replaceEntitiesBytes([]byte(input2), em)
		require.Equal(t, expected2, string(result))
	})
}

func BenchmarkReplaceEntitiesString(b *testing.B) {
	em := map[string][]byte{
		"entity1": []byte("Entity Value1"),
		"entity2": []byte("Entity Value2"),
		"entity3": []byte("Entity Value3"),
	}

	data := "This is a test string with &entity1;, &entity2;, and &entity3; to be replaced."

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = replaceEntitiesString(data, em)
	}
}

func BenchmarkReplaceEntitiesBytes(b *testing.B) {
	em := map[string][]byte{
		"entity1": []byte("Entity Value1"),
		"entity2": []byte("Entity Value2"),
		"entity3": []byte("Entity Value3"),
	}

	data := []byte("This is a test string with &entity1;, &entity2;, and &entity3; to be replaced.")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = replaceEntitiesBytes(data, em)
	}
}
