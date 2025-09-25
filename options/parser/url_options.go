package optionsparser

import (
	"strings"
)

type urlOption struct {
	Name string
	Args []string
}

type urlOptions []urlOption

func (p *Parser) parseURLOptions(opts []string) (urlOptions, []string) {
	parsed := make(urlOptions, 0, len(opts))
	urlStart := len(opts) + 1

	for i, opt := range opts {
		args := strings.Split(opt, p.config.ArgumentsSeparator)

		if len(args) == 1 {
			urlStart = i
			break
		}

		parsed = append(parsed, urlOption{Name: args[0], Args: args[1:]})
	}

	var rest []string

	if urlStart < len(opts) {
		rest = opts[urlStart:]
	} else {
		rest = []string{}
	}

	return parsed, rest
}
