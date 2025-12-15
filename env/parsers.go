package env

import (
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// URLReplacement represents a URL replacement configuration
type URLReplacement struct {
	Regexp      *regexp.Regexp
	Replacement string
}

// parseString returns the environment variable value as-is.
func parseString(env string) (string, error) {
	return env, nil
}

// parseEnumValue returns a parser function that looks up the input value in a map.
// The input value is trimmed before lookup. It assumes that m contains lowercase keys.
func parseEnumValue[T any](m map[string]T) ParseFn[T] {
	return func(env string) (T, error) {
		var zero T
		env = strings.ToLower(env)
		if val, ok := m[env]; ok {
			return val, nil
		}
		return zero, fmt.Errorf("invalid value %q", env)
	}
}

// parseFloat parses the environment variable value as a 64-bit float.
func parseFloat(env string) (float64, error) {
	return strconv.ParseFloat(env, 64)
}

// parseMegaInt parses a float value and multiplies it by 1,000,000 to convert to an integer.
// This is useful for environment variables that accept values like "1.5" to mean 1,500,000.
func parseMegaInt(env string) (int, error) {
	value, err := strconv.ParseFloat(env, 64)
	if err != nil {
		return 0, err
	}
	return int(value * 1_000_000), nil
}

// parseDuration parses an integer as seconds and returns a time.Duration.
func parseDuration(env string) (time.Duration, error) {
	value, err := strconv.Atoi(env)
	if err != nil {
		return 0, err
	}
	return time.Duration(value) * time.Second, nil
}

// parseDurationMillis parses an integer as milliseconds and returns a time.Duration.
func parseDurationMillis(env string) (time.Duration, error) {
	value, err := strconv.Atoi(env)
	if err != nil {
		return 0, err
	}
	return time.Duration(value) * time.Millisecond, nil
}

// parseStringSlice parses a comma-separated list of strings, trimming whitespace from each element.
func parseStringSlice(env string) ([]string, error) {
	parts := strings.Split(env, ",")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result, nil
}

// parseStringSliceSep returns a parser that splits by a custom separator.
// The separator is obtained from the provided StringVar descriptor.
func parseStringSliceSep(separatorDesc StringVar) ParseFn[[]string] {
	return func(env string) ([]string, error) {
		sep, ok := separatorDesc.GetEnv()
		if !ok || sep == "" {
			sep = "," // default to comma if separator not provided
		}
		parts := strings.Split(env, sep)
		result := make([]string, len(parts))
		for i, p := range parts {
			result[i] = strings.TrimSpace(p)
		}
		return result, nil
	}
}

// parseURLPath normalizes a URL path by removing query strings and fragments,
// ensuring it has a leading slash and no trailing slash.
func parseURLPath(env string) (string, error) {
	// Remove query string
	if i := strings.IndexByte(env, '?'); i >= 0 {
		env = env[:i]
	}
	// Remove fragment
	if i := strings.IndexByte(env, '#'); i >= 0 {
		env = env[:i]
	}
	// Remove trailing slash
	if len(env) > 0 && env[len(env)-1] == '/' {
		env = env[:len(env)-1]
	}
	// Ensure leading slash
	if len(env) > 0 && env[0] != '/' {
		env = "/" + env
	}
	return env, nil
}

// parseImageTypes parses a comma-separated list of image format names and returns
// a slice of imagetype.Type values.
func parseImageTypes(env string) ([]imagetype.Type, error) {
	parts := strings.Split(env, ",")
	result := make([]imagetype.Type, 0, len(parts))

	for _, p := range parts {
		part := strings.TrimSpace(p)
		t, ok := imagetype.GetTypeByName(part)
		if !ok {
			return nil, fmt.Errorf("unknown image format: %s", part)
		}
		result = append(result, t)
	}

	return result, nil
}

// parseImageTypesQuality parses format=quality pairs (e.g., "jpg=80,webp=90") and returns
// a map of image types to their quality values (1-100).
func parseImageTypesQuality(env string) (map[imagetype.Type]int, error) {
	result := make(map[imagetype.Type]int)
	parts := strings.SplitSeq(env, ",")

	for p := range parts {
		p = strings.TrimSpace(p)

		imgtypeStr, qStr, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid format quality string: %s", p)
		}

		imgtypeStr = strings.TrimSpace(imgtypeStr)
		qStr = strings.TrimSpace(qStr)

		if len(qStr) == 0 {
			return nil, fmt.Errorf("missing quality for format: %s", imgtypeStr)
		}

		if len(imgtypeStr) == 0 {
			return nil, fmt.Errorf("missing image format in: %s", p)
		}

		q, err := strconv.Atoi(qStr)
		if err != nil || q <= 0 || q > 100 {
			return nil, fmt.Errorf("invalid quality: %s", qStr)
		}

		t, ok := imagetype.GetTypeByName(imgtypeStr)
		if !ok {
			return nil, fmt.Errorf("unknown image format: %s", imgtypeStr)
		}

		result[t] = q
	}

	return result, nil
}

