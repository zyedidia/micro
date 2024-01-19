package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringWidth(t *testing.T) {
	bytes := []byte("\tPot să \tmănânc sticlă și ea nu mă rănește.")

	n := StringWidth(bytes, 23, 4)
	assert.Equal(t, 26, n)
}

func TestSliceVisualEnd(t *testing.T) {
	s := []byte("\thello")
	slc, n, _ := SliceVisualEnd(s, 2, 4)
	assert.Equal(t, []byte("\thello"), slc)
	assert.Equal(t, 2, n)

	slc, n, _ = SliceVisualEnd(s, 1, 4)
	assert.Equal(t, []byte("\thello"), slc)
	assert.Equal(t, 1, n)

	slc, n, _ = SliceVisualEnd(s, 4, 4)
	assert.Equal(t, []byte("hello"), slc)
	assert.Equal(t, 0, n)

	slc, n, _ = SliceVisualEnd(s, 5, 4)
	assert.Equal(t, []byte("ello"), slc)
	assert.Equal(t, 0, n)
}

func TestIntSliceOpt(t *testing.T) {
	tests := []struct {
		name    string
		arg     interface{}
		want    []int
		wantErr bool
	}{
		{"Single value", "1", []int{1}, false},
		{"Space-separated", "1 2", []int{1, 2}, false},
		{"Comma-separated", "1,2", []int{1, 2}, false},
		{"Array syntax", "[1, 2]", []int{1, 2}, false},
		{"Negative value", "-1", []int{-1}, false},
		{"Space-separated negatives", "-1 -2", []int{-1, -2}, false},
		{"Comma-separated negatives", "-1,-2", []int{-1, -2}, false},
		{"Array syntax negatives", "[-1, -2]", []int{-1, -2}, false},
		{"Int slice value", []int{1}, []int{1}, false},
		{"Int slice values", []int{1, 2}, []int{1, 2}, false},
		{"Random", "1n2g124n1g-23j-12n3", []int{1, 2, 124, 1, -23, -12, 3}, false},
		// Errors
		{"Nil", nil, nil, true},
		{"No values", "", nil, true},
		{"No int slice values", []int{}, nil, true},
		{"No valid values", "abc -a ~!@#$%^&*()_+{}:\"<>?`-=[];',./", nil, true},
		{"Invalid value", "190-23", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IntSliceOpt(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntSliceOpt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IntSliceOpt() = %v, want %v", got, tt.want)
			}
		})
	}
}
