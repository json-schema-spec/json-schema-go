package jsonschema

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/ucarion/json-pointer"
)

// Validator compiles schemas and evaluates instances.
type Validator struct {
	schemas  []map[string]interface{}
	registry registry
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
	return Validator{
		schemas: []map[string]interface{}{},
	}
}

func (v *Validator) Register(schema map[string]interface{}) {
	v.schemas = append(v.schemas, schema)
}

func (v *Validator) Seal() ([]url.URL, error) {
	registry := newRegistry(32)
	rawSchemas := map[url.URL]map[string]interface{}{}

	for i, schema := range v.schemas {
		parsed, err := parseRootSchema(&registry, schema)
		if err != nil {
			return nil, errors.Wrapf(err, "errors parsing schema %d", i)
		}

		rawSchemas[parsed.ID] = schema
	}

	missingURIs := registry.PopulateRefs() // uris which must be accounted for
	undefinedURIs := []url.URL{}           // uris which cannot be accounted for

	for len(missingURIs) > 0 && len(undefinedURIs) == 0 {
		for _, uri := range missingURIs {
			baseURI := uri
			baseURI.Fragment = ""

			if rawSchema, ok := rawSchemas[baseURI]; ok {
				ptr, err := jsonpointer.New(uri.Fragment)
				if err != nil {
					return nil, err
				}

				rawRefSchema, err := ptr.Eval(rawSchema)
				if err != nil {
					return nil, err
				}

				refSchemaObject, ok := (*rawRefSchema).(map[string]interface{})
				if !ok {
					return nil, schemaNotObject()
				}

				_, err = parseSubSchema(&registry, baseURI, ptr.Tokens, refSchemaObject)
				if err != nil {
					return nil, err
				}
			} else {
				undefinedURIs = append(undefinedURIs, baseURI)
			}
		}

		missingURIs = registry.PopulateRefs()
	}

	if len(undefinedURIs) > 0 {
		return undefinedURIs, uriNotDefined()
	}

	v.registry = registry
	return nil, nil
}

func (v *Validator) Validate(instance interface{}) (ValidationResult, error) {
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
