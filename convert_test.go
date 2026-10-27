package main

import "testing"

func TestConversions(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		wantF   float64
		wantK   float64
	}{
		{"example from the spec", 28.5, 83.3, 301.5},
		{"freezing point", 0, 32, 273},
		{"boiling point", 100, 212, 373},
		{"negative", -40, -40, 233},
		{"fractional", 21.7, 71.06, 294.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CelsiusToFahrenheit(tt.celsius); got != tt.wantF {
				t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", tt.celsius, got, tt.wantF)
			}
			if got := CelsiusToKelvin(tt.celsius); got != tt.wantK {
				t.Errorf("CelsiusToKelvin(%v) = %v, want %v", tt.celsius, got, tt.wantK)
			}
		})
	}
}
