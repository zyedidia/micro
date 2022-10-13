package config

import "testing"

func Test_validateNonNegativeIntSlice(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"Positive value", []int{1}, false},
		{"Positive values", []int{1, 2}, false},
		{"No value", []int{}, false},
		// Errors
		{"Negative value", []int{-1}, true},
		{"Negative values", []int{-1, -2}, true},
		{"Mixed values", []int{1, -2}, true},
		{"Nil", nil, true},
		{"Incorrect type", []float64{123}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateNonNegativeIntSlice("colorcolumns", tt.value); (err != nil) != tt.wantErr {
				t.Errorf("validateNonNegativeIntSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
