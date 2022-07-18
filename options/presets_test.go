package options

import (
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PresetsTestSuite struct{ suite.Suite }

func (s *PresetsTestSuite) SetupTest() {
	config.Reset()
	// Reset presets
	presets = make(map[string]urlOptions)
}

func (s *PresetsTestSuite) TestParsePreset() {
	err := parsePreset("test=resize:fit:100:200/sharpen:2")

	require.Nil(s.T(), err)

	require.Equal(s.T(), urlOptions{
		urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
		urlOption{Name: "sharpen", Args: []string{"2"}},
	}, presets["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	presetStr := "resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	require.Equal(s.T(), fmt.Errorf("Invalid preset string: %s", presetStr), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	presetStr := "=resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	require.Equal(s.T(), fmt.Errorf("Empty preset name: %s", presetStr), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	presetStr := "test="
	err := parsePreset(presetStr)

	require.Equal(s.T(), fmt.Errorf("Empty preset value: %s", presetStr), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	err := parsePreset(presetStr)

	require.Equal(s.T(), fmt.Errorf("Invalid preset value: %s", presetStr), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	err := parsePreset("  ")

	require.Nil(s.T(), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	err := parsePreset("#  test=resize:fit:100:200/sharpen:2")

	require.Nil(s.T(), err)
	require.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestValidatePresets() {
	presets = map[string]urlOptions{
		"test": urlOptions{
			urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	require.Nil(s.T(), err)
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	presets = map[string]urlOptions{
		"test": urlOptions{
			urlOption{Name: "resize", Args: []string{"fit", "-1", "-2"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	require.Error(s.T(), err)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
