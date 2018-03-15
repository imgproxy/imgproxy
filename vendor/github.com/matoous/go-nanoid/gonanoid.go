package gonanoid

import (
	"crypto/rand"
	"math"
)

// DefaultsType is the type of the default configuration for Nanoid
type DefaultsType struct {
	Alphabet string
	Size     int
	MaskSize int
}

// GetDefaults returns the default configuration for Nanoid
func GetDefaults() *DefaultsType {
	return &DefaultsType{
		Alphabet: "_~0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", // len=64
		Size:     22,
		MaskSize: 5,
	}
}

var defaults = GetDefaults()

func initMasks(params ...int) []uint {
	var size int
	if len(params) == 0 {
		size = defaults.MaskSize
	} else {
		size = params[0]
	}
	masks := make([]uint, size)
	for i := 0; i < size; i++ {
		shift := 3 + i
		masks[i] = (2 << uint(shift)) - 1
	}
	return masks
}

func getMask(alphabet string, masks []uint) int {
	for i := 0; i < len(masks); i++ {
		curr := int(masks[i])
		if curr >= len(alphabet)-1 {
			return curr
		}
	}
	return 0
}

// Random generates cryptographically strong pseudo-random data.
// The size argument is a number indicating the number of bytes to generate.
func Random(size int) ([]byte, error) {
	var randomBytes = make([]byte, size)
	_, err := rand.Read(randomBytes)
	return randomBytes, err
}

// Generate is a low-level function to change alphabet and ID size.
func Generate(alphabet string, size int) (string, error) {
	masks := initMasks(size)
	mask := getMask(alphabet, masks)
	ceilArg := 1.6 * float64(mask*size) / float64(len(alphabet))
	step := int(math.Ceil(ceilArg))

	id := make([]byte, size)
	bytes := make([]byte, step)
	for j := 0; ; {
		_, err := rand.Read(bytes)
		if err != nil {
			return "", err
		}
		for i := 0; i < step; i++ {
			currByte := bytes[i] & byte(mask)
			if currByte < byte(len(alphabet)) {
				id[j] = alphabet[currByte]
				j++
				if j == size {
					return string(id[:size]), nil
				}
			}
		}
	}
}

// Nanoid generates secure URL-friendly unique ID.
func Nanoid(param ...int) (string, error) {
	var size int
	if len(param) == 0 {
		size = defaults.Size
	} else {
		size = param[0]
	}
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	id := make([]byte, size)
	for i := 0; i < size; i++ {
		id[i] = defaults.Alphabet[bytes[i]&63]
	}
	return string(id[:size]), nil
}
