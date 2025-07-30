// stemext package provides methods which help to detect correct stem and ext
// for the content-disposition header.
package stemext

import (
	"mime"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	// fallbackStem is used when the stem cannot be determined from the URL.
	fallbackStem = "image"
)

// stemExt helps to detect correct stem and ext for content-disposition header.
type stemExt struct {
	stem string
	ext  string
}

// FromURL creates a new detectStemExt instance from the provided URL.
func FromURL(url *url.URL) *stemExt {
	_, filename := filepath.Split(url.Path)
	ext := filepath.Ext(filename)
	filename = strings.TrimSuffix(filename, ext)

	return &stemExt{
		stem: filename,
		ext:  ext,
	}
}

// SetExtFromContentTypeIfEmpty sets the ext field based on the provided content type.
func (cd *stemExt) SetExtFromContentTypeIfEmpty(contentType string) *stemExt {
	if len(contentType) == 0 || len(cd.ext) > 0 {
		return cd
	}

	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) != 0 {
		cd.ext = exts[0]
	}

	return cd
}

// OverrideExt sets the ext field if the provided ext is not empty.
func (cd *stemExt) OverrideExt(ext string) *stemExt {
	if len(ext) > 0 {
		cd.ext = ext
	}

	return cd
}

// OverrideStem sets the stem field if the provided stem is not empty.
func (cd *stemExt) OverrideStem(stem string) *stemExt {
	if len(stem) > 0 {
		cd.stem = stem
	}

	return cd
}

// StemExtWithFallback returns stem and ext, but if stem is empty, it uses a fallback value.
func (cd *stemExt) StemExtWithFallback() (string, string) {
	if len(cd.stem) == 0 {
		cd.stem = fallbackStem
	}

	return cd.stem, cd.ext
}

// StemExt returns the tuple of stem and ext.
func (cd *stemExt) StemExt() (string, string) {
	return cd.stem, cd.ext
}
