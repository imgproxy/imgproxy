package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PresetsTestSuite struct{ MainTestSuite }

func (s *PresetsTestSuite) TestParsePreset() {
	p := make(presets)

	err := parsePreset(p, "test=resize:fit:100:200/sharpen:2")

	require.Nil(s.T(), err)

	assert.Equal(s.T(), urlOptions{
		"resize":  []string{"fit", "100", "200"},
		"sharpen": []string{"2"},
	}, p["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	p := make(presets)

	presetStr := "resize:fit:100:200/sharpen:2"
	err := parsePreset(p, presetStr)

	assert.Equal(s.T(), fmt.Errorf("Invalid preset string: %s", presetStr), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	p := make(presets)

	presetStr := "=resize:fit:100:200/sharpen:2"
	err := parsePreset(p, presetStr)

	assert.Equal(s.T(), fmt.Errorf("Empty preset name: %s", presetStr), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	p := make(presets)

	presetStr := "test="
	err := parsePreset(p, presetStr)

	assert.Equal(s.T(), fmt.Errorf("Empty preset value: %s", presetStr), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	p := make(presets)

	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	err := parsePreset(p, presetStr)

	assert.Equal(s.T(), fmt.Errorf("Invalid preset value: %s", presetStr), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	p := make(presets)

	err := parsePreset(p, "  ")

	assert.Nil(s.T(), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	p := make(presets)

	err := parsePreset(p, "#  test=resize:fit:100:200/sharpen:2")

	assert.Nil(s.T(), err)
	assert.Empty(s.T(), p)
}

func (s *PresetsTestSuite) TestCheckPresets() {
	p := presets{
		"test": urlOptions{
			"resize":  []string{"fit", "100", "200"},
			"sharpen": []string{"2"},
		},
	}

	err := checkPresets(p)

	assert.Nil(s.T(), err)
}

func (s *PresetsTestSuite) TestCheckPresetsInvalid() {
	p := presets{
		"test": urlOptions{
			"resize":  []string{"fit", "-1", "-2"},
			"sharpen": []string{"2"},
		},
	}

	err := checkPresets(p)

	assert.Error(s.T(), err)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
