package options

import (
	"fmt"
	"strings"
)

var presets map[string]urlOptions

func ParsePresets(presetStrs []string) error {
	for _, presetStr := range presetStrs {
		if err := parsePreset(presetStr); err != nil {
			return err
		}
	}

	return nil
}

func parsePreset(presetStr string) error {
	presetStr = strings.Trim(presetStr, " ")

	if len(presetStr) == 0 || strings.HasPrefix(presetStr, "#") {
		return nil
	}

	parts := strings.Split(presetStr, "=")

	if len(parts) != 2 {
		return fmt.Errorf("Invalid preset string: %s", presetStr)
	}

	name := strings.Trim(parts[0], " ")
	if len(name) == 0 {
		return fmt.Errorf("Empty preset name: %s", presetStr)
	}

	value := strings.Trim(parts[1], " ")
	if len(value) == 0 {
		return fmt.Errorf("Empty preset value: %s", presetStr)
	}

	optsStr := strings.Split(value, "/")

	opts, rest := parseURLOptions(optsStr)

	if len(rest) > 0 {
		return fmt.Errorf("Invalid preset value: %s", presetStr)
	}

	if presets == nil {
		presets = make(map[string]urlOptions)
	}
	presets[name] = opts

	return nil
}

func ValidatePresets() error {
	for name, opts := range presets {
		po := NewProcessingOptions()
		if err := applyURLOptions(po, opts); err != nil {
			return fmt.Errorf("Error in preset `%s`: %s", name, err)
		}
	}

	return nil
}
