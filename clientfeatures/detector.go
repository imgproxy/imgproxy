package clientfeatures

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Maximum supported DPR from client hints
const maxClientHintDPR = 8

// Detector detects client features from request headers
type Detector struct {
	config *Config
	vary   string
}

// NewDetector creates a new Detector instance
func NewDetector(config *Config) *Detector {
	vary := make([]string, 0, 5)

	if config.AutoWebp || config.EnforceWebp ||
		config.AutoAvif || config.EnforceAvif ||
		config.AutoJxl || config.EnforceJxl {
		vary = append(vary, httpheaders.Accept)
	}

	if config.EnableClientHints {
		vary = append(
			vary,
			httpheaders.SecChDpr, httpheaders.Dpr, httpheaders.SecChWidth, httpheaders.Width,
		)
	}

	return &Detector{
		config: config,
		vary:   strings.Join(vary, ", "),
	}
}

// Features detects client features from HTTP headers
func (d *Detector) Features(header http.Header) Features {
	var f Features

	headerAccept := header.Get("Accept")

	if (d.config.AutoWebp || d.config.EnforceWebp) && strings.Contains(headerAccept, "image/webp") {
		f.PreferWebP = true
		f.EnforceWebP = d.config.EnforceWebp
	}

	if (d.config.AutoAvif || d.config.EnforceAvif) && strings.Contains(headerAccept, "image/avif") {
		f.PreferAvif = true
		f.EnforceAvif = d.config.EnforceAvif
	}

	if (d.config.AutoJxl || d.config.EnforceJxl) && strings.Contains(headerAccept, "image/jxl") {
		f.PreferJxl = true
		f.EnforceJxl = d.config.EnforceJxl
	}

	if !d.config.EnableClientHints {
		return f
	}
	for _, key := range []string{httpheaders.SecChDpr, httpheaders.Dpr} {
		val := header.Get(key)
		if len(val) == 0 {
			continue
		}

		if d, err := strconv.ParseFloat(val, 64); err == nil && (d > 0 && d <= maxClientHintDPR) {
			f.ClientHintsDPR = d
			break
		}
	}

	for _, key := range []string{httpheaders.SecChWidth, httpheaders.Width} {
		val := header.Get(key)
		if len(val) == 0 {
			continue
		}

		if w, err := strconv.Atoi(val); err == nil && w > 0 {
			f.ClientHintsWidth = w
			break
		}
	}

	return f
}

// SetVary sets the Vary header value based on enabled features
func (d *Detector) SetVary(header http.Header) {
	if len(d.vary) > 0 {
		header.Set(httpheaders.Vary, d.vary)
	}
}
