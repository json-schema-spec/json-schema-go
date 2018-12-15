package jsonschema

import (
	"math"
)

// DefaultEpsilon determines the tolerance for error in floating point comparisons. This value is always used in a
const DefaultEpsilon float64 = 1e-3

type Validator struct {
	schema  Schema
	Epsilon float64
}

func NewValidator(schema Schema) Validator {
	return Validator{
		schema:  schema,
		Epsilon: DefaultEpsilon,
	}
}

func (v Validator) IsValid(data interface{}) bool {
	if v.schema.IsTrivial {
		return v.schema.TrivialValue
	}

	document := v.schema.Document

	if document.Minimum != nil {
		if num, ok := data.(float64); ok {
			if num < *document.Minimum {
				return false
			}
		}
	}

	if document.ExclusiveMinimum != nil {
		if num, ok := data.(float64); ok {
			if num <= *document.ExclusiveMinimum {
				return false
			}
		}
	}

	if document.Maximum != nil {
		if num, ok := data.(float64); ok {
			if num > *document.Maximum {
				return false
			}
		}
	}

	if document.ExclusiveMaximum != nil {
		if num, ok := data.(float64); ok {
			if num >= *document.ExclusiveMaximum {
				return false
			}
		}
	}

	if document.MultipleOf != nil {
		if num, ok := data.(float64); ok {
			mod := math.Mod(math.Abs(num), *document.MultipleOf) / *document.MultipleOf

			if mod > v.Epsilon && mod < 1-v.Epsilon {
				return false
			}
		}
	}

	return true
}
