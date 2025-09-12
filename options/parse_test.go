package options

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase[T ~int | ~int64 | ~float32 | ~float64] struct {
	name      string
	arg       string
	expected  T
	expectErr bool
}

func runParseNumericTests[T ~int | ~int64 | ~float32 | ~float64](t *testing.T, testCases []testCase[T]) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseNumber[T](tc.arg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestParseNumericInt(t *testing.T) {
	testCases := []testCase[int]{
		{
			name:      "valid positive integer",
			arg:       "42",
			expected:  42,
			expectErr: false,
		},
		{
			name:      "valid negative integer",
			arg:       "-10",
			expected:  -10,
			expectErr: false,
		},
		{
			name:      "valid zero",
			arg:       "0",
			expected:  0,
			expectErr: false,
		},
		{
			name:      "invalid non-numeric string",
			arg:       "abc",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "empty string",
			arg:       "",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "float string for int",
			arg:       "3.14",
			expected:  0,
			expectErr: true,
		},
	}

	runParseNumericTests(t, testCases)
}

func TestParseNumericInt64(t *testing.T) {
	testCases := []testCase[int64]{
		{
			name:      "valid large positive integer",
			arg:       "1234567890",
			expected:  1234567890,
			expectErr: false,
		},
		{
			name:      "valid large negative integer",
			arg:       "-1234567890",
			expected:  -1234567890,
			expectErr: false,
		},
		{
			name:      "valid minimum boundary",
			arg:       "100",
			expected:  100,
			expectErr: false,
		},
		{
			name:      "valid maximum boundary",
			arg:       "200",
			expected:  200,
			expectErr: false,
		},
		{
			name:      "invalid string",
			arg:       "not_a_number",
			expected:  0,
			expectErr: true,
		},
	}

	runParseNumericTests(t, testCases)
}

func TestParseNumericFloat32(t *testing.T) {
	testCases := []testCase[float32]{
		{
			name:      "valid positive float",
			arg:       "3.14",
			expected:  3.14,
			expectErr: false,
		},
		{
			name:      "valid negative float",
			arg:       "-2.5",
			expected:  -2.5,
			expectErr: false,
		},
		{
			name:      "valid zero float",
			arg:       "0.0",
			expected:  0.0,
			expectErr: false,
		},
		{
			name:      "valid integer as float",
			arg:       "42",
			expected:  42.0,
			expectErr: false,
		},
		{
			name:      "invalid string",
			arg:       "invalid",
			expected:  0.0,
			expectErr: true,
		},
		{
			name:      "scientific notation",
			arg:       "1.23e2",
			expected:  123.0,
			expectErr: false,
		},
	}

	runParseNumericTests(t, testCases)
}

func TestParseNumericFloat64(t *testing.T) {
	testCases := []testCase[float64]{
		{
			name:      "valid high precision float",
			arg:       "3.141592653589793",
			expected:  3.141592653589793,
			expectErr: false,
		},
		{
			name:      "valid negative high precision float",
			arg:       "-2.718281828459045",
			expected:  -2.718281828459045,
			expectErr: false,
		},
		{
			name:      "valid very small float",
			arg:       "1e-10",
			expected:  1e-10,
			expectErr: false,
		},
		{
			name:      "valid very large float",
			arg:       "1.23456789e20",
			expected:  1.23456789e20,
			expectErr: false,
		},
		{
			name:      "invalid string",
			arg:       "not_a_float",
			expected:  0.0,
			expectErr: true,
		},
	}

	runParseNumericTests(t, testCases)
}

func TestParseGravity(t *testing.T) {
	type tc struct {
		name         string
		args         []string
		allowedTypes []GravityType
		expected     GravityOptions
		expectErr    bool
	}

	testCases := []tc{
		{
			name:         "smart gravity only",
			args:         []string{"sm"},
			allowedTypes: []GravityType{GravitySmart, GravityCenter, GravityNorth},
			expected:     GravityOptions{Type: GravitySmart, X: 0.0, Y: 0.0},
			expectErr:    false,
		},
		{
			name:         "center gravity only",
			args:         []string{"ce"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth, GravitySouth},
			expected:     GravityOptions{Type: GravityCenter, X: 0.0, Y: 0.0},
			expectErr:    false,
		},
		{
			name:         "center gravity with X offset",
			args:         []string{"ce", "0.5"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth, GravitySouth},
			expected:     GravityOptions{Type: GravityCenter, X: 0.5, Y: 0.0},
			expectErr:    false,
		},
		{
			name:         "center gravity with X and Y offsets",
			args:         []string{"ce", "0.3", "0.7"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth, GravitySouth},
			expected:     GravityOptions{Type: GravityCenter, X: 0.3, Y: 0.7},
			expectErr:    false,
		},
		{
			name:         "focus point gravity",
			args:         []string{"fp", "0.4", "0.6"},
			allowedTypes: []GravityType{GravityFocusPoint, GravityCenter},
			expected:     GravityOptions{Type: GravityFocusPoint, X: 0.4, Y: 0.6},
			expectErr:    false,
		},
		{
			name:         "north gravity",
			args:         []string{"no"},
			allowedTypes: []GravityType{GravityNorth, GravityCenter, GravitySouth},
			expected:     GravityOptions{Type: GravityNorth, X: 0.0, Y: 0.0},
			expectErr:    false,
		},
		{
			name:         "north gravity with offsets",
			args:         []string{"no", "10", "20"},
			allowedTypes: []GravityType{GravityNorth, GravityCenter, GravitySouth},
			expected:     GravityOptions{Type: GravityNorth, X: 10.0, Y: 20.0},
			expectErr:    false,
		},
		{
			name:         "invalid gravity type",
			args:         []string{"invalid"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth},
			expectErr:    true,
		},
		{
			name:         "gravity type not allowed",
			args:         []string{"ce"},
			allowedTypes: []GravityType{GravityNorth, GravitySouth},
			expectErr:    true,
		},
		{
			name:         "smart gravity with extra args",
			args:         []string{"sm", "0.5"},
			allowedTypes: []GravityType{GravitySmart, GravityCenter},
			expectErr:    true,
		},
		{
			name:         "focus point with too few args",
			args:         []string{"fp", "0.5"},
			allowedTypes: []GravityType{GravityFocusPoint, GravityCenter},
			expectErr:    true,
		},
		{
			name:         "focus point with too many args",
			args:         []string{"fp", "0.5", "0.6", "0.7"},
			allowedTypes: []GravityType{GravityFocusPoint, GravityCenter},
			expectErr:    true,
		},
		{
			name:         "center gravity with too many args",
			args:         []string{"ce", "0.1", "0.2", "0.3", "0.4"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth},
			expectErr:    true,
		},
		{
			name:         "invalid X offset",
			args:         []string{"ce", "invalid"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth},
			expectErr:    true,
		},
		{
			name:         "invalid Y offset",
			args:         []string{"ce", "0.5", "invalid"},
			allowedTypes: []GravityType{GravityCenter, GravityNorth},
			expectErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var g GravityOptions
			err := parseGravity(&g, "test", tc.args, tc.allowedTypes)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected.Type, g.Type)

				// Use InDelta for float comparison to avoid precision issues
				// Required by go vet
				require.InDelta(t, tc.expected.X, g.X, 0)
				require.InDelta(t, tc.expected.Y, g.Y, 0)
			}
		})
	}
}
