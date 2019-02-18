package main

import (
	"fmt"
	"strings"
)

type presets map[string]urlOptions

func parsePreset(p presets, presetStr string) error {
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

	p[name] = opts

	return nil
}

func checkPresets(p presets) error {
	var po processingOptions

	for name, opts := range p {
		if err := applyProcessingOptions(&po, opts); err != nil {
			return fmt.Errorf("Error in preset `%s`: %s", name, err)
		}
	}

	return nil
}
