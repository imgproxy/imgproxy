package imagetype

import (
	"github.com/imgproxy/imgproxy/v3/bufreader"
)

var (
	tiffLE = []byte("II*\x00")
	tiffBE = []byte("MM\x00*")
)

func init() {
	// Register TIFF detector.
	// We register it with a priority of 80 to run it right before SVG,
	// but after other magic number detectors.
	RegisterDetector(80, IsTIFF)
}

// IsTIFF detects if the image is a TIFF
func IsTIFF(r bufreader.ReadPeeker, ct, ext string) (Type, error) {
	b, err := r.Peek(max(len(tiffLE), len(tiffBE)))
	if err != nil {
		return Unknown, err
	}

	// If the file is detected as TIFF, but has a RAW extension, we skip it
	// since it is false positive.
	if (hasMagicBytes(b, tiffLE) || hasMagicBytes(b, tiffBE)) && !IsRawExtOrMime(ct, ext) {
		return TIFF, nil
	}

	return Unknown, nil
}
