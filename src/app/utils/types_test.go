package utils

import (
	"encoding/json"
	"testing"
)

func TestRelationship_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		rel      *Relationship[int]
		expected string
	}{
		{
			name:     "Nil Relationship",
			rel:      nil,
			expected: "null",
		},
		{
			name:     "Unloaded Relationship",
			rel:      &Relationship[int]{Loaded: false, Value: 0},
			expected: "null",
		},
		{
			name:     "Loaded Relationship with int value",
			rel:      &Relationship[int]{Loaded: true, Value: 42},
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.rel)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRelationship_MarshalJSON_String(t *testing.T) {
	tests := []struct {
		name     string
		rel      *Relationship[string]
		expected string
	}{
		{
			name:     "Nil Relationship",
			rel:      nil,
			expected: "null",
		},
		{
			name:     "Unloaded Relationship",
			rel:      &Relationship[string]{Loaded: false, Value: "unloaded"},
			expected: "null",
		},
		{
			name:     "Loaded Relationship with string value",
			rel:      &Relationship[string]{Loaded: true, Value: "test"},
			expected: `"test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.rel)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestRelationship_MarshalJSON_Struct(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	tests := []struct {
		name     string
		rel      *Relationship[TestStruct]
		expected string
	}{
		{
			name:     "Nil Relationship",
			rel:      nil,
			expected: "null",
		},
		{
			name:     "Unloaded Relationship",
			rel:      &Relationship[TestStruct]{Loaded: false, Value: TestStruct{Name: "unloaded"}},
			expected: "null",
		},
		{
			name:     "Loaded Relationship with struct value",
			rel:      &Relationship[TestStruct]{Loaded: true, Value: TestStruct{Name: "test"}},
			expected: `{"Name":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.rel)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON() = %s, want %s", result, tt.expected)
			}
		})
	}
}
