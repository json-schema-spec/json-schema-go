package jsonschema

import (
	"math"

	"github.com/ucarion/json-pointer"
)

type Validator struct {
	schema Schema
}

type ValidationResult struct {
	Errors []ValidationError
}

type ValidationError struct {
	InstancePath jsonpointer.Ptr
	SchemaPath   jsonpointer.Ptr
}

func NewValidator(s Schema) Validator {
	return Validator{schema: s}
}

func (v *Validator) Validate(instance interface{}) (ValidationResult, error) {
	result := ValidationResult{
		Errors: []ValidationError{},
	}

	typeErr := ValidationError{
		InstancePath: jsonpointer.Ptr{Tokens: []string{}},
		SchemaPath:   jsonpointer.Ptr{Tokens: []string{"type"}},
	}

	switch val := instance.(type) {
	case nil:
		if v.schema.Type != "null" {
			result.Errors = append(result.Errors, typeErr)
		}
	case bool:
		if v.schema.Type != "boolean" {
			result.Errors = append(result.Errors, typeErr)
		}
	case float64:
		if v.schema.Type == "integer" {
			if val != math.Trunc(val) {
				result.Errors = append(result.Errors, typeErr)
			}
		} else if v.schema.Type != "number" {
			result.Errors = append(result.Errors, typeErr)
		}
	case string:
		if v.schema.Type != "string" {
			result.Errors = append(result.Errors, typeErr)
		}
	case []interface{}:
		if v.schema.Type != "array" {
			result.Errors = append(result.Errors, typeErr)
		}
	case map[string]interface{}:
		if v.schema.Type != "object" {
			result.Errors = append(result.Errors, typeErr)
		}
	default:
		// todo errors
		panic("bad type")
	}

	return result, nil
}
