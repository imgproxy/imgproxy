package imagetype

import (
	"strings"
)

var (
	// .RAW file extension list, they may mimic to .TIFF
	rawExtensions = map[string]bool{
		".3fr":    true,
		".adng":   true,
		".arw":    true,
		".cap":    true,
		".cr2":    true,
		".cr3":    true,
		".crw":    true,
		".data":   true,
		".dcr":    true,
		".dng":    true,
		".eip":    true,
		".erf":    true,
		".fff":    true,
		".gpr":    true,
		".iiq":    true,
		".k25":    true,
		".kdc":    true,
		".mef":    true,
		".moc":    true,
		".mos":    true,
		".mdc":    true,
		".mrw":    true,
		".nef":    true,
		".nrw":    true,
		".orf":    true,
		".ori":    true,
		".pef":    true,
		".ppm":    true,
		".proraw": true,
		".raf":    true,
		".raw":    true,
		".rw2":    true,
		".rwl":    true,
		".sr2":    true,
		".srf":    true,
		".srw":    true,
		".x3f":    true,
	}

	// RAW image MIME types (in case extension is missing)
	rawMimeTypes = map[string]bool{
		"image/x-hasselblad-3fr": true,
		"image/x-adobe-dng":      true,
		"image/x-sony-arw":       true,
		"image/x-phaseone-cap":   true,
		"image/x-canon-cr2":      true,
		"image/x-canon-cr3":      true,
		"image/x-canon-crw":      true,
		"image/x-kodak-dcr":      true,
		"image/x-epson-erf":      true,
		"image/x-hasselblad-fff": true,
		"image/x-gopro-gpr":      true,
		"image/x-phaseone-iiq":   true,
		"image/x-kodak-k25":      true,
		"image/x-kodak-kdc":      true,
		"image/x-mamiya-mef":     true,
		"image/x-leaf-mos":       true,
		"image/x-minolta-mrw":    true,
		"image/x-nikon-nef":      true,
		"image/x-nikon-nrw":      true,
		"image/x-olympus-orf":    true,
		"image/x-sony-ori":       true,
		"image/x-pentax-pef":     true,
		"image/x-apple-proraw":   true,
		"image/x-fuji-raf":       true,
		"image/x-raw":            true,
		"image/x-panasonic-rw2":  true,
		"image/x-leica-rwl":      true,
		"image/x-sony-sr2":       true,
		"image/x-sony-srf":       true,
		"image/x-samsung-srw":    true,
		"image/x-sigma-x3f":      true,
	}
)

// IsRawExtOrMime checks if the given content type or extension belongs to a RAW image format
func IsRawExtOrMime(ct, ext string) bool {
	return rawExtensions[strings.ToLower(ext)] || rawMimeTypes[ct]
}
