package svg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/processing/svg"
)

func TestLoadConfigFromEnvLists(t *testing.T) {
	t.Setenv("IMGPROXY_SANITIZE_SVG_TAGS", "+foo, -circle ,bar")
	t.Setenv("IMGPROXY_SANITIZE_SVG_ATTRS", "-fill,+data-x")

	c, err := svg.LoadConfigFromEnv(nil)
	require.NoError(t, err)

	require.Equal(t, []string{"+foo", "-circle", "bar"}, c.Tags)
	require.Equal(t, []string{"-fill", "+data-x"}, c.Attrs)
	require.NoError(t, c.Validate())
}

func TestConfigValidateSignWithoutName(t *testing.T) {
	c := svg.NewDefaultConfig()
	c.Tags = []string{"+"}
	require.Error(t, c.Validate())

	c = svg.NewDefaultConfig()
	c.Attrs = []string{"-fill", "-"}
	require.Error(t, c.Validate())
}

func TestConfigValidateDeniedTag(t *testing.T) {
	c := svg.NewDefaultConfig()
	c.Tags = []string{"+script"}
	require.NoError(t, c.Validate())
}
