package httpheaders

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

const (
	// fallbackStem is used when the stem cannot be determined from the URL.
	fallbackStem = "image"

	// Content-Disposition header format
	contentDispositionsHeader = "%s; filename=\"%s%s\""

	// "inline" disposition types
	inlineDisposition = "inline"

	// "attachment" disposition type
	attachmentDisposition = "attachment"
)

// ContentDispositionValue generates the content-disposition header value.
//
// It uses the following priorities:
// 1. By default, it uses the filename and extension from the URL.
// 2. If `filename` is provided, it overrides the URL filename.
// 3. If `contentType` is provided, it tries to determine the extension from the content type.
// 4. If `ext` is provided, it overrides any extension determined from the URL or header.
// 5. If the filename is still empty, it uses fallback stem.
func ContentDispositionValue(url, filename, ext, contentType string, returnAttachment bool) string {
	// By default, let's use the URL filename and extension
	_, urlFilename := filepath.Split(url)
	urlExt := filepath.Ext(urlFilename)

	var rStem string

	// Avoid strings.TrimSuffix allocation by using slice operation
	if urlExt != "" {
		rStem = urlFilename[:len(urlFilename)-len(urlExt)]
	} else {
		rStem = urlFilename
	}

	var rExt = urlExt

	// If filename is provided explicitly, use it
	if len(filename) > 0 {
		rStem = filename
	}

	// If ext is provided explicitly, use it
	if len(ext) > 0 {
		rExt = ext
	} else if len(contentType) > 0 && rExt == "" {
		exts, err := mime.ExtensionsByType(contentType)
		if err == nil && len(exts) != 0 {
			rExt = exts[0]
		}
	}

	// If fallback is requested, and filename is still empty, override it with fallbackStem
	if len(rStem) == 0 {
		rStem = fallbackStem
	}

	disposition := inlineDisposition

	// Create the content-disposition header value
	if returnAttachment {
		disposition = attachmentDisposition
	}

	return fmt.Sprintf(contentDispositionsHeader, disposition, strings.ReplaceAll(rStem, `"`, "%22"), rExt)
}
