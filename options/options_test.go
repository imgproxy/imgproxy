package options

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptionsTestSuite struct {
	suite.Suite
}

func (s *OptionsTestSuite) TestGet() {
	o := New()
	o.Set("string_key", "string_value")
	o.Set("bool_key", true)

	// Existing keys
	s.Require().Equal("string_value", Get(o, "string_key", "default_value"))
	s.Require().True(Get(o, "bool_key", false))

	// Non-existing keys
	s.Require().Equal("default_value", Get(o, "non_existing_key", "default_value"))
	s.Require().False(Get(o, "another_non_existing_key", false))

	// Type mismatch
	s.Require().Panics(func() {
		_ = Get(o, "string_key", 42)
	})
	s.Require().Panics(func() {
		_ = Get(o, "bool_key", "not_a_bool")
	})
}

func (s *OptionsTestSuite) TestAppendToSlice() {
	o := New()
	o.Set("slice", []int{1, 2, 3})

	// Append to existing slice
	AppendToSlice(o, "slice", 4, 5)
	s.Require().Equal([]int{1, 2, 3, 4, 5}, Get(o, "slice", []int{}))

	// Append to non-existing slice
	AppendToSlice(o, "new_slice", 10, 20)
	s.Require().Equal([]int{10, 20}, Get(o, "new_slice", []int{}))

	// Type mismatch
	s.Require().Panics(func() {
		AppendToSlice(o, "slice", "not_an_int")
	})
}

func (s *OptionsTestSuite) TestSliceContains() {
	o := New()
	o.Set("slice", []string{"apple", "banana", "cherry"})

	// Existing values
	s.Require().True(SliceContains(o, "slice", "banana"))
	s.Require().False(SliceContains(o, "slice", "date"))

	// Non-existing slice
	s.Require().False(SliceContains(o, "non_existing_slice", "anything"))

	// Type mismatch
	s.Require().Panics(func() {
		SliceContains(o, "slice", 42)
	})
}

func (s *OptionsTestSuite) TestPropagate() {
	o := New()
	o.Set("key1", "value1")
	o.Set("key2", 100)
	o.Set("key3", false)

	child := o.AddChild()
	child.Set("key1", "child_value1")
	child.Set("key3", true)

	grandChild := child.AddChild()
	grandChild.Set("key2", 300)

	o.Propagate("key1")
	o.Propagate("key2")

	s.Require().Equal("value1", Get(child, "key1", ""))
	s.Require().Equal(100, Get(child, "key2", 0))
	s.Require().True(Get(child, "key3", false))

	s.Require().Equal("value1", Get(grandChild, "key1", ""))
	s.Require().Equal(100, Get(grandChild, "key2", 0))
	s.Require().False(grandChild.Has("key3"))
}

func (s *OptionsTestSuite) TestDeleteFromDescendants() {
	o := New()
	o.Set("key1", "value1")

	child := o.AddChild()
	child.Set("key1", "child_value1")
	child.Set("key2", 200)

	grandChild := child.AddChild()
	grandChild.Set("key1", "grandchild_value1")
	grandChild.Set("key2", 300)

	o.DeleteFromDescendants("key1")

	s.Require().Equal("value1", Get(o, "key1", ""))

	s.Require().False(child.Has("key1"))
	s.Require().Equal(200, Get(child, "key2", 0))

	s.Require().False(grandChild.Has("key1"))
	s.Require().Equal(300, Get(grandChild, "key2", 0))
}

func (s *OptionsTestSuite) TestCopyValue() {
	o := New()
	o.Set("key1", 100)
	o.Set("key2", 200)
	o.Set("key3", 200)

	// Existing to existing
	o.CopyValue("key1", "key2")
	s.Require().Equal(100, Get(o, "key2", 0))

	// Existing to new
	o.CopyValue("key1", "key4")
	s.Require().Equal(100, Get(o, "key4", 0))

	// Non-existing to new
	o.CopyValue("non_existing_key", "key5")
	s.Require().False(o.Has("key5"))

	// Non-existing to existing
	o.CopyValue("another_non_existing_key", "key3")
	s.Require().Equal(200, Get(o, "key3", 0))
}

func (s *OptionsTestSuite) TestGetInt() {
	o := New()
	o.Set("int", 42)
	o.Set("int32", int32(32))
	o.Set("int16", int16(16))
	o.Set("int8", int8(8))
	o.Set("float", 3.14)
	o.Set("string", "not_an_int")

	// Integer types
	s.Require().Equal(42, o.GetInt("int", 0))
	s.Require().Equal(32, o.GetInt("int32", 0))
	s.Require().Equal(16, o.GetInt("int16", 0))
	s.Require().Equal(8, o.GetInt("int8", 0))

	// Non-existing key
	s.Require().Equal(100, o.GetInt("non_existing_key", 100))

	// Type mismatch
	s.Require().Panics(func() {
		o.GetInt("float", 0)
	})
	s.Require().Panics(func() {
		o.GetInt("string", 0)
	})
}

