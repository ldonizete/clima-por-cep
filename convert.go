package main

import "math"

// CelsiusToFahrenheit applies F = C × 1.8 + 32, rounded to two decimals.
func CelsiusToFahrenheit(c float64) float64 {
	return round2(c*1.8 + 32)
}

// CelsiusToKelvin applies K = C + 273, as the challenge statement defines it.
func CelsiusToKelvin(c float64) float64 {
	return round2(c + 273)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
