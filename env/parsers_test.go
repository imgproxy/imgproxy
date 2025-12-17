package env

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase[T any] struct {
	input     string
	want      T
	wantErr   bool
	errSubstr string
}

func (tc testCase[T]) name() string {
	if tc.wantErr {
		return tc.input + "=>error"
	}
	return tc.input + "=>" + fmt.Sprint(tc.want)
}

func (tc testCase[T]) assert(t *testing.T, result T, err error) {
	t.Helper()

	if tc.wantErr {
		require.Error(t, err)
		if tc.errSubstr != "" {
			assert.Contains(t, err.Error(), tc.errSubstr)
		}
	} else {
		require.NoError(t, err)
		assert.Equal(t, tc.want, result)
	}
}

func TestParseString(t *testing.T) {
	result, err := parseString("hello world")
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestParseEnumValue(t *testing.T) {
	enumMap := map[string]int{
		"small":  1,
		"medium": 2,
		"large":  3,
	}
	parser := parseEnumValue(enumMap)

	tests := []testCase[int]{
		{input: "small", want: 1},
		{input: "LARGE", want: 3},
		{input: "MeDiUm", want: 2},
		{
			input:     "invalid",
			wantErr:   true,
			errSubstr: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parser(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []testCase[float64]{
		{input: "3.14", want: 3.14},
		{input: "42", want: 42.0},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseFloat(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseMegaInt(t *testing.T) {
	tests := []testCase[int]{
		{input: "1.5", want: 1_500_000},
		{input: "2", want: 2_000_000},
		{input: "0.5", want: 500_000},
		{input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseMegaInt(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []testCase[time.Duration]{
		{input: "30", want: 30 * time.Second},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseDuration(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseDurationMillis(t *testing.T) {
	tests := []testCase[time.Duration]{
		{input: "500", want: 500 * time.Millisecond},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseDurationMillis(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseStringSlice(t *testing.T) {
	tests := []testCase[[]string]{
		{input: "one,two,three", want: []string{"one", "two", "three"}},
		{input: " one , two , three ", want: []string{"one", "two", "three"}},
		{input: "single", want: []string{"single"}},
		{input: "", want: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseStringSlice(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseStringSliceSep(t *testing.T) {
	tests := []struct {
		envValue string
		input    string
		want     []string
	}{
		{envValue: "|", input: "one|two|three", want: []string{"one", "two", "three"}},
		{envValue: "", input: "one,two,three", want: []string{"one", "two", "three"}},
		{envValue: ";", input: " one ; two ; three ", want: []string{"one", "two", "three"}},
	}

	for _, tt := range tests {
		name := tt.input + "=>" + fmt.Sprint(tt.want)
		t.Run(name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("TEST_SEP", tt.envValue)
			}
			sepDesc := String("TEST_SEP")
			parser := parseStringSliceSep(sepDesc)

			result, err := parser(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestParseURLPath(t *testing.T) {
	tests := []testCase[string]{
		{input: "path/to/resource", want: "/path/to/resource"},
		{input: "/path/to/resource/", want: "/path/to/resource"},
		{input: "/path?query=value", want: "/path"},
		{input: "/path#fragment", want: "/path"},
		{input: "path/to/resource/?query=value#fragment", want: "/path/to/resource"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseURLPath(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseImageTypes(t *testing.T) {
	tests := []testCase[[]imagetype.Type]{
		{input: "jpg,png,webp", want: []imagetype.Type{imagetype.JPEG, imagetype.PNG, imagetype.WEBP}},
		{input: " jpg , png , webp ", want: []imagetype.Type{imagetype.JPEG, imagetype.PNG, imagetype.WEBP}},
		{input: "jpg,unknown,png", wantErr: true, errSubstr: "unknown image format"},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseImageTypes(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.want, result)
			}
		})
	}
}

func TestParseImageTypesQuality(t *testing.T) {
	tests := []testCase[map[imagetype.Type]int]{
		{
			input: "jpg=80,png=90,webp=85",
			want: map[imagetype.Type]int{
				imagetype.JPEG: 80,
				imagetype.PNG:  90,
				imagetype.WEBP: 85,
			},
		},
		{input: " jpg = 80 , png = 90 ", want: map[imagetype.Type]int{imagetype.JPEG: 80, imagetype.PNG: 90}},
		{input: "jpg80", wantErr: true, errSubstr: "invalid format quality string"},
		{input: "jpg=invalid", wantErr: true, errSubstr: "invalid quality"},
		{input: "jpg=150", wantErr: true, errSubstr: "invalid quality"},
		{input: "unknown=80", wantErr: true, errSubstr: "unknown image format"},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseImageTypesQuality(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		input         string
		wantLen       int
		shouldMatch   []string
		shouldntMatch []string
	}{
		{input: "*.jpg,*.png", wantLen: 2},
		{
			input:         "image*.jpg",
			wantLen:       1,
			shouldMatch:   []string{"image123.jpg"},
			shouldntMatch: []string{"photo.jpg"},
		},
		{
			input:         "images/*/photo.jpg",
			wantLen:       1,
			shouldMatch:   []string{"images/vacation/photo.jpg"},
			shouldntMatch: []string{"images/a/b/photo.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseURLPatterns(tt.input)
			require.NoError(t, err)
			assert.Len(t, result, tt.wantLen)
			if len(result) > 0 {
				for _, match := range tt.shouldMatch {
					assert.True(t, result[0].MatchString(match))
				}
				for _, noMatch := range tt.shouldntMatch {
					assert.False(t, result[0].MatchString(noMatch))
				}
			}
		})
	}
}

func TestParseHexSlice(t *testing.T) {
	tests := []testCase[[][]byte]{
		{
			input: "48656c6c6f,576f726c64",
			want:  [][]byte{[]byte("Hello"), []byte("World")},
		},
		{
			input: " 48656c6c6f , 576f726c64 ",
			want:  [][]byte{[]byte("Hello"), []byte("World")},
		},
		{input: "not-hex", wantErr: true, errSubstr: "expected to be hex-encoded"},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseHexSlice(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestParseStringMap(t *testing.T) {
	tests := []testCase[map[string]string]{
		{input: "key1=value1;key2=value2", want: map[string]string{"key1": "value1", "key2": "value2"}},
		{input: " key1=value1 ; key2=value2 ", want: map[string]string{"key1": "value1", "key2": "value2"}},
		{input: "key1=value1;;key2=value2", want: map[string]string{"key1": "value1", "key2": "value2"}},
		{input: "", want: map[string]string{}},
		{input: "invalidentry", wantErr: true, errSubstr: "invalid key/value"},
	}

	for _, tt := range tests {
		t.Run(tt.name(), func(t *testing.T) {
			result, err := parseStringMap(tt.input)
			tt.assert(t, result, err)
		})
	}
}

func TestRegexpFromPattern(t *testing.T) {
	tests := []struct {
		pattern       string
		shouldMatch   []string
		shouldntMatch []string
	}{
		{pattern: "test.jpg", shouldMatch: []string{"test.jpg"}, shouldntMatch: []string{"test2.jpg"}},
		{pattern: "*.jpg", shouldMatch: []string{"image.jpg", "photo123.jpg"}, shouldntMatch: []string{"dir/image.jpg"}},
		{pattern: "*-*-*.jpg", shouldMatch: []string{"img-01-thumb.jpg"}, shouldntMatch: []string{"img-01.jpg"}},
		{pattern: "test.jpg", shouldMatch: []string{"test.jpg"}, shouldntMatch: []string{"testXjpg"}},
		{pattern: "prefix*", shouldMatch: []string{"prefix123"}, shouldntMatch: []string{"notprefix"}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := RegexpFromPattern(tt.pattern)
			for _, match := range tt.shouldMatch {
				assert.True(t, re.MatchString(match))
			}
			for _, noMatch := range tt.shouldntMatch {
				assert.False(t, re.MatchString(noMatch))
			}
		})
	}
}

func TestParseStringSliceFile(t *testing.T) {
	tests := []struct {
		content   string
		want      []string
		wantErr   bool
		errSubstr string
	}{
		{content: "line1\nline2\nline3\n", want: []string{"line1", "line2", "line3"}},
		{content: "line1\n\nline2\n\n\nline3\n", want: []string{"line1", "line2", "line3"}},
		{content: "line1\n# comment\nline2\n# another comment\nline3\n", want: []string{"line1", "line2", "line3"}},
		{content: "  line1  \n  line2\nline3  \n", want: []string{"line1", "line2", "line3"}},
		{content: "line1\nline2\nline3", want: []string{"line1", "line2", "line3"}},
	}

	for _, tt := range tests {
		name := fmt.Sprint(tt.want)
		if tt.wantErr {
			name = "error"
		}
		t.Run(name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp(t.TempDir(), "test-*.txt")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.WriteString(tt.content)
			require.NoError(t, err)
			tmpfile.Close()

			result, err := parseStringSliceFile(tmpfile.Name())
			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}

	t.Run("returns nil for empty path", func(t *testing.T) {
		result, err := parseStringSliceFile("")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := parseStringSliceFile("/nonexistent/file.txt")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read file")
	})
}
