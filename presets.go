package main

import "strings"

type presets map[string]urlOptions

func parsePreset(p *presets, presetStr string) {
	presetStr = strings.Trim(presetStr, " ")

	if len(presetStr) == 0 || strings.HasPrefix(presetStr, "#") {
		return
	}

	parts := strings.Split(presetStr, "=")

	if len(parts) != 2 {
		warning("Invalid preset string, omitted: %s", presetStr)
		return
	}

	name := strings.Trim(parts[0], " ")
	if len(name) == 0 {
		warning("Empty preset name, omitted: %s", presetStr)
		return
	}

	value := strings.Trim(parts[1], " ")
	if len(value) == 0 {
		warning("Empty preset value, omitted: %s", presetStr)
		return
	}

	optsStr := strings.Split(value, "/")

	if opts, rest := parseURLOptions(optsStr); len(rest) == 0 {
		(*p)[name] = opts
	} else {
		warning("Invalid preset value, omitted: %s", presetStr)
	}
}
