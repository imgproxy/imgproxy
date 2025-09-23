package options

import (
	"fmt"
	"strings"
)

// parsePresets parses presets from the config and fills the presets map
func (p *Parser) parsePresets() error {
	for _, presetStr := range p.config.Presets {
		if err := p.parsePreset(presetStr); err != nil {
			return err
		}
	}

	return nil
}

// parsePreset parses a preset string and returns the name and options
func (p *Parser) parsePreset(presetStr string) error {
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

	opts, rest := p.parseURLOptions(optsStr)

	if len(rest) > 0 {
		return fmt.Errorf("invalid preset value: %s", presetStr)
	}

	if p.presets == nil {
		p.presets = make(Presets)
	}

	p.presets[name] = opts

	return nil
}

// validatePresets validates all presets by applying them to a new Options instance
func (p *Parser) validatePresets() error {
	for name, opts := range p.presets {
		po := New()
		if err := p.applyURLOptions(po, opts, true, name); err != nil {
			return fmt.Errorf("Error in preset `%s`: %s", name, err)
		}
	}

	return nil
}
