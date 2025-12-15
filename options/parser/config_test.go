package optionsparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPresetsPresent(t *testing.T) {
	// Setup environment
	t.Setenv("IMGPROXY_PRESETS_SEPARATOR", ",")
	t.Setenv("IMGPROXY_PRESETS", "preset1,preset2,preset3")

	// Load config
	cfg, err := LoadConfigFromEnv(nil)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, []string{"preset1", "preset2", "preset3"}, cfg.Presets)
}

func TestPresetsEmpty(t *testing.T) {
	// Setup environment with empty presets
	t.Setenv("IMGPROXY_PRESETS_SEPARATOR", ",")
	t.Setenv("IMGPROXY_PRESETS", "")

	// Load config
	cfg, err := LoadConfigFromEnv(nil)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Empty(t, cfg.Presets)
}

func TestPresetsFromFile(t *testing.T) {
	// Create temporary preset file
	tmpFile, err := os.CreateTemp(t.TempDir(), "presets-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write presets to file
	_, err = tmpFile.WriteString("file_preset1\nfile_preset2\n")
	require.NoError(t, err)
	tmpFile.Close()

	// Setup environment
	t.Setenv("IMGPROXY_PRESETS_PATH", tmpFile.Name())
	t.Setenv("IMGPROXY_PRESETS", "")

	// Load config
	cfg, err := LoadConfigFromEnv(nil)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, []string{"file_preset1", "file_preset2"}, cfg.Presets)
}

func TestPresetsFileEmpty(t *testing.T) {
	// Create temporary empty preset file
	tmpFile, err := os.CreateTemp(t.TempDir(), "presets-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Setup environment
	t.Setenv("IMGPROXY_PRESETS_PATH", tmpFile.Name())
	t.Setenv("IMGPROXY_PRESETS", "")

	// Load config
	cfg, err := LoadConfigFromEnv(nil)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Empty(t, cfg.Presets)
}

func TestPresetsFileNotFound(t *testing.T) {
	// Use a non-existent file path
	nonExistentPath := filepath.Join(os.TempDir(), "non-existent-presets-file-"+string(rune(os.Getpid()))+".txt")

	// Setup environment with non-existent file
	t.Setenv("IMGPROXY_PRESETS_PATH", nonExistentPath)
	t.Setenv("IMGPROXY_PRESETS", "")

	// Load config - should error because file doesn't exist
	_, err := LoadConfigFromEnv(nil)

	// Verify that error occurred due to missing file
	require.Error(t, err)
}

func TestPresetsFromBothSources(t *testing.T) {
	// Create temporary preset file
	tmpFile, err := os.CreateTemp(t.TempDir(), "presets-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write presets to file
	_, err = tmpFile.WriteString("file_preset1\nfile_preset2\n")
	require.NoError(t, err)
	tmpFile.Close()

	// Setup environment with both sources
	t.Setenv("IMGPROXY_PRESETS_SEPARATOR", ",")
	t.Setenv("IMGPROXY_PRESETS", "preset1,preset2")
	t.Setenv("IMGPROXY_PRESETS_PATH", tmpFile.Name())

	// Load config
	cfg, err := LoadConfigFromEnv(nil)

	// Verify both presets are loaded
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, []string{"preset1", "preset2", "file_preset1", "file_preset2"}, cfg.Presets)
}
