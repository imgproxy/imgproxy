package imath

import "math"

func MinNonZero(a, b int) int {
	switch {
	case a == 0:
		return b
	case b == 0:
		return a
	}

	return min(a, b)
}

func Round(a float64) int {
	return int(math.Round(a))
}

func RoundToEven(a float64) int {
	return int(math.RoundToEven(a))
}

func Scale(a int, scale float64) int {
	if a == 0 {
		return 0
	}

	return Round(float64(a) * scale)
}

func ScaleToEven(a int, scale float64) int {
	if a == 0 {
		return 0
	}

	return RoundToEven(float64(a) * scale)
}

func Shrink(a int, shrink float64) int {
	if a == 0 {
		return 0
	}

	return Round(float64(a) / shrink)
}

func ShrinkToEven(a int, shrink float64) int {
	if a == 0 {
		return 0
	}

	return RoundToEven(float64(a) / shrink)
}
