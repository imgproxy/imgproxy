package env

import (
	"regexp"
	"strings"
)

// URLReplacement represents a URL replacement configuration
type URLReplacement struct {
	Regexp      *regexp.Regexp
	Replacement string
}

// URLReplacements parses URL replacements from the environment variable
func URLReplacements(s *[]URLReplacement, desc *StringVar) error {
	value, ok := desc.GetEnv()
	if !ok {
		return nil
	}

	ss := []URLReplacement(nil)

	keyvalues := strings.SplitSeq(value, ";")

	for keyvalue := range keyvalues {
		parts := strings.SplitN(keyvalue, "=", 2)
		if len(parts) != 2 {
			return desc.Errorf("invalid key/value: %s", keyvalue)
		}
		ss = append(ss, URLReplacement{
			Regexp:      regexpFromPattern(parts[0]),
			Replacement: parts[1],
		})
	}

	*s = ss

	return nil
}
