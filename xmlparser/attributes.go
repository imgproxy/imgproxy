package xmlparser

import (
	"iter"
	"slices"
)

// Attribute represents a single XML attribute.
type Attribute struct {
	Name  Name
	Value string
}

// Attributes represents a collection of XML attributes.
type Attributes struct {
	s []*Attribute        // Slice for ordered iteration
	m map[Name]*Attribute // Map for quick lookup
}

// NewAttributes creates a new [Attributes] instance from the given attributes.
func NewAttributes(attrs ...*Attribute) *Attributes {
	// If no attributes are provided, return an empty Attributes instance
	if len(attrs) == 0 {
		return &Attributes{}
	}

	// Create a map for quick lookup
	m := make(map[Name]*Attribute, len(attrs))
	for _, attr := range attrs {
		m[attr.Name] = attr
	}

	return &Attributes{
		s: attrs,
		m: m,
	}
}

// Get retrieves the value of the attribute with the given name.
// It returns the value and true if the attribute exists, or an empty string and false otherwise.
func (a *Attributes) Get(name Name) (string, bool) {
	if attr, ok := a.m[name]; ok {
		return attr.Value, true
	}
	return "", false
}

// GetByString is like Get but accepts a string name.
func (a *Attributes) GetByString(name string) (string, bool) {
	return a.Get(Name(name))
}

// Set sets the value of the attribute with the given name.
// If the attribute does not exist, it is created.
func (a *Attributes) Set(name Name, value string) {
	// Initialize the map if it is nil
	if a.m == nil {
		a.m = make(map[Name]*Attribute)
	}

	// If attribute exists, update its value
	if attr, exists := a.m[name]; exists {
		attr.Value = value
		return
	}

	// Otherwise, create a new attribute
	attr := &Attribute{Name: name, Value: value}
	a.m[name] = attr
	a.s = append(a.s, attr)
}

// SetByString is like Set but accepts a string name.
func (a *Attributes) SetByString(name, value string) {
	a.Set(Name(name), value)
}

// Delete removes the attribute with the given name.
func (a *Attributes) Delete(name Name) {
	// If attribute does not exist, nothing to do
	attr, exists := a.m[name]
	if !exists {
		return
	}

	// Remove from map
	delete(a.m, name)

	// Remove from slice
	if idx := slices.Index(a.s, attr); idx != -1 {
		a.s = slices.Delete(a.s, idx, idx+1)
	}
}

// DeleteByString is like Delete but accepts a string name.
func (a *Attributes) DeleteByString(name string) {
	a.Delete(Name(name))
}

// Has returns true if the attribute with the given name exists.
func (a *Attributes) Has(name Name) bool {
	_, exists := a.m[name]
	return exists
}

// HasByString is like Has but accepts a string name.
func (a *Attributes) HasByString(name string) bool {
	return a.Has(Name(name))
}

// Len returns the number of attributes.
func (a *Attributes) Len() int {
	return len(a.s)
}

// Iter returns an iterator over the attributes.
//
// !WARNING! Don't delete items from the Attributes while iterating.
// If you need to filter attributes, use the Filter method.
func (a *Attributes) Iter() iter.Seq[*Attribute] {
	return slices.Values(a.s)
}

// Filter removes all attributes that do not satisfy the predicate function.
func (a *Attributes) Filter(pred func(*Attribute) bool) {
	a.s = slices.DeleteFunc(a.s, func(attr *Attribute) bool {
		del := !pred(attr)
		if del {
			delete(a.m, attr.Name)
		}
		return del
	})
}

// Sort sorts the attributes according to the provided function.
func (a *Attributes) Sort(cmp func(i, j *Attribute) int) {
	slices.SortFunc(a.s, cmp)
}
