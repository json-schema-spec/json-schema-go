package jsonschema

import (
	"net/url"

	"github.com/ucarion/json-pointer"
)

// DefaultMaxStackDepth is the default value for MaxStackDepth in
// ValidatorConfig.
const DefaultMaxStackDepth = 128

// Validator compiles schemas and evaluates instances.
type Validator struct {
	schemas       []map[string]interface{}
	registry      registry
	maxStackDepth int
	maxErrors     int
}

// ValidatorConfig contains configuration for a Validator.
type ValidatorConfig struct {
	// MaxStackDepth is the maximum number of cross-references a Validator will
	// follow before returning ErrStackOverflow.
	MaxStackDepth int

	// MaxErrors is the maximum number of errors to return before the Validator
	// quits early.
	//
	// A value of zero indicates to produce all errors.
	MaxErrors int
}

// ValidationResult contains information on whether an instance successfully
// validated, as well as any relevant validation errors.
type ValidationResult struct {
	Errors     []ValidationError
	Overflowed bool
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

// NewValidator constructs a new Validator that will use the given schemas.
//
// If any of the given schemas lack an "$id" field, then the last such schema
// will be used as the default schema of the Validator.
//
// If any schemas cross-reference schemas not present in the given list, then an
// instance of ErrMissingURIs will be returned. This value can be inspected to
// find which URIs are missing.
//
// Each reference to a missing schema will result in an additional entry in the
// returned list. It is therefore possible for the same URI to appear multiple
// times in the list.
func NewValidator(schemas []map[string]interface{}) (Validator, error) {
	return NewValidatorWithConfig(schemas, ValidatorConfig{
		MaxStackDepth: DefaultMaxStackDepth,
	})
}

// NewValidatorWithConfig constructs a new Validator that will use the given
// schemas and config.
//
// See NewValidator for how schemas will be used. See ValidatorConfig for
// configuration options.
func NewValidatorWithConfig(schemas []map[string]interface{}, config ValidatorConfig) (Validator, error) {
	v := Validator{
		schemas:       schemas,
		maxStackDepth: config.MaxStackDepth,
		maxErrors:     config.MaxErrors,
	}

	err := v.seal()
	return v, err
}

func (v *Validator) seal() error {
	registry := newRegistry(32)
	rawSchemas := map[url.URL]map[string]interface{}{}

	for _, schema := range v.schemas {
		parsed, err := parseRootSchema(&registry, schema)
		if err != nil {
			return err
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
					return err
				}

				rawRefSchema, err := ptr.Eval(rawSchema)
				if err != nil {
					return err
				}

				refSchemaObject, ok := (*rawRefSchema).(map[string]interface{})
				if !ok {
					return ErrorInvalidSchema
				}

				_, err = parseSubSchema(&registry, baseURI, ptr.Tokens, refSchemaObject)
				if err != nil {
					return err
				}
			} else {
				undefinedURIs = append(undefinedURIs, baseURI)
			}
		}

		missingURIs = registry.PopulateRefs()
	}

	if len(undefinedURIs) > 0 {
		return ErrMissingURIs{URIs: undefinedURIs}
	}

	v.registry = registry
	return nil
}

// Validate evaluates the given instance against the default schema of the
// Validator.
func (v *Validator) Validate(instance interface{}) (ValidationResult, error) {
	id := url.URL{}
	vm := newVM(v.registry, v.maxStackDepth, v.maxErrors)

	err := vm.Exec(id, instance)
	if err != nil {
		return ValidationResult{}, err
	}

	return vm.ValidationResult(), nil
}
