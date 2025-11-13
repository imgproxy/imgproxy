package env

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// Int parses an integer from the environment variable
func Int(i *int, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	value, err := strconv.Atoi(env)
	if err != nil {
		return desc.ErrorParse(err)
	}
	*i = value

	return nil
}

// Float parses a float64 value from the environment variable
func Float(i *float64, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	value, err := strconv.ParseFloat(env, 64)
	if err != nil {
		return desc.ErrorParse(err)
	}
	*i = value

	return nil
}

// MegaInt parses a "megascale" integer from the environment variable
func MegaInt(f *int, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	value, err := strconv.ParseFloat(env, 64)
	if err != nil {
		return desc.ErrorParse(err)
	}
	*f = int(value) * 1_000_000

	return nil
}

// duration parses a duration (in resolution) from the environment variable
func duration(d *time.Duration, desc Desc, resolution time.Duration) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	value, err := strconv.Atoi(env)
	if err != nil {
		return desc.ErrorParse(err)
	}
	*d = time.Duration(value) * resolution

	return nil
}

// Duration parses a duration (in seconds) from the environment variable
func Duration(d *time.Duration, desc Desc) error {
	return duration(d, desc, time.Second)
}

// DurationMils parses a duration (in milliseconds) from the environment variable
func DurationMils(d *time.Duration, desc Desc) error {
	return duration(d, desc, time.Millisecond)
}

// String sets the string from the environment variable. Empty value is allowed.
func String(s *string, desc Desc) error {
	if env, ok := desc.Get(); ok {
		*s = env
	}

	return nil
}

// Bool parses a boolean from the environment variable
func Bool(b *bool, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	value, err := strconv.ParseBool(env)
	if err != nil {
		return desc.ErrorParse(err)
	}
	*b = value

	return nil
}

// StringSliceSep parses a string slice from the environment variable, using the given separator
func StringSliceSep(s *[]string, desc Desc, sep string) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	parts := strings.Split(env, sep)

	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}

	*s = parts

	return nil
}

// StringSliceFile parses a string slice from a file, one entry per line
func StringSliceFile(s *[]string, desc Desc, path string) error {
	if len(path) == 0 {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return desc.Errorf("can't open file %s", path)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := strings.TrimSpace(scanner.Text())
		if len(str) == 0 || strings.HasPrefix(str, "#") {
			continue
		}

		*s = append(*s, str)
	}

	if err := scanner.Err(); err != nil {
		return desc.Errorf("failed to read presets file: %s", err)
	}

	return nil
}

// StringSlice parses a string slice from the environment variable, using comma as a separator
func StringSlice(s *[]string, desc Desc) error {
	StringSliceSep(s, desc, ",")
	return nil
}

// URLPath parses and normalizes a URL path from the environment variable
func URLPath(s *string, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	if i := strings.IndexByte(env, '?'); i >= 0 {
		env = env[:i]
	}
	if i := strings.IndexByte(env, '#'); i >= 0 {
		env = env[:i]
	}
	if len(env) > 0 && env[len(env)-1] == '/' {
		env = env[:len(env)-1]
	}
	if len(env) > 0 && env[0] != '/' {
		env = "/" + env
	}

	*s = env

	return nil
}

// ImageTypes parses a slice of image types from the environment variable
func ImageTypes(it *[]imagetype.Type, desc Desc) error {
	// Get image types from environment variable
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	parts := strings.Split(env, ",")
	*it = make([]imagetype.Type, 0, len(parts))

	for _, p := range parts {
		part := strings.TrimSpace(p)

		// For every part passed through the environment variable,
		// check if it matches any of the image types defined in
		// the imagetype package or return error.
		t, ok := imagetype.GetTypeByName(part)
		if !ok {
			return desc.Errorf("unknown image format: %s", part)
		}
		*it = append(*it, t)
	}

	return nil
}

// ImageTypesQuality parses a string of format=queality pairs
func ImageTypesQuality(m map[imagetype.Type]int, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	parts := strings.SplitSeq(env, ",")

	for p := range parts {
		i := strings.Index(p, "=")
		if i < 0 {
			return desc.Errorf("invalid format quality string: %s", p)
		}

		// Split the string into image type and quality
		imgtypeStr, qStr := strings.TrimSpace(p[:i]), strings.TrimSpace(p[i+1:])

		// Check if quality is a valid integer
		q, err := strconv.Atoi(qStr)
		if err != nil || q <= 0 || q > 100 {
			return desc.Errorf("invalid quality: %s", p)
		}

		t, ok := imagetype.GetTypeByName(imgtypeStr)
		if !ok {
			return desc.Errorf("unknown image format: %s", imgtypeStr)
		}

		m[t] = q
	}

	return nil
}

// Patterns parses a slice of regexps from the environment variable
func Patterns(s *[]*regexp.Regexp, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	parts := strings.Split(env, ",")
	result := make([]*regexp.Regexp, len(parts))

	for i, p := range parts {
		result[i] = RegexpFromPattern(strings.TrimSpace(p))
	}

	*s = result

	return nil
}

// RegexpFromPattern creates a regexp from a wildcard pattern
func RegexpFromPattern(pattern string) *regexp.Regexp {
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

// HexSlice parses a slice of hex-encoded byte slices from the environment variable
func HexSlice(b *[][]byte, desc Desc) error {
	var err error

	env, ok := desc.Get()
	if !ok {
		return nil
	}

	parts := strings.Split(env, ",")
	keys := make([][]byte, len(parts))

	for i, part := range parts {
		if keys[i], err = hex.DecodeString(part); err != nil {
			return desc.Errorf("%s expected to be hex-encoded string", part)
		}
	}

	*b = keys

	return nil
}

// FromMap sets a value from a enum map based on the environment variable
func FromMap[T any](v *T, m map[string]T, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	if val, ok := m[env]; ok {
		*v = val
	} else {
		return desc.Errorf("%s", env)
	}

	return nil
}

// StringMap parses a map of string key-value pairs from the environment variable
func StringMap(m *map[string]string, desc Desc) error {
	env, ok := desc.Get()
	if !ok {
		return nil
	}

	mm := make(map[string]string)

	keyvalues := strings.SplitSeq(env, ";")

	for keyvalue := range keyvalues {
		parts := strings.SplitN(keyvalue, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid key/value: %s", keyvalue)
		}
		mm[parts[0]] = parts[1]
	}

	*m = mm

	return nil
}
