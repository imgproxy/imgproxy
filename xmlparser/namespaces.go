package xmlparser

import (
	"maps"
	"strings"
)

const (
	// NamespaceXML is the namespace bound to the predefined `xml` prefix.
	NamespaceXML = "http://www.w3.org/XML/1998/namespace"

	// namespaceXMLNS is the namespace of xmlns declarations.
	namespaceXMLNS = "http://www.w3.org/2000/xmlns/"
)

// nsBindings holds the xmlns bindings in effect at a point in the tree: the
// default namespace (xmlns="...") and the prefixes (xmlns:p="...") declared
// on the current element or any of its ancestors.
//
// An element or attribute name carries only a prefix (or nothing), while the
// URI it stands for is declared on the element itself or on an ancestor. The
// parser passes nsBindings down the tree, so each node can be resolved to its
// namespace URI without looking at its parents.
//
// It is immutable: with returns new bindings and never modifies the receiver.
type nsBindings struct {
	// defaultNS is the namespace of unprefixed elements ("" if none).
	// Unprefixed attributes never use it.
	defaultNS string
	// prefixes maps prefixes to namespace URIs.
	prefixes map[string]string
}

// with returns the bindings for an element with the given attributes:
// the receiver plus the xmlns declarations found in attrs.
func (s nsBindings) with(attrs *Attributes) nsBindings {
	next := s
	copied := false

	for attr := range attrs.Iter() {
		prefix, local := attr.Name.Split()

		switch {
		case prefix == "" && local == "xmlns":
			next.defaultNS = strings.TrimSpace(attr.Value)
		case prefix == "xmlns":
			uri := strings.TrimSpace(attr.Value)

			// Generators often redeclare the same prefix on nested elements;
			// that changes nothing, so don't copy the map for it.
			if next.prefixes[local] == uri {
				continue
			}

			if !copied {
				// maps.Clone returns nil for a nil map
				next.prefixes = maps.Clone(s.prefixes)
				if next.prefixes == nil {
					next.prefixes = make(map[string]string)
				}

				copied = true
			}

			next.prefixes[local] = uri
		}
	}

	return next
}

// elementNS returns the namespace of an element ("" if none or unresolved).
func (s nsBindings) elementNS(name Name) string {
	prefix := name.Space()

	if prefix == "" {
		return s.defaultNS
	}

	return s.lookup(prefix)
}

// attrNS returns the namespace of an attribute.
// Unprefixed attributes have no namespace.
func (s nsBindings) attrNS(name Name) string {
	prefix := name.Space()

	if prefix == "" {
		if name == "xmlns" {
			return namespaceXMLNS
		}

		return ""
	}

	if prefix == "xmlns" {
		return namespaceXMLNS
	}

	return s.lookup(prefix)
}

// lookup returns the namespace bound to the prefix ("" if unresolved).
func (s nsBindings) lookup(prefix string) string {
	if prefix == "xml" {
		return NamespaceXML
	}

	return s.prefixes[prefix]
}
