package env

import (
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type StringVar = Desc[string, ParseFn[string]]
type BoolVar = Desc[bool, ParseFn[bool]]
type IntVar = Desc[int, ParseFn[int]]
type FloatVar = Desc[float64, ParseFn[float64]]
type DurationVar = Desc[time.Duration, ParseFn[time.Duration]]
type StringSliceVar = Desc[[]string, ParseFn[[]string]]
type ImageTypesVar = Desc[[]imagetype.Type, ParseFn[[]imagetype.Type]]
type ImageTypesQualityVar = Desc[map[imagetype.Type]int, ParseFn[map[imagetype.Type]int]]
type URLPatternsVar = Desc[[]*regexp.Regexp, ParseFn[[]*regexp.Regexp]]
type HexSliceVar = Desc[[][]byte, ParseFn[[][]byte]]
type StringMapVar = Desc[map[string]string, ParseFn[map[string]string]]
type EnumVar[T any] = Desc[T, ParseFn[T]]

// String defines string env var descriptor
func String(name string) StringVar {
	return StringVar{
		Name:    name,
		format:  "string",
		parseFn: parseString,
	}
}

// Enum defines enum value env var descriptor that maps string keys to values of type T.
// The format string is generated from the map keys.
func Enum[T any](name string, m map[string]T) EnumVar[T] {
	keys := slices.Collect(maps.Keys(m))

	return EnumVar[T]{
		Name:    name,
		format:  strings.Join(keys, "|"),
		parseFn: parseEnumValue(m),
	}
}

// Bool defines boolean env var descriptor
func Bool(name string) BoolVar {
	return BoolVar{
		Name:    name,
		format:  "bool",
		parseFn: strconv.ParseBool,
	}
}

// Float defines float64 env var descriptor
func Float(name string) FloatVar {
	return FloatVar{
		Name:    name,
		format:  "float",
		parseFn: parseFloat,
	}
}

// Int defines integer env var descriptor
func Int(name string) IntVar {
	return IntVar{
		Name:    name,
		format:  "int",
		parseFn: strconv.Atoi,
	}
}

// MegaInt defines "megascale" integer env var descriptor.
// Parses a float value and multiplies by 1,000,000.
func MegaInt(name string) IntVar {
	return IntVar{
		Name:    name,
		format:  "float, multiplied by 1,000,000 internally",
		parseFn: parseMegaInt,
	}
}

// Duration defines duration env var descriptor.
// Parses an integer as seconds.
func Duration(name string) DurationVar {
	return DurationVar{
		Name:    name,
		format:  "seconds",
		parseFn: parseDuration,
	}
}

// DurationMillis defines duration env var descriptor.
// Parses an integer as milliseconds.
func DurationMillis(name string) DurationVar {
	return DurationVar{
		Name:    name,
		format:  "milliseconds",
		parseFn: parseDurationMillis,
	}
}

// StringSlice defines string slice env var descriptor.
// Parses a comma-separated list of strings.
func StringSlice(name string) StringSliceVar {
	return StringSliceVar{
		Name:    name,
		format:  "comma-separated strings",
		parseFn: parseStringSlice,
	}
}

// URLPath defines URL path env var descriptor.
// Normalizes the path by removing query strings, fragments, and ensuring proper slashes.
func URLPath(name string) StringVar {
	return StringVar{
		Name:    name,
		format:  "URL path",
		parseFn: parseURLPath,
	}
}

// ImageTypes defines image types slice env var descriptor.
// Parses a comma-separated list of image format names.
func ImageTypes(name string) ImageTypesVar {
	return ImageTypesVar{
		Name:    name,
		format:  "comma-separated image formats",
		parseFn: parseImageTypes,
	}
}

// ImageTypesQuality defines image format quality map env var descriptor.
// Parses format=quality pairs (e.g., "jpg=80,webp=90").
func ImageTypesQuality(name string) ImageTypesQualityVar {
	return ImageTypesQualityVar{
		Name:    name,
		format:  "format=quality pairs (e.g. jpg=80,webp=90)",
		parseFn: parseImageTypesQuality,
	}
}

// URLPatterns defines regexp patterns slice env var descriptor.
// Parses comma-separated wildcard patterns and converts them to regexps.
func URLPatterns(name string) URLPatternsVar {
	return URLPatternsVar{
		Name:    name,
		format:  "comma-separated wildcard patterns",
		parseFn: parseURLPatterns,
	}
}

// HexSlice defines hex-encoded byte slices env var descriptor.
// Parses comma-separated hex strings into byte slices.
func HexSlice(name string) HexSliceVar {
	return HexSliceVar{
		Name:    name,
		format:  "comma-separated hex-encoded strings",
		parseFn: parseHexSlice,
	}
}

// StringMap defines string key-value map env var descriptor.
// Parses semicolon-separated key=value pairs.
func StringMap(name string) StringMapVar {
	return StringMapVar{
		Name:    name,
		format:  "key=value pairs separated by semicolons",
		parseFn: parseStringMap,
	}
}

// StringSliceSep defines string slice env var descriptor with custom separator.
// The separator is read from the provided StringVar descriptor's env variable.
func StringSliceSep(name string, separatorDesc StringVar) StringSliceVar {
	return StringSliceVar{
		Name:    name,
		format:  "separated list of strings",
		parseFn: parseStringSliceSep(separatorDesc),
	}
}

// StringSliceFile defines string slice from file env var descriptor.
// The file path is read from CLI args (--{cliArgName}) or falls back to the env variable.
// Each line in the file becomes an entry. Empty lines and lines starting with # are ignored.
func StringSliceFile(name string) StringSliceVar {
	return StringSliceVar{
		Name:    name,
		format:  "path to file with one entry per line",
		parseFn: parseStringSliceFile,
	}
}
