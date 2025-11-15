package xmlparser

import (
	"bytes"
	"regexp"
	"strings"
)

var entityDeclRe = regexp.MustCompile(`<!ENTITY\s+(\S+)\s+("[^"]+"|'[^']+')\s*>`)

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
	return replaceEntities(data, em, bytes.IndexByte)
}

// replaceEntitiesString replaces XML entities in the given string
// according to the provided entity map.
func replaceEntitiesString(data string, em map[string][]byte) string {
	return replaceEntities(data, em, strings.IndexByte)
}

func replaceEntities[T string | []byte](
	data T,
	em map[string][]byte,
	indexFn func(T, byte) int,
) T {
	var replaced []byte

	searchInd := 0

	for {
		// Find the entity start
		i := indexFn(data[searchInd:], '&')
		if i == -1 {
			break
		}
		i += searchInd

		// Find the entity end
		j := indexFn(data[i:], ';')
		if j == -1 {
			break
		}
		j += i

		// Get the entity name
		name := data[i+1 : j]

		// Find the replacement value
		if val, ok := em[string(name)]; ok {
			// If this is the first replacement, prealloc the replaced slice
			if len(replaced) == 0 {
				replaced = make([]byte, 0, len(data))
			}

			// Append the data before the entity and the replacement value
			replaced = append(replaced, data[:i]...)
			replaced = append(replaced, val...)

			// Move the data pointer forward and reset search index
			data = data[j+1:]
			searchInd = 0
		} else {
			// Didn't find replacement, just move the search index forward
			searchInd = j + 1
		}
	}

	// If no replacements were made, return the original data
	if len(replaced) == 0 {
		return data
	}

	// Append any remaining data after the last entity
	replaced = append(replaced, data...)

	return T(replaced)
}
