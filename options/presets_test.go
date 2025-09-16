package options

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type PresetsTestSuite struct {
	testutil.LazySuite

	security *security.Checker
}

func (s *PresetsTestSuite) SetupSuite() {
	c := security.NewDefaultConfig()
	security, err := security.New(&c)
	s.Require().NoError(err)
	s.security = security
}

func (s *PresetsTestSuite) newFactory(presets ...string) (*Factory, error) {
	c := NewDefaultConfig()
	c.Presets = presets
	return NewFactory(&c, s.security)
}

func (s *PresetsTestSuite) TestParsePreset() {
	f, err := s.newFactory("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Equal(urlOptions{
		urlOption{Name: "resize", Args: []string{"fit", "100", "200"}},
		urlOption{Name: "sharpen", Args: []string{"2"}},
	}, f.presets["test"])
}

func (s *PresetsTestSuite) TestParsePresetInvalidString() {
	presetStr := "resize:fit:100:200/sharpen:2"
	_, err := s.newFactory(presetStr)

	s.Require().Error(err, "invalid preset string: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyName() {
	presetStr := "=resize:fit:100:200/sharpen:2"
	_, err := s.newFactory(presetStr)

	s.Require().Error(err, "empty preset name: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyValue() {
	presetStr := "test="
	_, err := s.newFactory(presetStr)

	s.Require().Error(err, "empty preset value: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetInvalidValue() {
	presetStr := "test=resize:fit:100:200/sharpen:2/blur"
	_, err := s.newFactory(presetStr)

	s.Require().Error(err, "invalid preset value: %s", presetStr)
}

func (s *PresetsTestSuite) TestParsePresetEmptyString() {
	f, err := s.newFactory("   ")

	s.Require().NoError(err)
	s.Require().Empty(f.presets)
}

func (s *PresetsTestSuite) TestParsePresetComment() {
	f, err := s.newFactory("#  test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().Empty(f.presets)
}

func (s *PresetsTestSuite) TestValidatePresets() {
	f, err := s.newFactory("test=resize:fit:100:200/sharpen:2")

	s.Require().NoError(err)
	s.Require().NotEmpty(f.presets)
}

func (s *PresetsTestSuite) TestValidatePresetsInvalid() {
	_, err := s.newFactory("test=resize:fit:-1:-2/sharpen:2")

	s.Require().Error(err)
}

func TestPresets(t *testing.T) {
	suite.Run(t, new(PresetsTestSuite))
}
