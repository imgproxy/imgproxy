package optionsparser_test

import (
	"testing"

	optionsparser "github.com/imgproxy/imgproxy/v3/options/parser"
	"github.com/stretchr/testify/suite"
)

type PresetsTestSuite struct {
	suite.Suite
}

func (s *PresetsTestSuite) TestParsePreset() {
	f, err := s.newParser("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Equal([]optionsparser.URLOption{
		{Name: "resize", Args: []string{"fit", "100", "200"}},
		{Name: "sharpen", Args: []string{"2"}},
	}, f.Presets()["test"])
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
	s.Require().Empty(f.Presets())
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	f, err := s.newParser("#  test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Empty(f.Presets())
}

func (s *PresetsTestSuite) TestValidatePresets() {
	f, err := s.newParser("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().NotEmpty(f.Presets())
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	_, err := s.newParser("test=resize:fit:-1:-2/sharpen:2")

	s.Require().Error(err)
}

func (s *PresetsTestSuite) newParser(presets ...string) (*optionsparser.Parser, error) {
	c := optionsparser.NewDefaultConfig()
	c.Presets = presets
	return optionsparser.New(s.T().Context(), &c)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
