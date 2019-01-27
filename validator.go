package jsonschema

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/ucarion/json-pointer"
)

// Validator compiles schemas and evaluates instances.
type Validator struct {
	schemas  []map[string]interface{}
	registry map[url.URL]schema
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

func (v *Validator) Seal() error {
	registry := map[url.URL]schema{}
	rawSchemas := map[url.URL]map[string]interface{}{}
	urisToProcess := []url.URL{}

	for i, schema := range v.schemas {
		parsed, err := parseRootSchema(schema)
		if err != nil {
			return errors.Wrapf(err, "errors parsing schema %d", i)
		}

		// fmt.Printf("parsed: %#v\n", parsed)

		registry[parsed.ID] = parsed
		rawSchemas[parsed.ID] = schema
		urisToProcess = append(urisToProcess, parsed.ID)
	}

	for len(urisToProcess) > 0 {
		toProcess := urisToProcess
		urisToProcess = []url.URL{}

		for _, uri := range toProcess {
			// fmt.Println("processing", uri.String())
			schema := registry[uri]

			// fmt.Printf("%#v\n", rawSchemas)
			// fmt.Printf("%#v\n", schema)
			if !schema.Ref.IsSet || schema.Ref.Schema != nil {
				// fmt.Println("skipping")
				continue
			}

			// fmt.Println("process ref", schema.Ref.URI.String())

			if registrySchema, ok := registry[schema.Ref.URI]; ok {
				schema.Ref.Schema = &registrySchema
				registry[uri] = schema
				continue
			}

			baseRawSchema, ok := rawSchemas[schema.Ref.BaseURI]
			if !ok {
				return invalidRef()
			}

			rawRefValue, err := schema.Ref.Ptr.Eval(baseRawSchema)
			if err != nil {
				return errors.Wrapf(err, "error evaluating JSON Pointer %#v", schema.Ref.Ptr)
			}

			rawRefSchema, ok := (*rawRefValue).(map[string]interface{})
			if !ok {
				return schemaNotObject()
			}

			refSchema, err := parseSubSchema(schema.Ref.BaseURI, rawRefSchema)
			if err != nil {
				return err
			}

			// update the referring schema and add it back to registry
			schema.Ref.Schema = &refSchema
			registry[uri] = schema

			// add the referred schema to registry and enqueue it for processing
			registry[schema.Ref.URI] = refSchema
			urisToProcess = append(urisToProcess, schema.Ref.URI)
		}
	}

	fmt.Println("--- DUMPING REGISTRY FROM VALIDATOR ---")
	for k, v := range registry {
		fmt.Printf("%#v :: %#v\n", k.String(), v)
	}
	fmt.Println("---")

	fmt.Printf("what does root refer to? %#v\n", registry[url.URL{}].Ref.Schema)

	v.registry = registry
	return nil
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

// // Register parses and compiles a Schema and adds it to the Validator's
// // registry.
// //
// // If the registered schema lacks an "$id" keyword, then that schema will be
// // considered the new "default" schema.
// func (v *Validator) Register(schema map[string]interface{}) error {
// 	parsed, err := parseSchema(schema)
// 	if err != nil {
// 		return err
// 	}

// 	v.schemas[parsed.ID] = schema
// 	v.registry[parsed.ID] = &parsed

// 	return nil
// }

// func (v *Validator) Seal() error {
// 	// The body of this loop will modify the map it iterates over. This is fine,
// 	// because entries created during iteration won't be visisted. Only entries
// 	// that exist prior to the start of the loop need to be visited.
// 	for _, schema := range v.registry {
// 		// err := v.populateRefs(uri)
// 		err := v.populateRefs(schema)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (v *Validator) populateRefs(schema *schema) error {
// 	err := v.populateRef(schema)
// 	if err != nil {
// 		return err
// 	}

// 	if schema.Items.Schemas != nil {
// 		for i := range schema.Items.Schemas {
// 			err := v.populateRefs(&schema.Items.Schemas[i])
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func (v *Validator) populateRef(schema *schema) error {
// 	if schema.Ref.IsSet && schema.Ref.Schema == nil {
// 		fmt.Printf("populating REF! %#v\n", schema.Ref.URI)

// 		registryRefSchema, ok := v.registry[schema.Ref.URI]
// 		if ok {
// 			schema.Ref.Schema = registryRefSchema
// 			return nil
// 		}

// 		refSchemaBaseValue, ok := v.schemas[schema.Ref.BaseURI]
// 		if !ok {
// 			return errors.New("no schema with URI") // todo error type
// 		}

// 		refSchemaValue, err := schema.Ref.Ptr.Eval(refSchemaBaseValue)
// 		if err != nil {
// 			return errors.Wrap(err, "error evaluating $ref JSON Pointer")
// 		}

// 		refSchema, err := parseSchema(*refSchemaValue)
// 		if err != nil {
// 			return errors.Wrap(err, "$ref points to non-schema value")
// 		}

// 		v.registry[schema.Ref.URI] = &refSchema
// 		err = v.populateRefs(&refSchema)
// 		if err != nil {
// 			return err
// 		}

// 		schema.Ref.Schema = &refSchema
// 	}

// 	return nil
// }

// // func (v *Validator) populateRefs(uri url.URL) error {
// // 	schema := v.registry[uri]

// // 	if schema.Ref.IsSet && schema.Ref.Schema == nil {
// // 		refSchemaBaseValue, ok := v.schemas[schema.Ref.BaseURI]
// // 		if !ok {
// // 			return errors.New("no schema with URI") // todo error type
// // 		}

// // 		refSchemaValue, err := schema.Ref.Ptr.Eval(refSchemaBaseValue)
// // 		if err != nil {
// // 			return errors.Wrap(err, "error evaluating $ref JSON Pointer")
// // 		}

// // 		refSchema, err := parseSchema(*refSchemaValue)
// // 		if err != nil {
// // 			return errors.Wrap(err, "$ref points to non-schema value")
// // 		}

// // 		v.registry[schema.Ref.URI] = &refSchema
// // 		err = v.populateRefs(schema.Ref.URI)
// // 		if err != nil {
// // 			return err
// // 		}

// // 		schema.Ref.Schema = &refSchema
// // 	}

// // 	// if schema.Items.Schemas != nil {
// // 	// 	for _, subSchema := range schema.Items.Schemas {

// // 	// 	}
// // 	// }

// // 	return nil
// // }

// // Validate validates an instance against the default schema of a Validator.
// func (v *Validator) Validate(instance interface{}) (ValidationResult, error) {
// }
