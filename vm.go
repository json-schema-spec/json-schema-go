package jsonschema

type vm struct {
	instanceTokens stack
	schemaTokens   stack
	schemaURIs     stack
	registry       map[string]Schema
}

func execSchema(registry map[string]Schema, uri string, instance interface{}) (ValidationResult, error) {
	errors := []ValidationError{}
	vm := vm{
		instanceTokens: stack{elems: []string{}},
		schemaTokens:   stack{elems: []string{}},
		schemaURIs:     stack{elems: []string{}},
	}

	return ValidationResult{}, nil
}
