package main

import "math"

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

func roundToInt(a float64) int {
	return int(math.Round(a))
}
