package jsonschema

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/ucarion/json-pointer"
)

// DefaultMaxStackDepth is the default maximum number of cross-references a
// Validator will follow before returning ErrStackOverflow.
const DefaultMaxStackDepth = 128

// Validator compiles schemas and evaluates instances.
type Validator struct {
	schemas       []map[string]interface{}
	registry      registry
	maxStackDepth int
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
// error will be included, and the missing schema's ID will be returned in the
// list of url.URL.
//
// Each reference to a missing schema will result in an additional entry in the
// returned list. It is therefore possible for the same URI to appear multiple
// times in the list.
func NewValidator(schemas []map[string]interface{}) (Validator, []url.URL, error) {
	v := Validator{
		schemas:       schemas,
		maxStackDepth: DefaultMaxStackDepth,
	}

	missingURIs, err := v.seal()
	return v, missingURIs, err
}

func (v *Validator) seal() ([]url.URL, error) {
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

// Validate evaluates the given instance against the default schema of the
// Validator.
func (v *Validator) Validate(instance interface{}) (ValidationResult, error) {
	id := url.URL{}
	vm := newVM(v.registry, v.maxStackDepth)

	err := vm.Exec(id, instance)
	if err != nil {
		return ValidationResult{}, err
	}

	return vm.ValidationResult(), nil
}
