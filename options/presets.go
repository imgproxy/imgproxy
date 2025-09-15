package options

import (
	"fmt"
	"strings"
)

// parsePresets parses presets from the config and fills the presets map
func (f *Factory) parsePresets() error {
	for _, presetStr := range f.config.Presets {
		if err := f.parsePreset(presetStr); err != nil {
			return err
		}
	}

	return nil
}

// validatePresets validates all presets by applying them to a new ProcessingOptions instance
func (f *Factory) validatePresets() error {
	for name, opts := range f.presets {
		po := f.NewProcessingOptions()
		if err := f.applyURLOptions(po, opts, true, name); err != nil {
			return fmt.Errorf("Error in preset `%s`: %s", name, err)
		}
	}

	return nil
}

// parsePreset parses a preset string and returns the name and options
func (f *Factory) parsePreset(presetStr string) error {
	presetStr = strings.Trim(presetStr, " ")

	if len(presetStr) == 0 || strings.HasPrefix(presetStr, "#") {
		return nil
	}

	parts := strings.Split(presetStr, "=")

	if len(parts) != 2 {
		return fmt.Errorf("invalid preset string: %s", presetStr)
	}

	name := strings.Trim(parts[0], " ")
	if len(name) == 0 {
		return fmt.Errorf("empty preset name: %s", presetStr)
	}

	value := strings.Trim(parts[1], " ")
	if len(value) == 0 {
		return fmt.Errorf("empty preset value: %s", presetStr)
	}

	optsStr := strings.Split(value, "/")

	opts, rest := f.parseURLOptions(optsStr)

	if len(rest) > 0 {
		return fmt.Errorf("invalid preset value: %s", presetStr)
	}

	if f.presets == nil {
		f.presets = make(Presets)
	}

	f.presets[name] = opts

	return nil
}
