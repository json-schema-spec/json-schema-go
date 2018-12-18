package jsonschema

import (
	"math"
)

type Validator struct {
	schema Schema
}

func NewValidator(schema Schema) Validator {
	return Validator{
		schema: schema,
	}
}

func (v Validator) IsValid(instance interface{}) bool {
	s := v.schema

	switch i := instance.(type) {
	case bool:
		return !s.Bool.Reject
	case float64:
		if s.Number.Reject {
			return false
		}

		if s.Number.Integer && math.Mod(i, 1.0) != 0.0 {
			return false
		}

		return true
	case string:
		return !s.String.Reject
	case []interface{}:
		return !s.Array.Reject
	case map[string]interface{}:
		return !s.Object.Reject
	case nil:
		return !s.Null.Reject
	default:
		return false
	}
}
