package optionsparser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PresetsTestSuite struct {
	suite.Suite
}

func (s *PresetsTestSuite) TestParsePreset() {
	f, err := s.newParser("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Equal(urlOptions{
		urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
		urlOption{Name: "sharpen", Args: []string{"2"}},
	}, f.presets["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	presetStr := "resize:fit:100:200/sharpen:2"
	_, err := s.newParser(presetStr)

	s.Require().Error(err, "invalid preset string: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	presetStr := "=resize:fit:100:200/sharpen:2"
	_, err := s.newParser(presetStr)

	s.Require().Error(err, "empty preset name: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	presetStr := "test="
	_, err := s.newParser(presetStr)

	s.Require().Error(err, "empty preset value: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	_, err := s.newParser(presetStr)

	s.Require().Error(err, "invalid preset value: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	f, err := s.newParser("   ")

	s.Require().NoError(err)
	s.Require().Empty(f.presets)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	f, err := s.newParser("#  test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Empty(f.presets)
}

func (s *PresetsTestSuite) TestValidatePresets() {
	f, err := s.newParser("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().NotEmpty(f.presets)
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	_, err := s.newParser("test=resize:fit:-1:-2/sharpen:2")

	s.Require().Error(err)
}

func (s *PresetsTestSuite) newParser(presets ...string) (*Parser, error) {
	c := NewDefaultConfig()
	c.Presets = presets
	return New(s.T().Context(), &c)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