// parseURLPatterns parses a comma-separated list of wildcard patterns and converts them
// to compiled regular expressions using regexpFromPattern.
func parseURLPatterns(env string) ([]*regexp.Regexp, error) {
	parts := strings.Split(env, ",")
	result := make([]*regexp.Regexp, len(parts))

	for i, p := range parts {
		result[i] = regexpFromPattern(strings.TrimSpace(p))
	}

	return result, nil
}

// parseHexSlice parses a comma-separated list of hex-encoded strings and returns
// a slice of byte slices.
func parseHexSlice(env string) ([][]byte, error) {
	parts := strings.Split(env, ",")
	result := make([][]byte, len(parts))

	for i, part := range parts {
		b, err := hex.DecodeString(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("%s expected to be hex-encoded string: %w", part, err)
		}
		result[i] = b
	}

	return result, nil
}

// parseStringMap parses semicolon-separated key=value pairs and returns a map.
// Empty entries are skipped.
func parseStringMap(env string) (map[string]string, error) {
	result := make(map[string]string)
	keyvalues := strings.SplitSeq(env, ";")

	for kv := range keyvalues {
		kv = strings.TrimSpace(kv)
		if len(kv) == 0 {
			continue
		}

		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key/value: %s", kv)
		}
		result[parts[0]] = parts[1]
	}

	return result, nil
}

// regexpFromPattern creates a regexp from a wildcard pattern.
// Converts shell-style wildcards to regexp patterns.
func regexpFromPattern(pattern string) *regexp.Regexp {
	var result strings.Builder
	// Perform prefix matching
	result.WriteString("^")
	for i, part := range strings.Split(pattern, "*") {
		// Add a regexp match all without slashes for each wildcard character
		if i > 0 {
			result.WriteString("([^/]*)")
		}
		// Quote other parts of the pattern
		result.WriteString(regexp.QuoteMeta(part))
	}
	// It is safe to use regexp.MustCompile since the expression is always valid
	return regexp.MustCompile(result.String())
}

// parseStringSliceFile returns a parser that reads a string slice from a file.
// It reads the file path from CLI args (--{cliArgName}) or falls back to the env var.
// Each line in the file becomes an entry. Empty lines and lines starting with # are ignored.
func parseStringSliceFile(env string) ([]string, error) {
	// If no path provided, return empty slice
	if env == "" {
		return nil, nil
	}

	content, err := os.ReadFile(env)
	if err != nil {
		return nil, fmt.Errorf("can't read file %s: %w", env, err)
	}

	result := make([]string, 0)

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		result = append(result, line)
	}

	return result, nil
}

// parseURLReplacements parses URL replacements from the environment variable
func parseURLReplacements(env string) ([]URLReplacement, error) {
	s := []URLReplacement(nil)

	keyvalues := strings.SplitSeq(env, ";")

	for keyvalue := range keyvalues {
		parts := strings.SplitN(keyvalue, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key/value: %s", keyvalue)
		}
		s = append(s, URLReplacement{
			Regexp:      regexpFromPattern(parts[0]),
			Replacement: parts[1],
		})
	}

	return s, nil
}
