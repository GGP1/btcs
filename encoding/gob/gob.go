// Package gob provides the functions necessary for encoding and decoding Go structs
// using encoding/gob standard library package.
package gob

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// Encode serializes an object.
func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, fmt.Errorf("encoding %T: %w", v, err)
	}

	return buf.Bytes(), nil
}

// Decode deserializes data into an object.
func Decode[T any](data []byte) (T, error) {
	var value T

	reader := bytes.NewBuffer(data)
	if err := gob.NewDecoder(reader).Decode(&value); err != nil {
		return value, fmt.Errorf("decoding %T: %w", value, err)
	}

	return value, nil
}
