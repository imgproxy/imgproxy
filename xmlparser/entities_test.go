package xmlparser_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/xmlparser"
)

func TestParseEntityMap(t *testing.T) {
	directive := &xmlparser.Directive{
		Data: []byte(`<!DOCTYPE svg [
		<!ENTITY entity1 "Value1">
		<!ENTITY entity2 'Value2'>
		<!ENTITY entity3 "Value with spaces">
		<!--
			<!ENTITY fake "ShouldNotBeParsed">
		-->
	]>`),
	}

	nodes := []xmlparser.Token{directive}

	em := xmlparser.ParseEntityMap(nodes)

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
		result := xmlparser.ReplaceEntitiesString(input1, em)
		require.Equal(t, expected1, result)

		result = xmlparser.ReplaceEntitiesString(input2, em)
		require.Equal(t, expected2, result)
	})

	t.Run("bytes", func(t *testing.T) {
		result := xmlparser.ReplaceEntitiesBytes([]byte(input1), em)
		require.Equal(t, expected1, string(result))

		result = xmlparser.ReplaceEntitiesBytes([]byte(input2), em)
		require.Equal(t, expected2, string(result))
	})
}

func BenchmarkParseEntityMap(b *testing.B) {
	directive := &xmlparser.Directive{
		Data: []byte(`<!DOCTYPE svg [
		<!ENTITY entity1 "Value1">
		<!ENTITY entity2 'Value2'>
		<!ENTITY entity3 "Value with spaces">
	]>`),
	}

	nodes := []xmlparser.Token{directive}

	b.ResetTimer()

	for b.Loop() {
		_ = xmlparser.ParseEntityMap(nodes)
	}
}

func BenchmarkReplaceEntitiesString(b *testing.B) {
	em := map[string][]byte{
		"entity1": []byte("Entity Value1"),
		"entity2": []byte("Entity Value2"),
		"entity3": []byte("Entity Value3"),
	}

	data := "This is a test string with &entity1;, &entity2;, and &entity3; to be replaced."

	b.ResetTimer()

	for b.Loop() {
		_ = xmlparser.ReplaceEntitiesString(data, em)
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

	for b.Loop() {
		_ = xmlparser.ReplaceEntitiesBytes(data, em)
	}
}
