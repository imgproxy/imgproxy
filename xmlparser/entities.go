package xmlparser

import (
	"bytes"
	"regexp"
	"strings"
)

var (
	entityDeclRe = regexp.MustCompile(`<!ENTITY\s+(\S+)\s+("[^"]+"|'[^']+')\s*>`)
	entityRe     = regexp.MustCompile(`&[a-zA-Z_:][a-zA-Z0-9_.:-]*;`)
)

// parseEntityMap searches for a DOCTYPE directive in the given nodes
// and extracts entity declarations into a map.
func parseEntityMap(nodes []any) map[string][]byte {
	// Find doctype
	var doctype *Directive
	for _, node := range nodes {
		if dir, ok := node.(*Directive); ok {
			doctype = dir
			break
		}
	}

	if doctype == nil {
		return nil
	}

	start := bytes.IndexByte(doctype.Data, '[')
	if start == -1 {
		return nil
	}

	matches := entityDeclRe.FindAllSubmatch(doctype.Data[start:], -1)
	if matches == nil {
		return nil
	}

	em := make(map[string][]byte, len(matches))
	for _, match := range matches {
		name := match[1]
		value := bytes.Trim(match[2], `"'`)
		em[string(name)] = value
	}

	return em
}

// replaceEntitiesBytes replaces XML entities in the given byte slice
// according to the provided entity map.
func replaceEntitiesBytes(data []byte, em map[string][]byte) []byte {
	// Quick check to avoid unnecessary processing
	if bytes.IndexByte(data, '&') == -1 {
		return data
	}

	return entityRe.ReplaceAllFunc(data, func(entity []byte) []byte {
		name := entity[1 : len(entity)-1]
		if val, ok := em[string(name)]; ok {
			return val
		}
		// Not found, return original
		return entity
	})
}

// replaceEntitiesString replaces XML entities in the given string
// according to the provided entity map.
func replaceEntitiesString(data string, em map[string][]byte) string {
	// Quick check to avoid unnecessary processing
	if strings.IndexByte(data, '&') == -1 {
		return data
	}

	return entityRe.ReplaceAllStringFunc(data, func(entity string) string {
		name := entity[1 : len(entity)-1]
		if val, ok := em[name]; ok {
			return string(val)
		}
		// Not found, return original
		return entity
	})
}
