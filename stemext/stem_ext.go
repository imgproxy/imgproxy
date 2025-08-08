// stemext package provides methods which help to generate correct
// content-disposition header.
package contentdisposition

import (
	"mime"
	"net/url"
	"path/filepath"
)

const (
	// fallbackStem is used when the stem cannot be determined from the URL.
	fallbackStem = "image"
)

// StemExt helps to detect correct stem and ext for content-disposition header.
type StemExt struct {
	stem string
	ext  string
}

// FromURL creates a new StemExt instance from the provided URL.
// Returns a value type to avoid heap allocation.
func FromURL(url *url.URL) StemExt {
	_, filename := filepath.Split(url.Path)
	ext := filepath.Ext(filename)

	// Avoid strings.TrimSuffix allocation by using slice operation
	var stem string
	if ext != "" {
		stem = filename[:len(filename)-len(ext)]
	} else {
		stem = filename
	}

	return StemExt{
		stem: stem,
		ext:  ext,
	}
}

// SetExtFromContentTypeIfEmpty sets the ext field based on the provided content type.
// Uses pointer receiver for zero-copy method chaining.
func (cd *StemExt) SetExtFromContentTypeIfEmpty(contentType string) *StemExt {
	if len(contentType) == 0 || len(cd.ext) > 0 {
		return cd
	}

	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) != 0 {
		cd.ext = exts[0]
	}

	return cd
}

// OverrideExt sets the ext field if the provided ext is not empty.
// Uses pointer receiver for zero-copy method chaining.
func (cd *StemExt) OverrideExt(ext string) *StemExt {
	if len(ext) > 0 {
		cd.ext = ext
	}

	return cd
}

// OverrideStem sets the stem field if the provided stem is not empty.
// Uses pointer receiver for zero-copy method chaining.
func (cd *StemExt) OverrideStem(stem string) *StemExt {
	if len(stem) > 0 {
		cd.stem = stem
	}

	return cd
}

// StemExtWithFallback returns stem and ext, but if stem is empty, it uses a fallback value.
func (cd StemExt) StemExtWithFallback() (string, string) {
	stem := cd.stem
	if len(stem) == 0 {
		stem = fallbackStem
	}

	return stem, cd.ext
}

// StemExt returns the tuple of stem and ext.
func (cd StemExt) StemExt() (string, string) {
	return cd.stem, cd.ext
}
