package jsonschema

import (
	"net/url"

	"github.com/ucarion/json-pointer"
)

// Validator compiles schemas and evaluates instances.
type Validator struct {
	registry map[url.URL]Schema
}

// ValidationResult contains information on whether an instance successfully
// validated, as well as any relevant validation errors.
type ValidationResult struct {
	Errors []ValidationError
}

// ValidationError is a single error during validation.
type ValidationError struct {
	// A JSON Pointer to the part of the instance which was rejected.
	InstancePath jsonpointer.Ptr

	// A JSON Pointer to the part of the schema which rejected part of the
	// instance.
	SchemaPath jsonpointer.Ptr

	// The URI of the schema which rejected part of the instance.
	URI url.URL
}

// NewValidator constructs a new, empty Validator.
func NewValidator() Validator {
	return Validator{registry: map[url.URL]Schema{}}
}

// Register compiles a Schema and adds it to the Validator's registry. Once
// registered, schemas can validate instances or be referred to by other
// schemas.
//
// If the registered schema lacks an "$id" keyword, then that schema will be
// considered the "default" schema.
func (v Validator) Register(s Schema) error {
	v.registry[url.URL{}] = s
	return nil
}

// Validate validates an instance against the default schema of a Validator.
func (v Validator) Validate(instance interface{}) (ValidationResult, error) {
	id := url.URL{}

	vm := vm{
		registry: v.registry,
		stack: stack{
			instance: []string{},
			schemas:  []schemaStack{},
		},
		errors: []ValidationError{},
	}

	err := vm.exec(id, instance)
	if err != nil {
		return ValidationResult{}, err
	}

	return ValidationResult{
		Errors: vm.errors,
	}, nil
}
