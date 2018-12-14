package jsonschema

type Validator struct {
}

func NewValidator(schema Schema) Validator {
	return Validator{}
}

func (v Validator) IsValid(data interface{}) bool {
	return false
}
