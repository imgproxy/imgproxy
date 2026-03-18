package optionsparser

type URLOption = urlOption

func (p *Parser) Presets() map[string][]urlOption {
	return p.presets
}
