package imagedetect

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

const (
	// maxDetectionLimit is maximum bytes detectors allowed to read from the source
	maxDetectionLimit = 32768
)

// Detect attempts to detect the image type from a reader.
// It first tries magic byte detection, then custom detectors in registration order
func Detect(r io.Reader) (imagetype.Type, error) {
	br := bufreader.New(io.LimitReader(r, maxDetectionLimit))

	for _, fn := range registry.detectors {
		br.Rewind()
		if typ, err := fn(br); err == nil && typ != imagetype.Unknown {
			return typ, nil
		}
	}

	return imagetype.Unknown, newUnknownFormatError()
}

// hasMagicBytes checks if the data matches a magic byte signature
// Supports '?' characters in signature which match any byte
func hasMagicBytes(data []byte, magic []byte) bool {
	if len(data) < len(magic) {
		return false
	}

	for i, c := range magic {
		if c != data[i] && c != '?' {
			return false
		}
	}
	return true
}

// init registers default magic bytes for common image formats
func init() {
	// JPEG magic bytes
	RegisterMagicBytes([]byte("\xff\xd8"), imagetype.JPEG)

	// JXL magic bytes
	//
	// NOTE: for "naked" jxl (0xff 0x0a) there is no way to ensure this is a JXL file, except to fully
	// decode it. The data starts right after it, no additional marker bytes are provided.
	// We stuck with the potential false positives here.
	RegisterMagicBytes([]byte{0xff, 0x0a}, imagetype.JXL)                                                             // JXL codestream (can't use string due to 0x0a)
	RegisterMagicBytes([]byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x4C, 0x20, 0x0D, 0x0A, 0x87, 0x0A}, imagetype.JXL) // JXL container (has null bytes)

	// PNG magic bytes
	RegisterMagicBytes([]byte("\x89PNG\r\n\x1a\n"), imagetype.PNG)

	// WEBP magic bytes (RIFF container with WEBP fourcc) - using wildcard for size
	RegisterMagicBytes([]byte("RIFF????WEBP"), imagetype.WEBP)

	// GIF magic bytes
	RegisterMagicBytes([]byte("GIF8?a"), imagetype.GIF)

	// ICO magic bytes
	RegisterMagicBytes([]byte{0, 0, 1, 0}, imagetype.ICO) // ICO (has null bytes)

	// HEIC/HEIF magic bytes with wildcards for size
	RegisterMagicBytes([]byte("????ftypheic"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftypheix"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftyphevc"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftypheim"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftypheis"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftyphevm"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftyphevs"), imagetype.HEIC)
	RegisterMagicBytes([]byte("????ftypmif1"), imagetype.HEIC)

	// AVIF magic bytes
	RegisterMagicBytes([]byte("????ftypavif"), imagetype.AVIF)

	// BMP magic bytes
	RegisterMagicBytes([]byte("BM"), imagetype.BMP)

	// TIFF magic bytes (little-endian and big-endian)
	RegisterMagicBytes([]byte("II*\x00"), imagetype.TIFF) // Little-endian
	RegisterMagicBytes([]byte("MM\x00*"), imagetype.TIFF) // Big-endian
}
