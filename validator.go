package jsonschema

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/errors-go"
	"github.com/ucarion/json-pointer"
)

// Validator compiles schemas and evaluates instances.
type Validator struct {
	schemas  map[url.URL]map[string]interface{}
	registry map[url.URL]*Schema
}

var decoderConfig = mapstructure.DecoderConfig{
	DecodeHook: func(source, target reflect.Type, value interface{}) (interface{}, error) {
		if target == reflect.TypeOf(SchemaType{}) {
			switch val := value.(type) {
			case string:
				return map[string]interface{}{
					"IsSingle": true,
					"Single":   val,
				}, nil
			case []interface{}:
				return map[string]interface{}{
					"IsSingle": false,
					"List":     val,
				}, nil
			}
		}

		if target == reflect.TypeOf(SchemaItems{}) {
			switch val := value.(type) {
			case map[string]interface{}:
				return map[string]interface{}{
					"IsSingle": true,
					"Single":   val,
				}, nil
			case []interface{}:
				return map[string]interface{}{
					"IsSingle": false,
					"List":     val,
				}, nil
			}
		}

		return value, nil
	},
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
		schemas:  map[url.URL]map[string]interface{}{},
		registry: map[url.URL]*Schema{},
	}
}

// Register parses and compiles a Schema and adds it to the Validator's
// registry.
//
// If the registered schema lacks an "$id" keyword, then that schema will be
// considered the new "default" schema.
func (v *Validator) Register(schema map[string]interface{}) error {
	parsed, err := parseSchema(schema)
	if err != nil {
		return err
	}

	uri := url.URL{} // todo get this from the parsed schema
	v.schemas[uri] = schema
	v.registry[uri] = &parsed

	return nil
}

func (v *Validator) Seal() error {
	// The body of this loop will modify the map it iterates over. This is fine,
	// because entries created during iteration won't be visisted. Only entries
	// that exist prior to the start of the loop need to be visited.
	for uri := range v.registry {
		err := v.populateRefs(uri)
		if err != nil {
			return err
		}
	}

	return nil
}

// whatever function is doing JSON Pointer derefs needs to have access to a
// map[string]interface{} version of the schema

// TODO not sure about this TODO lol
//
// TODO make seal take an interface{} and a baseURI pointer. Its job is to parse
// the inputted schema, and assure its proper formatting.
//
// If the baseURI inputted is nil, then this is a top-level schema. The baseURI
// should be taken from the inputted schema.
func (v *Validator) populateRefs(uri url.URL) error {
	schema := v.registry[uri]

	if schema.Ref != nil {
		refURI, err := uri.Parse(*schema.Ref)
		if err != nil {
			return errors.Wrap(err, "error parsing $ref")
		}

		// After this line, we have no need for the fragment part of the URI.
		ptr, err := jsonpointer.New(refURI.Fragment)
		if err != nil {
			return errors.Wrap(err, "error parsing URI fragment as JSON Pointer")
		}

		fmt.Printf("writing to: %#v\n", *refURI)

		refBaseURI := *refURI
		refBaseURI.Fragment = ""
		refSchemaBaseValue, ok := v.schemas[refBaseURI]
		if !ok {
			return errors.New("no schema with URI") // todo error type
		}

		refSchemaValue, err := ptr.Eval(refSchemaBaseValue)
		if err != nil {
			fmt.Printf("%#v\n%#v\n%#v\n", refSchemaBaseValue, ptr, err)
			return errors.Wrap(err, "error evaluating $ref JSON Pointer")
		}

		refSchema, err := parseSchema(*refSchemaValue)
		if err != nil {
			fmt.Printf("%#v\n", *refSchemaValue)
			return errors.Wrap(err, "$ref points to non-schema value")
		}

		fmt.Printf("writing to: %#v\n", *refURI)

		v.registry[*refURI] = &refSchema
		err = v.populateRefs(*refURI)
		if err != nil {
			return err
		}

		schema.refSchema = &refSchema
	}

	return nil
}

// func (v *Validator) sealSchema(baseURI *url.URL, schema Schema) error {
// 	if schema.Ref != nil {

// 		// Check if the URI has already been added to the registry. If it has been,
// 		// then there is nothing more to be done here.
// 		if _, ok := v.registry[*uri]; ok {
// 			return nil
// 		}

// 		ptr, err := jsonpointer.New(uri.Fragment)
// 		if err != nil {
// 			return errors.Wrap(err, "error parsing URI fragment as JSON Pointer")
// 		}

// 		deref, err := ptr.Eval(schema)
// 		if err != nil {
// 			return errors.Wrap(err, "error dereferencing JSON Pointer")
// 		}

// 		derefSchema, ok := (*deref).(map[string]interface{})
// 		if !ok {
// 			return errors.New("JSON Pointer points to a non-schema value")
// 		}

// 		// TODO I need a Schema parsed object here to put into the registry.
// 		// Otherwise, the recursive call below may result in unbounded recursion.

// 		err = v.seal(baseURI, derefSchema)
// 		if err != nil {
// 			return err
// 		}
// 	}
// }

func parseSchema(value interface{}) (Schema, error) {
	var s Schema
	decoderConfig.Result = &s
	decoder, err := mapstructure.NewDecoder(&decoderConfig)

	if err != nil {
		return s, errors.Wrap(err, "error creating decoder")
	}

	err = decoder.Decode(value)
	if err != nil {
		return s, errors.Wrap(err, "error decoding schema")
	}

	return s, nil
}

// Validate validates an instance against the default schema of a Validator.
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
