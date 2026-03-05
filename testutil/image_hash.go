package testutil

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#cgo LDFLAGS: -lm
#include "dct2.h"
*/
import "C"
import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"unsafe"

	"github.com/corona10/goimagehash"
)

// ImageHashType defines the type of hash algorithm to use
type ImageHashType byte

const (
	HashTypeSHA256 = iota // SHA256 of pixel data
	HashTypeDct           // DCT-based hash
)

// ImageHash wraps different hash types with a unified interface
type ImageHash struct {
	hashType   ImageHashType
	imageHash  *goimagehash.ImageHash // for dHash/pHash
	sha256Hash [32]byte               // for SHA256
	dctHash    []float32              // for DCT hash
}

// NewImageHash creates a new hash from an image reader
func NewImageHash(r io.Reader, hashType ImageHashType) (*ImageHash, error) {
	h := &ImageHash{hashType: hashType}

	switch hashType {
	case HashTypeDct:
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read image data: %w", err)
		}

		bufPtr := unsafe.Pointer(unsafe.SliceData(buf))
		var dctArray *C.float
		var length C.size_t

		//nolint:gocritic
		result := C.vips_dct2_hash(bufPtr, C.size_t(len(buf)), &dctArray, &length)
		if dctArray != nil {
			defer C.free(unsafe.Pointer(dctArray))
		}

		if result != 0 {
			return nil, fmt.Errorf("failed to calculate DCT hash")
		}

		dctHash := make([]float32, int(length))

		// Convert C array to Go slice safely
		cSlice := unsafe.Slice(dctArray, int(length))
		for i, v := range cSlice {
			dctHash[i] = float32(v)
		}

		h.dctHash = dctHash

	case HashTypeSHA256:
		img, err := LoadImage(r)
		if err != nil {
			return nil, fmt.Errorf("failed to load image: %w", err)
		}

		h.sha256Hash = sha256.Sum256(img.Pix)

	default:
		return nil, fmt.Errorf("unsupported hash type: %d", hashType)
	}

	return h, nil
}

// NewImageHashFromPath loads an image from a file path and calculates its hash
func NewImageHashFromPath(path string, hashType ImageHashType) (*ImageHash, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return NewImageHash(file, hashType)
}

// Distance calculates the distance between two hashes
// Returns error if hash types don't match
func (h *ImageHash) Distance(other *ImageHash) (float32, error) {
	if h.hashType != other.hashType {
		return 0, fmt.Errorf("cannot compare hash type %d with %d", h.hashType, other.hashType)
	}

	switch h.hashType {
	case HashTypeSHA256:
		if h.sha256Hash == other.sha256Hash {
			return 0, nil
		}
		return 1, nil

	case HashTypeDct:
		if len(h.dctHash) != len(other.dctHash) {
			return math.MaxFloat32, nil
		}

		var sumSquaredDiff float32
		for i := range h.dctHash {
			diff := h.dctHash[i] - other.dctHash[i]
			sumSquaredDiff += diff * diff
		}
		mse := sumSquaredDiff / float32(len(h.dctHash))

		return mse, nil

	default:
		return 0, fmt.Errorf("unsupported hash type: %d", h.hashType)
	}
}

// Dump writes the hash to a writer
// Format: [1 byte: type][variable: hash data]
func (h *ImageHash) Dump(w io.Writer) error {
	// Write type byte
	if err := binary.Write(w, binary.LittleEndian, h.hashType); err != nil {
		return fmt.Errorf("failed to write hash type: %w", err)
	}

	switch h.hashType {
	case HashTypeSHA256:
		if _, err := w.Write(h.sha256Hash[:]); err != nil {
			return fmt.Errorf("failed to write SHA256 hash: %w", err)
		}

	case HashTypeDct:
		// Write length
		if err := binary.Write(w, binary.LittleEndian, uint16(len(h.dctHash))); err != nil {
			return fmt.Errorf("failed to write DCT hash length: %w", err)
		}
		// Write array
		for _, v := range h.dctHash {
			if err := binary.Write(w, binary.LittleEndian, v); err != nil {
				return fmt.Errorf("failed to write DCT hash value: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported hash type: %d", h.hashType)
	}

	return nil
}

// LoadImageHash loads a hash from a reader
func LoadImageHash(r io.Reader) (*ImageHash, error) {
	h := &ImageHash{}

	// Read type byte
	if err := binary.Read(r, binary.LittleEndian, &h.hashType); err != nil {
		return nil, fmt.Errorf("failed to read hash type: %w", err)
	}

	switch h.hashType {
	case HashTypeSHA256:
		if _, err := io.ReadFull(r, h.sha256Hash[:]); err != nil {
			return nil, fmt.Errorf("failed to read SHA256 hash: %w", err)
		}

	case HashTypeDct:
		var length uint16
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return nil, fmt.Errorf("failed to read DCT hash length: %w", err)
		}
		h.dctHash = make([]float32, length)
		for i := range h.dctHash {
			if err := binary.Read(r, binary.LittleEndian, &h.dctHash[i]); err != nil {
				return nil, fmt.Errorf("failed to read DCT hash value: %w", err)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported hash type: %d", h.hashType)
	}

	return h, nil
}

// String returns a string representation of the hash
func (h *ImageHash) String() string {
	switch h.hashType {
	case HashTypeSHA256:
		return fmt.Sprintf("SHA256:%x", h.sha256Hash)
	case HashTypeDct:
		return fmt.Sprintf("DctHash:%v", h.dctHash)
	default:
		return fmt.Sprintf("unknown(%d)", h.hashType)
	}
}
