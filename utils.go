package main

import (
	"math"
	"strings"
	"unsafe"
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minNonZeroInt(a, b int) int {
	switch {
	case a == 0:
		return b
	case b == 0:
		return a
	}

	return minInt(a, b)
}

func roundToInt(a float64) int {
	return int(math.Round(a))
}

func scaleInt(a int, scale float64) int {
	if a == 0 {
		return 0
	}

	return roundToInt(float64(a) * scale)
}

func trimAfter(s string, sep byte) string {
	i := strings.IndexByte(s, sep)
	if i < 0 {
		return s
	}
	return s[:i]
}

func ptrToBytes(ptr unsafe.Pointer, size int) []byte {
	return (*[math.MaxInt32]byte)(ptr)[:int(size):int(size)]
}
