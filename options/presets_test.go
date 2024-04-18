package options

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type PresetsTestSuite struct{ suite.Suite }

func (s *PresetsTestSuite) SetupTest() {
	config.Reset()
	// Reset presets
	presets = make(map[string]urlOptions)
}

func (s *PresetsTestSuite) TestParsePreset() {
	err := parsePreset("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)

	s.Require().Equal(urlOptions{
		urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
		urlOption{Name: "sharpen", Args: []string{"2"}},
	}, presets["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	presetStr := "resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	s.Require().Equal(fmt.Errorf("Invalid preset string: %s", presetStr), err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	presetStr := "=resize:fit:100:200/sharpen:2"
	err := parsePreset(presetStr)

	s.Require().Equal(fmt.Errorf("Empty preset name: %s", presetStr), err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	presetStr := "test="
	err := parsePreset(presetStr)

	s.Require().Equal(fmt.Errorf("Empty preset value: %s", presetStr), err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	err := parsePreset(presetStr)

	s.Require().Equal(fmt.Errorf("Invalid preset value: %s", presetStr), err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	err := parsePreset("  ")

	s.Require().NoError(err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	err := parsePreset("#  test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Empty(presets)
}

func (s *PresetsTestSuite) TestValidatePresets() {
	presets = map[string]urlOptions{
		"test": {
			urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	s.Require().NoError(err)
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	presets = map[string]urlOptions{
		"test": {
			urlOption{Name: "resize", Args: []string{"fit", "-1", "-2"}},
			urlOption{Name: "sharpen", Args: []string{"2"}},
		},
	}

	err := ValidatePresets()

	s.Require().Error(err)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
