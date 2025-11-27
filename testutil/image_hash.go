package testutil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/corona10/goimagehash"
)

// ImageHashType defines the type of hash algorithm to use
type ImageHashType byte

const (
	HashTypeDifference ImageHashType = iota // dHash
	HashTypePerception                      // pHash
	HashTypeSHA256                          // SHA256 of pixel data
)

// ImageHash wraps different hash types with a unified interface
type ImageHash struct {
	hashType   ImageHashType
	imageHash  *goimagehash.ImageHash // for dHash/pHash
	sha256Hash [32]byte               // for SHA256
}

// NewImageHash creates a new hash from an image
func NewImageHash(img *image.RGBA, hashType ImageHashType) (*ImageHash, error) {
	h := &ImageHash{hashType: hashType}

	switch hashType {
	case HashTypeDifference:
		hash, err := goimagehash.DifferenceHash(img)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate difference hash: %w", err)
		}
		h.imageHash = hash

	case HashTypePerception:
		hash, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate perception hash: %w", err)
		}
		h.imageHash = hash

	case HashTypeSHA256:
		h.sha256Hash = sha256.Sum256(img.Pix)

	default:
		return nil, fmt.Errorf("unsupported hash type: %d", hashType)
	}

	return h, nil
}

// NewImageHashFromReader loads an image from a reader and calculates its hash
func NewImageHashFromReader(r io.Reader, hashType ImageHashType) (*ImageHash, error) {
	img, err := LoadImage(r)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	return NewImageHash(img, hashType)
}

// NewImageHashFromPath loads an image from a file path and calculates its hash
func NewImageHashFromPath(path string, hashType ImageHashType) (*ImageHash, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return NewImageHashFromReader(file, hashType)
}

// Distance calculates the distance between two hashes
// Returns error if hash types don't match
func (h *ImageHash) Distance(other *ImageHash) (int, error) {
	if h.hashType != other.hashType {
		return 0, fmt.Errorf("cannot compare hash type %d with %d", h.hashType, other.hashType)
	}

	switch h.hashType {
	case HashTypeDifference, HashTypePerception:
		return h.imageHash.Distance(other.imageHash)

	case HashTypeSHA256:
		if h.sha256Hash == other.sha256Hash {
			return 0, nil
		}
		return 1, nil

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
	case HashTypeDifference, HashTypePerception:
		if err := h.imageHash.Dump(w); err != nil {
			return fmt.Errorf("failed to dump perceptual hash: %w", err)
		}

	case HashTypeSHA256:
		if _, err := w.Write(h.sha256Hash[:]); err != nil {
			return fmt.Errorf("failed to write SHA256 hash: %w", err)
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
	case HashTypeDifference, HashTypePerception:
		hash, err := goimagehash.LoadImageHash(r)
		if err != nil {
			return nil, fmt.Errorf("failed to load perceptual hash: %w", err)
		}
		h.imageHash = hash

	case HashTypeSHA256:
		if _, err := io.ReadFull(r, h.sha256Hash[:]); err != nil {
			return nil, fmt.Errorf("failed to read SHA256 hash: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported hash type: %d", h.hashType)
	}

	return h, nil
}

// String returns a string representation of the hash
func (h *ImageHash) String() string {
	switch h.hashType {
	case HashTypeDifference:
		return fmt.Sprintf("dHash:%s", h.imageHash.ToString())
	case HashTypePerception:
		return fmt.Sprintf("pHash:%s", h.imageHash.ToString())
	case HashTypeSHA256:
		return fmt.Sprintf("SHA256:%x", h.sha256Hash)
	default:
		return fmt.Sprintf("unknown(%d)", h.hashType)
	}
}
