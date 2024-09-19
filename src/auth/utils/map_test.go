package utils

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		mapFunc  func(int) int
		expected []int
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			mapFunc:  func(x int) int { return x * x },
			expected: []int{},
		},
		{
			name:     "Square of integers",
			input:    []int{1, 2, 3, 4},
			mapFunc:  func(x int) int { return x * x },
			expected: []int{1, 4, 9, 16},
		},
		{
			name:     "Double the integers",
			input:    []int{1, 2, 3, 4},
			mapFunc:  func(x int) int { return x * 2 },
			expected: []int{2, 4, 6, 8},
		},
		{
			name:     "Map to constant value",
			input:    []int{1, 2, 3, 4},
			mapFunc:  func(x int) int { return 10 },
			expected: []int{10, 10, 10, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.input, tt.mapFunc)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Map() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapStringToLength(t *testing.T) {
	input := []string{"a", "ab", "abc"}
	expected := []int{1, 2, 3}
	result := Map(input, func(s string) int { return len(s) })

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Map() = %v, want %v", result, expected)
	}
}
