package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

var (
	ErrNotValType = errors.New("not of value type")
)

// ParseJSON parses a json string and construct it into a Go object of type defined by (generic) type parameter T.
// T must be a pointer type, otherwise the function returns an error ErrNotValType.
func ParseJSON[T any](reader io.Reader) (T, error) {
	var target T
	kind := reflect.ValueOf(target).Kind()
	if kind != reflect.Struct && kind != reflect.Pointer && kind != reflect.Map {
		return *new(T), fmt.Errorf("the type parameter T must be of struct type: %w", ErrNotValType)
	}

	err := json.NewDecoder(reader).Decode(&target)
	if err != nil && !errors.Is(err, io.EOF) {
		return *new(T), fmt.Errorf("failed to unmarhsal json to %T when parsing a json object: %w", target, err)
	}

	return target, nil
}
