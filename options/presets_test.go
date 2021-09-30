package options

import (
	"fmt"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(s.T(), urlOptions{
		urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
		urlOption{Name: "sharpen", Args: []string{"2"}},
	}, presets["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	presetStr := "resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	assert.Equal(s.T(), fmt.Errorf("Invalid preset string: %s", presetStr), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	presetStr := "=resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	assert.Equal(s.T(), fmt.Errorf("Empty preset name: %s", presetStr), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	presetStr := "test="
	err := parsePreset(presetStr)

	assert.Equal(s.T(), fmt.Errorf("Empty preset value: %s", presetStr), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	err := parsePreset(presetStr)

	assert.Equal(s.T(), fmt.Errorf("Invalid preset value: %s", presetStr), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	err := parsePreset("  ")

	assert.Nil(s.T(), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	err := parsePreset("#  test=resize:fit:100:200/sharpen:2")

	assert.Nil(s.T(), err)
	assert.Empty(s.T(), presets)
}

func (s *PresetsTestSuite) TestValidatePresets() {
	presets = map[string]urlOptions{
		"test": urlOptions{
			urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	assert.Nil(s.T(), err)
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	presets = map[string]urlOptions{
		"test": urlOptions{
			urlOption{Name: "resize", Args: []string{"fit", "-1", "-2"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	assert.Error(s.T(), err)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
