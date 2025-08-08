package imagedetect

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/imagetype_new"
)

// Detect attempts to detect the image type from a reader.
// It first tries magic byte detection, then custom detectors in registration order
func Detect(r io.Reader) (imagetype_new.Type, error) {
	// Start with 64 bytes to cover magic bytes
	buf := make([]byte, 64)

	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return imagetype_new.Unknown, err
	}

	data := buf[:n]

	// First try magic byte detection
	for _, magic := range registry.magicBytes {
		if hasMagicBytes(data, magic) {
			return magic.Type, nil
		}
	}

	// Then try custom detectors
	for _, detector := range registry.detectors {
		// Check if we have enough bytes for this detector
		if len(data) < detector.BytesNeeded {
			// Need to read more data
			additionalBytes := detector.BytesNeeded - len(data)
			extraBuf := make([]byte, additionalBytes)
			extraN, err := io.ReadFull(r, extraBuf)

			// It's fine if we can't read required number of bytes
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				return imagetype_new.Unknown, err
			}

			// Extend our data buffer
			data = append(data, extraBuf[:extraN]...)
		}

		if typ, err := detector.Func(data); err == nil && typ != imagetype_new.Unknown {
			return typ, nil
		}
	}

	return imagetype_new.Unknown, newUnknownFormatError()
}

// hasMagicBytes checks if the data matches a magic byte signature
// Supports '?' characters in signature which match any byte
func hasMagicBytes(data []byte, magic MagicBytes) bool {
	if len(data) < len(magic.Signature) {
		return false
	}

	for i, c := range magic.Signature {
		if c != data[i] && c != '?' {
			return false
		}
	}
	return true
}

// init registers default magic bytes for common image formats
func init() {
	// JPEG magic bytes
	RegisterMagicBytes([]byte("\xff\xd8"), imagetype_new.JPEG)

	// JXL magic bytes
	//
	// NOTE: for "naked" jxl (0xff 0x0a) there is no way to ensure this is a JXL file, except to fully
	// decode it. The data starts right after it, no additional marker bytes are provided.
	// We stuck with the potential false positives here.
	RegisterMagicBytes([]byte{0xff, 0x0a}, imagetype_new.JXL)                                                             // JXL codestream (can't use string due to 0x0a)
	RegisterMagicBytes([]byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x4C, 0x20, 0x0D, 0x0A, 0x87, 0x0A}, imagetype_new.JXL) // JXL container (has null bytes)

	// PNG magic bytes
	RegisterMagicBytes([]byte("\x89PNG\r\n\x1a\n"), imagetype_new.PNG)

	// WEBP magic bytes (RIFF container with WEBP fourcc) - using wildcard for size
	RegisterMagicBytes([]byte("RIFF????WEBP"), imagetype_new.WEBP)

	// GIF magic bytes
	RegisterMagicBytes([]byte("GIF8?a"), imagetype_new.GIF)

	// ICO magic bytes
	RegisterMagicBytes([]byte{0, 0, 1, 0}, imagetype_new.ICO) // ICO (has null bytes)

	// HEIC/HEIF magic bytes with wildcards for size
	RegisterMagicBytes([]byte("????ftypheic"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftypheix"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftyphevc"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftypheim"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftypheis"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftyphevm"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftyphevs"), imagetype_new.HEIC)
	RegisterMagicBytes([]byte("????ftypmif1"), imagetype_new.HEIC)

	// AVIF magic bytes
	RegisterMagicBytes([]byte("????ftypavif"), imagetype_new.AVIF)

	// BMP magic bytes
	RegisterMagicBytes([]byte("BM"), imagetype_new.BMP)

	// TIFF magic bytes (little-endian and big-endian)
	RegisterMagicBytes([]byte("II*\x00"), imagetype_new.TIFF) // Little-endian
	RegisterMagicBytes([]byte("MM\x00*"), imagetype_new.TIFF) // Big-endian
}
