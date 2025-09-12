package options

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imath"
)

// ParsePath parses the given request path and returns the processing options and image URL
func (f *Factory) ParsePath(
	path string,
	headers http.Header,
) (po *ProcessingOptions, imageURL string, err error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if f.config.OnlyPresets {
		po, imageURL, err = f.parsePathPresets(parts, headers)
	} else {
		po, imageURL, err = f.parsePathOptions(parts, headers)
	}

	if err != nil {
		return nil, "", ierrors.Wrap(err, 0)
	}

	return po, imageURL, nil
}

// parsePathOptions parses processing options from the URL path
func (f *Factory) parsePathOptions(parts []string, headers http.Header) (*ProcessingOptions, string, error) {
	if _, ok := resizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	po, err := f.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	options, urlParts := f.parseURLOptions(parts)

	if err = f.applyURLOptions(po, options, false); err != nil {
		return nil, "", err
	}

	url, extension, err := f.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !po.Raw && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}

// parsePathPresets parses presets from the URL path
func (f *Factory) parsePathPresets(parts []string, headers http.Header) (*ProcessingOptions, string, error) {
	po, err := f.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], f.config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = f.applyPresetOption(po, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := f.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !po.Raw && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}

func (f *Factory) defaultProcessingOptions(headers http.Header) (*ProcessingOptions, error) {
	po := f.New()

	headerAccept := headers.Get("Accept")

	if strings.Contains(headerAccept, "image/webp") {
		po.PreferWebP = f.config.AutoWebp || f.config.EnforceWebp
		po.EnforceWebP = f.config.EnforceWebp
	}

	if strings.Contains(headerAccept, "image/avif") {
		po.PreferAvif = f.config.AutoAvif || f.config.EnforceAvif
		po.EnforceAvif = f.config.EnforceAvif
	}

	if strings.Contains(headerAccept, "image/jxl") {
		po.PreferJxl = f.config.AutoJxl || f.config.EnforceJxl
		po.EnforceJxl = f.config.EnforceJxl
	}

	if f.config.EnableClientHints {
		headerDPR := headers.Get("Sec-CH-DPR")
		if len(headerDPR) == 0 {
			headerDPR = headers.Get("DPR")
		}
		if len(headerDPR) > 0 {
			if dpr, err := strconv.ParseFloat(headerDPR, 64); err == nil && (dpr > 0 && dpr <= maxClientHintDPR) {
				po.Dpr = dpr
			}
		}

		headerWidth := headers.Get("Sec-CH-Width")
		if len(headerWidth) == 0 {
			headerWidth = headers.Get("Width")
		}
		if len(headerWidth) > 0 {
			if w, err := strconv.Atoi(headerWidth); err == nil {
				po.Width = imath.Shrink(w, po.Dpr)
			}
		}
	}

	if _, ok := f.presets["default"]; ok {
		if err := f.applyPresetOption(po, []string{"default"}); err != nil {
			return po, err
		}
	}

	return po, nil
}
