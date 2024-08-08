package utils

import (
	"encoding/json"
)

type Relationship[T any] struct {
	Loaded bool
	Value  T
}

// Marshal the Relationship's value to JSON
func (e *Relationship[T]) MarshalJSON() ([]byte, error) {
	if e == nil || !e.Loaded {
		return []byte("null"), nil
	}

	return json.Marshal(e.Value)
}
