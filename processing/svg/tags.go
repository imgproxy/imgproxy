package svg

import (
	"fmt"
	"strings"
)

// defaultTags lists SVG-namespace elements that survive sanitization.
// Names are lowercase: matching is case-insensitive.
//
// The base is DOMPurify's `svg` and `svgFilters` lists (src/tags.ts).
// On top of that we keep elements DOMPurify forbids but which are common
// in real-world files. They are made safe by the per-element checks in
// sanitizeElement (href rules, SMIL link targets, denied descendants):
// use, animate, set, foreignObject, SVG fonts and a few legacy elements.
var defaultTags = []string{
	// DOMPurify svg
	"svg", "a", "altglyph", "altglyphdef", "altglyphitem", "animatecolor",
	"animatemotion", "animatetransform", "circle", "clippath", "defs", "desc",
	"ellipse", "filter", "font", "g", "glyph", "glyphref", "hkern", "image",
	"line", "lineargradient", "marker", "mask", "metadata", "mpath", "path",
	"pattern", "polygon", "polyline", "radialgradient", "rect", "stop",
	"style", "switch", "symbol", "text", "textpath", "title", "tref", "tspan",
	"view", "vkern",

	// DOMPurify svgFilters
	"feblend", "fecolormatrix", "fecomponenttransfer", "fecomposite",
	"feconvolvematrix", "fediffuselighting", "fedisplacementmap",
	"fedistantlight", "fedropshadow", "feflood", "fefunca", "fefuncb",
	"fefuncg", "fefuncr", "fegaussianblur", "feimage", "femerge",
	"femergenode", "femorphology", "feoffset", "fepointlight",
	"fespecularlighting", "fespotlight", "fetile", "feturbulence",

	// Not in DOMPurify, kept for fidelity
	"animate", "set", "use", "foreignobject",
	"color-profile", "cursor", "discard",
	"font-face", "font-face-format", "font-face-name", "font-face-src",
	"font-face-uri", "missing-glyph",
	"hatch", "hatchpath", "solidcolor",
	"mesh", "meshgradient", "meshpatch", "meshrow",
}

// defaultTagSet is defaultTags as a set.
var defaultTagSet = newSet(defaultTags)

// deniedTags are removed with their subtree whatever namespace they are in,
// unless the user explicitly allows them via IMGPROXY_SANITIZE_SVG_TAGS.
var deniedTags = map[string]struct{}{
	"script": {},
	"iframe": {},
	"form":   {},
	"object": {},
	"embed":  {},
}

// applyListOverrides builds an allowlist from the defaults and the user's
// entries. Entries are "+name" (add), "-name" (remove) and "name" (replace).
//
// Bare entries are collected first: when there is at least one, they become
// the base list and the defaults are dropped. Otherwise the base is the
// defaults. Then "+" and "-" entries are applied to the base in order.
//
// Names are case-insensitive. Empty entries are skipped. An entry that is only
// a sign ("+" or "-") is an error, and no set is returned.
func applyListOverrides(defaults, entries []string) (map[string]struct{}, error) {
	var (
		bare []string
		// name -> true to add, false to remove. The last entry for a name wins.
		modifiers = make(map[string]bool)
	)

	for _, entry := range entries {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if len(entry) == 0 {
			continue
		}

		if entry[0] != '+' && entry[0] != '-' {
			bare = append(bare, entry)
			continue
		}

		name := strings.TrimSpace(entry[1:])
		if len(name) == 0 {
			return nil, fmt.Errorf("an empty +/- entry: %q", entry)
		}

		modifiers[name] = entry[0] == '+'
	}

	var set map[string]struct{}
	if len(bare) > 0 {
		set = newSet(bare)
	} else {
		// Defaults are already lowercase
		set = newSet(defaults)
	}

	for name, add := range modifiers {
		if add {
			set[name] = struct{}{}
		} else {
			delete(set, name)
		}
	}

	return set, nil
}