func (s *OptionsTestSuite) TestGetFloat() {
	o := New()
	o.Set("float64", 3.14)
	o.Set("float32", float32(2.71))
	o.Set("int", 42)
	o.Set("int16", int16(16))
	o.Set("string", "not_a_float")

	// Float types
	s.Require().InDelta(3.14, o.GetFloat("float64", 0.0), 0.000001)
	s.Require().InDelta(2.71, o.GetFloat("float32", 0.0), 0.000001)

	// Integer types
	s.Require().InDelta(42.0, o.GetFloat("int", 0.0), 0.000001)
	s.Require().InDelta(16.0, o.GetFloat("int16", 0.0), 0.000001)

	// Non-existing key
	s.Require().InDelta(1.618, o.GetFloat("non_existing_key", 1.618), 0.000001)

	// Type mismatch
	s.Require().Panics(func() {
		o.GetFloat("string", 0.0)
	})
}

func testOptions() *Options {
	o := New()
	o.Set("string_key", "string_value")
	o.Set("int_key", 42)
	o.Set("float_key", 3.14)
	o.Set("bool_key", true)
	o.Set("group1.key1", "value1")
	o.Set("group1.key2", 100)
	o.Set("group2.key1", false)
	o.Set("group2.key2", 2.71)
	o.Set("group2.subgroup.key", "subvalue")
	o.Set("group2.subgroup.num", 256)
	return o
}

func testNestedOptions() *Options {
	o := testOptions()

	child := o.AddChild()
	child.Set("string_key", "child_string_value")
	child.Set("int_key", 84)
	child.Set("child_only_key", "only_in_child")

	grandChild := child.AddChild()
	grandChild.Set("string_key", "grandchild_string_value")
	grandChild.Set("int_key", 168)
	grandChild.Set("grandchild_only_key", "only_in_grandchild")

	return o
}

func (s *OptionsTestSuite) TestDepth() {
	o := testNestedOptions()

	s.Require().Equal(0, o.Depth())
	s.Require().Equal(1, o.Child().Depth())
	s.Require().Equal(2, o.Child().Child().Depth())
}

func (s *OptionsTestSuite) TestMap() {
	s.Run("WithoutChildren", func() {
		o := testOptions()

		expected := map[string]any{
			"string_key":          "string_value",
			"int_key":             42,
			"float_key":           3.14,
			"bool_key":            true,
			"group1.key1":         "value1",
			"group1.key2":         100,
			"group2.key1":         false,
			"group2.key2":         2.71,
			"group2.subgroup.key": "subvalue",
			"group2.subgroup.num": 256,
		}

		s.Require().Equal(expected, o.Map())
	})

	s.Run("WithChildren", func() {
		o := testNestedOptions()

		expected := map[string]any{
			"0.string_key":          "string_value",
			"0.int_key":             42,
			"0.float_key":           3.14,
			"0.bool_key":            true,
			"0.group1.key1":         "value1",
			"0.group1.key2":         100,
			"0.group2.key1":         false,
			"0.group2.key2":         2.71,
			"0.group2.subgroup.key": "subvalue",
			"0.group2.subgroup.num": 256,
			"1.string_key":          "child_string_value",
			"1.int_key":             84,
			"1.child_only_key":      "only_in_child",
			"2.string_key":          "grandchild_string_value",
			"2.int_key":             168,
			"2.grandchild_only_key": "only_in_grandchild",
		}

		s.Require().Equal(expected, o.Map())
	})
}

func (s *OptionsTestSuite) TestNestedMap() {
	s.Run("WithoutChildren", func() {
		o := testOptions()

		expected := map[string]any{
			"string_key": "string_value",
			"int_key":    42,
			"float_key":  3.14,
			"bool_key":   true,
			"group1": map[string]any{
				"key1": "value1",
				"key2": 100,
			},
			"group2": map[string]any{
				"key1": false,
				"key2": 2.71,
				"subgroup": map[string]any{
					"key": "subvalue",
					"num": 256,
				},
			},
		}

		s.Require().Equal(expected, o.NestedMap())
	})

	s.Run("WithChildren", func() {
		o := testNestedOptions()

		expected := map[string]any{
			"0": map[string]any{
				"string_key": "string_value",
				"int_key":    42,
				"float_key":  3.14,
				"bool_key":   true,
				"group1": map[string]any{
					"key1": "value1",
					"key2": 100,
				},
				"group2": map[string]any{
					"key1": false,
					"key2": 2.71,
					"subgroup": map[string]any{
						"key": "subvalue",
						"num": 256,
					},
				},
			},
			"1": map[string]any{
				"string_key":     "child_string_value",
				"int_key":        84,
				"child_only_key": "only_in_child",
			},
			"2": map[string]any{
				"string_key":          "grandchild_string_value",
				"int_key":             168,
				"grandchild_only_key": "only_in_grandchild",
			},
		}

		s.Require().Equal(expected, o.NestedMap())
	})
}

func TestOptions(t *testing.T) {
	suite.Run(t, new(OptionsTestSuite))
}

func BenchmarkLogValue(b *testing.B) {
	o := testNestedOptions()

	b.ResetTimer()

	for range b.N {
		_ = o.LogValue()
	}
}

func BenchmarkNestedMap(b *testing.B) {
	o := testNestedOptions()

	b.ResetTimer()

	for range b.N {
		_ = o.NestedMap()
	}
}

func BenchmarkMap(b *testing.B) {
	o := testNestedOptions()

	b.ResetTimer()

	for range b.N {
		_ = o.Map()
	}
}
