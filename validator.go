package jsonschema

import (
	"math"
	"reflect"
	"regexp"
	"unicode/utf8"
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
	return v.isValid(data, v.schema)
}

func (v Validator) isValid(data interface{}, schema Schema) bool {
	if schema.IsTrivial {
		return schema.TrivialValue
	}

	document := schema.Document

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

	if document.MaxLength != nil {
		if str, ok := data.(string); ok {
			if utf8.RuneCountInString(str) > *document.MaxLength {
				return false
			}
		}
	}

	if document.MinLength != nil {
		if str, ok := data.(string); ok {
			if utf8.RuneCountInString(str) < *document.MinLength {
				return false
			}
		}
	}

	if document.Pattern != nil {
		if str, ok := data.(string); ok {
			re, err := regexp.Compile(*document.Pattern)
			if err != nil {
				// TODO: Validate inputted patterns in advance, and error on validator
				// creation.
				panic(err)
			}

			if !re.MatchString(str) {
				return false
			}
		}
	}

	if document.Items != nil {
		if arr, ok := data.([]interface{}); ok {
			if document.Items.IsSingle {
				for _, val := range arr {
					if !v.isValid(val, document.Items.Single) {
						return false
					}
				}
			} else {
				numItems := len(arr)
				if numItems > len(document.Items.List) {
					numItems = len(document.Items.List)
				}

				for i, s := range document.Items.List[:numItems] {
					if !v.isValid(arr[i], s) {
						return false
					}
				}

				if document.AdditionalItems != nil {
					for _, val := range arr[numItems:] {
						if !v.isValid(val, *document.AdditionalItems) {
							return false
						}
					}
				}
			}
		}
	}

	if document.MaxItems != nil {
		if arr, ok := data.([]interface{}); ok {
			if len(arr) > *document.MaxItems {
				return false
			}
		}
	}

	if document.MinItems != nil {
		if arr, ok := data.([]interface{}); ok {
			if len(arr) < *document.MinItems {
				return false
			}
		}
	}

	if document.UniqueItems != nil {
		if arr, ok := data.([]interface{}); ok && len(arr) > 0 {
			for _, val := range arr[1:] {
				if reflect.DeepEqual(arr[0], val) {
					return false
				}
			}
		}
	}

	if document.Contains != nil {
		if arr, ok := data.([]interface{}); ok {
			// TODO: Early return.
			allFailed := true
			for _, val := range arr {
				if v.isValid(val, *document.Contains) {
					allFailed = false
				}
			}

			if allFailed {
				return false
			}
		}
	}

	if document.MaxProperties != nil {
		if obj, ok := data.(map[string]interface{}); ok {
			if len(obj) > *document.MaxProperties {
				return false
			}
		}
	}

	if document.Const != nil {
		if !reflect.DeepEqual(data, *document.Const) {
			return false
		}
	}

	if document.Type != nil {
		if document.Type.IsSingle {
			if !assertSimpleType(document.Type.Single, data) {
				return false
			}
		} else {
			// TODO: Early return.
			allFailed := true
			for _, simpleType := range document.Type.List {
				if assertSimpleType(simpleType, data) {
					allFailed = false
				}
			}

			if allFailed {
				return false
			}
		}
	}

	return true
}

func assertSimpleType(simpleType SimpleType, data interface{}) bool {
	switch simpleType {
	case IntegerSimpleType:
		if num, ok := data.(float64); !ok || num != math.Trunc(num) {
			return false
		}
	case NumberSimpleType:
		if _, ok := data.(float64); !ok {
			return false
		}
	case StringSimpleType:
		if _, ok := data.(string); !ok {
			return false
		}
	case ObjectSimpleType:
		if _, ok := data.(map[string]interface{}); !ok {
			return false
		}
	case ArraySimpleType:
		if _, ok := data.([]interface{}); !ok {
			return false
		}
	case BooleanSimpleType:
		if _, ok := data.(bool); !ok {
			return false
		}
	case NullSimpleType:
		if data != nil {
			return false
		}
	}

	return true
}
