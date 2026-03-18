package env_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVar = "TEST_ENV_VAR"

func TestString(t *testing.T) {
	t.Setenv(testVar, "hello world")
	desc := env.String(testVar)

	var result string
	require.NoError(t, desc.Parse(&result))

	assert.Equal(t, "hello world", result)
}

func TestBool(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{input: "true", want: true},
		{input: "false", want: false},
		{input: "1", want: true},
		{input: "0", want: false},
		{input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.Bool(testVar)

			var result bool
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{input: "3.14", want: 3.14},
		{input: "42", want: 42.0},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.Float(testVar)

			var result float64
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InEpsilon(t, tt.want, result, 1e-9)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "42", want: 42},
		{input: "-1", want: -1},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.Int(testVar)

			var result int
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestMegaInt(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "1.5", want: 1_500_000},
		{input: "2", want: 2_000_000},
		{input: "0.5", want: 500_000},
		{input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.MegaInt(testVar)

			var result int
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{input: "30", want: 30 * time.Second},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.Duration(testVar)

			var result time.Duration
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestDurationMillis(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{input: "500", want: 500 * time.Millisecond},
		{input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.DurationMillis(testVar)

			var result time.Duration
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestEnum(t *testing.T) {
	enumMap := map[string]int{
		"small":  1,
		"medium": 2,
		"large":  3,
	}

	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{input: "small", want: 1},
		{input: "LARGE", want: 3},
		{input: "MeDiUm", want: 2},
		{input: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.Enum(testVar, enumMap)

			var result int
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestStringSlice(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "one,two,three", want: []string{"one", "two", "three"}},
		{input: " one , two , three ", want: []string{"one", "two", "three"}},
		{input: "single", want: []string{"single"}},
		{input: "", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.StringSlice(testVar)

			var result []string
			require.NoError(t, desc.Parse(&result))

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestStringSliceSep(t *testing.T) {
	const sepVar = "TEST_SEP_VAR"

	tests := []struct {
		sep   string
		input string
		want  []string
	}{
		{sep: "|", input: "one|two|three", want: []string{"one", "two", "three"}},
		{sep: "", input: "one,two,three", want: []string{"one", "two", "three"}},
		{
			sep:   ";",
			input: " one ; two ; three ",
			want:  []string{"one", "two", "three"},
		},
		{sep: "", input: "", want: nil},
	}

	for _, tt := range tests {
		name := tt.input + "=>" + fmt.Sprint(tt.want)
		t.Run(name, func(t *testing.T) {
			if tt.sep != "" {
				t.Setenv(sepVar, tt.sep)
			}
			t.Setenv(testVar, tt.input)

			sepDesc := env.String(sepVar)
			desc := env.StringSliceSep(testVar, sepDesc)

			var result []string
			require.NoError(t, desc.Parse(&result))

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestURLPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "path/to/resource", want: "/path/to/resource"},
		{input: "/path/to/resource/", want: "/path/to/resource"},
		{input: "/path?query=value", want: "/path"},
		{input: "/path#fragment", want: "/path"},
		{input: "path/to/resource/?query=value#fragment", want: "/path/to/resource"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input+"=>"+tt.want, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.URLPath(testVar)

			var result string
			require.NoError(t, desc.Parse(&result))

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestImageTypes(t *testing.T) {
	tests := []struct {
		input     string
		want      []imagetype.Type
		wantErr   bool
		errSubstr string
	}{
		{
			input: "jpg,png,webp",
			want:  []imagetype.Type{imagetype.JPEG, imagetype.PNG, imagetype.WEBP},
		},
		{
			input: " jpg , png , webp ",
			want:  []imagetype.Type{imagetype.JPEG, imagetype.PNG, imagetype.WEBP},
		},
		{input: "jpg,unknown,png", wantErr: true, errSubstr: "unknown image format"},
		{input: "", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.ImageTypes(testVar)

			var result []imagetype.Type
			err := desc.Parse(&result)

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

func TestImageTypesQuality(t *testing.T) {
	tests := []struct {
		input     string
		want      map[imagetype.Type]int
		wantErr   bool
		errSubstr string
	}{
		{
			input: "jpg=80,png=90,webp=85",
			want: map[imagetype.Type]int{
				imagetype.JPEG: 80,
				imagetype.PNG:  90,
				imagetype.WEBP: 85,
			},
		},
		{
			input: " jpg = 80 , png = 90 ",
			want:  map[imagetype.Type]int{imagetype.JPEG: 80, imagetype.PNG: 90},
		},
		{input: "jpg80", wantErr: true, errSubstr: "invalid format quality string"},
		{input: "jpg=invalid", wantErr: true, errSubstr: "invalid quality"},
		{input: "jpg=150", wantErr: true, errSubstr: "invalid quality"},
		{input: "unknown=80", wantErr: true, errSubstr: "unknown image format"},
		{input: "", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.ImageTypesQuality(testVar)

			var result map[imagetype.Type]int
			err := desc.Parse(&result)

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
}

func TestURLPatterns(t *testing.T) {
	tests := []struct {
		input         string
		wantLen       int
		shouldMatch   []string
		shouldntMatch []string
	}{
		{input: "", wantLen: 0},
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
			t.Setenv(testVar, tt.input)
			desc := env.URLPatterns(testVar)

			var result []*regexp.Regexp
			require.NoError(t, desc.Parse(&result))

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

func TestHexSlice(t *testing.T) {
	tests := []struct {
		input     string
		want      [][]byte
		wantErr   bool
		errSubstr string
	}{
		{
			input: "48656c6c6f,576f726c64",
			want:  [][]byte{[]byte("Hello"), []byte("World")},
		},
		{
			input: " 48656c6c6f , 576f726c64 ",
			want:  [][]byte{[]byte("Hello"), []byte("World")},
		},
		{input: "not-hex", wantErr: true, errSubstr: "expected to be hex-encoded"},
		{input: "", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.HexSlice(testVar)

			var result [][]byte
			err := desc.Parse(&result)

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
}

func TestStringMap(t *testing.T) {
	tests := []struct {
		input     string
		want      map[string]string
		wantErr   bool
		errSubstr string
	}{
		{
			input: "key1=value1;key2=value2",
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			input: " key1=value1 ; key2=value2 ",
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			input: "key1=value1;;key2=value2",
			want:  map[string]string{"key1": "value1", "key2": "value2"},
		},
		{input: "invalidentry", wantErr: true, errSubstr: "invalid key/value"},
		{input: "", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.StringMap(testVar)

			var result map[string]string
			err := desc.Parse(&result)

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
}

func TestStringSliceFile(t *testing.T) {
	tests := []struct {
		content string
		want    []string
	}{
		{content: "line1\nline2\nline3\n", want: []string{"line1", "line2", "line3"}},
		{
			content: "line1\n\nline2\n\n\nline3\n",
			want:    []string{"line1", "line2", "line3"},
		},
		{
			content: "line1\n# comment\nline2\n# another comment\nline3\n",
			want:    []string{"line1", "line2", "line3"},
		},
		{
			content: "  line1  \n  line2\nline3  \n",
			want:    []string{"line1", "line2", "line3"},
		},
		{content: "line1\nline2\nline3", want: []string{"line1", "line2", "line3"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.want), func(t *testing.T) {
			tmpfile, err := os.CreateTemp(t.TempDir(), "test-*.txt")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			_, err = tmpfile.WriteString(tt.content)
			require.NoError(t, err)
			tmpfile.Close()

			t.Setenv(testVar, tmpfile.Name())
			desc := env.StringSliceFile(testVar)

			var result []string
			require.NoError(t, desc.Parse(&result))

			assert.Equal(t, tt.want, result)
		})
	}

	t.Run("returns nil for empty path", func(t *testing.T) {
		t.Setenv(testVar, "")
		desc := env.StringSliceFile(testVar)

		var result []string
		require.NoError(t, desc.Parse(&result))

		assert.Nil(t, result)
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		t.Setenv(testVar, "/nonexistent/file.txt")
		desc := env.StringSliceFile(testVar)

		var result []string
		err := desc.Parse(&result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read file")
	})
}

func TestDateTime(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Time
		wantErr bool
	}{
		{input: "", want: time.Time{}},
		{
			input: "Mon, 02 Jan 2006 15:04:05 GMT",
			want:  time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{input: "invalid-date", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input+"=>"+tt.want.String(), func(t *testing.T) {
			t.Setenv(testVar, tt.input)
			desc := env.DateTime(testVar)

			var result time.Time
			err := desc.Parse(&result)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestRegexpFromPattern(t *testing.T) {
	tests := []struct {
		pattern       string
		shouldMatch   []string
		shouldntMatch []string
	}{
		{
			pattern:       "test.jpg",
			shouldMatch:   []string{"test.jpg"},
			shouldntMatch: []string{"test2.jpg"},
		},
		{
			pattern:       "*.jpg",
			shouldMatch:   []string{"image.jpg", "photo123.jpg"},
			shouldntMatch: []string{"dir/image.jpg"},
		},
		{
			pattern:       "*-*-*.jpg",
			shouldMatch:   []string{"img-01-thumb.jpg"},
			shouldntMatch: []string{"img-01.jpg"},
		},
		{
			pattern:       "prefix*",
			shouldMatch:   []string{"prefix123"},
			shouldntMatch: []string{"notprefix"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			re := env.RegexpFromPattern(tt.pattern)
			for _, match := range tt.shouldMatch {
				assert.True(t, re.MatchString(match))
			}
			for _, noMatch := range tt.shouldntMatch {
				assert.False(t, re.MatchString(noMatch))
			}
		})
	}
}
