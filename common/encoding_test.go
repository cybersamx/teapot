package common

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
)

func TestParseJSON(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}

	t.Run("With non-struct type", func(t *testing.T) {
		in := "mike"
		_, err := ParseJSON[string](strings.NewReader(in))
		assert.Error(t, err)
	})

	t.Run("With struct value type", func(t *testing.T) {
		in := `{"name": "mike", "age": 25}`

		wantPerson := person{
			Name: "mike",
			Age:  25,
		}

		obj, err := ParseJSON[person](strings.NewReader(in))
		assert.NoError(t, err)
		assert.Equal(t, wantPerson, obj)
	})

	t.Run("With struct pointer type", func(t *testing.T) {
		in := `{"name": "mike", "age": 25}`

		wantPersonPtr := &person{
			Name: "mike",
			Age:  25,
		}

		objPtr, err := ParseJSON[*person](strings.NewReader(in))
		assert.NoError(t, err)
		assert.Equal(t, wantPersonPtr, objPtr)
	})

	t.Run("With map[string]any type", func(t *testing.T) {
		in := `{"name": "mike", "age": 25}`

		wantPersonMap := map[string]any{
			"name": "mike",
			"age":  25,
		}

		objMap, err := ParseJSON[map[string]any](strings.NewReader(in))
		assert.NoError(t, err)
		diff := pretty.Compare(wantPersonMap, objMap)
		assert.Empty(t, diff)
	})
}
