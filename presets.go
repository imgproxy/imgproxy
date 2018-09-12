package main

import (
	"log"
	"strings"
)

type presets map[string]urlOptions

func parsePreset(p presets, presetStr string) {
	presetStr = strings.Trim(presetStr, " ")

	if len(presetStr) == 0 || strings.HasPrefix(presetStr, "#") {
		return
	}

	parts := strings.Split(presetStr, "=")

	if len(parts) != 2 {
		log.Fatalf("Invalid preset string: %s", presetStr)
		return
	}

	name := strings.Trim(parts[0], " ")
	if len(name) == 0 {
		log.Fatalf("Empty preset name: %s", presetStr)
		return
	}

	value := strings.Trim(parts[1], " ")
	if len(value) == 0 {
		log.Fatalf("Empty preset value: %s", presetStr)
		return
	}

	optsStr := strings.Split(value, "/")

	opts, rest := parseURLOptions(optsStr)

	if len(rest) > 0 {
		log.Fatalf("Invalid preset value: %s", presetStr)
	}

	p[name] = opts
}

func checkPresets(p presets) {
	var po processingOptions

	for name, opts := range p {
		if err := applyProcessingOptions(&po, opts); err != nil {
			log.Fatalf("Error in preset `%s`: %s", name, err)
		}
	}
}
