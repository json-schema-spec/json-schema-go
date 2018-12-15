package jsonschema

type Validator struct {
	schema Schema
}

func NewValidator(schema Schema) Validator {
	return Validator{
		schema: schema,
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

	return true
}
