package jsonschema

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/ucarion/json-pointer"
)

const epsilon = 1e-3

type vm struct {
	registry registry

	// stack holds state used for error-message generation
	stack stack

	// errors holds all the errors to be produced
	errors vmErrors
}

type vmErrors struct {
	hasErrors bool
	errors    []ValidationError
}

// stack keeps track of where we are in an instance and schema. It is meant to
// be used in cohort with the ordinary function call stack in order to produce
// error messages.
type stack struct {
	// instance is a stack of tokens into the instance, meant to construct a JSON
	// Pointer.
	instance []string

	// schema is a stack of stacks of tokens into the schema, meant to construct a
	// JSON Pointer. Each schema gets its own stack; because of cross-references,
	// there may be many schemas in use.
	schemas []schemaStack
}

// schemaStack keeps track of where we are in a schema, and which schema we are
// in.
type schemaStack struct {
	// id is the (non-relative) ID of the schema
	id url.URL

	// tokens is a stack of tokens into the schema, meant to construct a JSON
	// Pointer.
	tokens []string
}

func newVM(registry registry) vm {
	return vm{
		registry: registry,
		stack: stack{
			instance: []string{},
			schemas:  []schemaStack{},
		},
		errors: vmErrors{
			hasErrors: false,
			errors:    []ValidationError{},
		},
	}
}

func (vm *vm) ValidationResult() ValidationResult {
	return ValidationResult{
		Errors: vm.errors.errors,
	}
}

func (vm *vm) Exec(uri url.URL, instance interface{}) error {
	schema, ok := vm.registry.Get(uri)
	if !ok {
		// TODO custom error types
		return fmt.Errorf("no schema with uri: %#v", uri)
	}

	fragPtr, err := jsonpointer.New(uri.Fragment)
	if err != nil {
		// TODO wrap
		return err
	}

	vm.pushNewSchema(uri, fragPtr.Tokens)
	vm.execSchema(schema, instance)
	return nil
}

func (vm *vm) execSchema(schema schema, instance interface{}) {
	if schema.Ref.IsSet {
		refSchema := vm.registry.GetIndex(schema.Ref.Schema)

		schemaTokens := make([]string, len(schema.Ref.Ptr.Tokens))
		copy(schemaTokens, schema.Ref.Ptr.Tokens)

		vm.pushNewSchema(schema.Ref.BaseURI, schemaTokens)
		vm.execSchema(refSchema, instance)
		vm.popSchema()
	}

	if schema.Not.IsSet {
		notSchema := vm.registry.GetIndex(schema.Not.Schema)
		notErrors := vm.psuedoExec(notSchema, instance)

		if !notErrors {
			vm.pushSchemaToken("not")
			vm.reportError()
			vm.popSchemaToken()
		}
	}

	if schema.If.IsSet {
		ifSchema := vm.registry.GetIndex(schema.If.Schema)
		ifErrors := vm.psuedoExec(ifSchema, instance)

		if !ifErrors {
			if schema.Then.IsSet {
				thenSchema := vm.registry.GetIndex(schema.Then.Schema)

				vm.pushSchemaToken("then")
				vm.execSchema(thenSchema, instance)
				vm.popSchemaToken()
			}
		} else {
			if schema.Else.IsSet {
				elseSchema := vm.registry.GetIndex(schema.Else.Schema)

				vm.pushSchemaToken("else")
				vm.execSchema(elseSchema, instance)
				vm.popSchemaToken()
			}
		}
	}

	if schema.Const.IsSet {
		if !reflect.DeepEqual(instance, schema.Const.Value) {
			vm.pushSchemaToken("const")
			vm.reportError()
			vm.popSchemaToken()
		}
	}

	if schema.Enum.IsSet {
		enumOk := false
		for _, value := range schema.Enum.Values {
			if reflect.DeepEqual(instance, value) {
				enumOk = true
				break
			}
		}

		if !enumOk {
			vm.pushSchemaToken("enum")
			vm.reportError()
			vm.popSchemaToken()
		}
	}

	switch val := instance.(type) {
	case nil:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeNull) {
			vm.pushSchemaToken("type")
			vm.reportError()
			vm.popSchemaToken()
		}
	case bool:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeBoolean) {
			vm.pushSchemaToken("type")
			vm.reportError()
			vm.popSchemaToken()
		}
	case float64:
		if schema.Type.IsSet {
			typeOk := false
			if schema.Type.contains(jsonTypeInteger) {
				typeOk = val == math.Round(val)
			}

			if !typeOk && !schema.Type.contains(jsonTypeNumber) {
				vm.pushSchemaToken("type")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.MultipleOf.IsSet {
			if math.Abs(math.Mod(val, schema.MultipleOf.Value)) > epsilon {
				vm.pushSchemaToken("multipleOf")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.Maximum.IsSet {
			if val > schema.Maximum.Value {
				vm.pushSchemaToken("maximum")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.Minimum.IsSet {
			if val < schema.Minimum.Value {
				vm.pushSchemaToken("minimum")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.ExclusiveMaximum.IsSet {
			if val > schema.ExclusiveMaximum.Value-epsilon {
				vm.pushSchemaToken("exclusiveMaximum")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.ExclusiveMinimum.IsSet {
			if val < schema.ExclusiveMinimum.Value+epsilon {
				vm.pushSchemaToken("exclusiveMinimum")
				vm.reportError()
				vm.popSchemaToken()
			}
		}
	case string:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeString) {
			vm.pushSchemaToken("type")
			vm.reportError()
			vm.popSchemaToken()
		}

		if schema.MaxLength.IsSet {
			if utf8.RuneCountInString(val) > schema.MaxLength.Value {
				vm.pushSchemaToken("maxLength")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.MinLength.IsSet {
			if utf8.RuneCountInString(val) < schema.MinLength.Value {
				vm.pushSchemaToken("minLength")
				vm.reportError()
				vm.popSchemaToken()
			}
		}

		if schema.Pattern.IsSet {
			if !schema.Pattern.Value.MatchString(val) {
				vm.pushSchemaToken("pattern")
				vm.reportError()
				vm.popSchemaToken()
			}
		}
	case []interface{}:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeArray) {
			vm.pushSchemaToken("type")
			vm.reportError()
			vm.popSchemaToken()
		}

		if schema.Items.IsSet {
			if schema.Items.IsSingle {
				vm.pushSchemaToken("items")

				itemSchema := vm.registry.GetIndex(schema.Items.Schemas[0])
				for i, elem := range val {
					vm.pushInstanceToken(strconv.FormatInt(int64(i), 10))
					vm.execSchema(itemSchema, elem)
					vm.popInstanceToken()
				}
				vm.popSchemaToken()
			} else {
				vm.pushSchemaToken("items")
				for i := 0; i < len(schema.Items.Schemas) && i < len(val); i++ {
					itemSchema := vm.registry.GetIndex(schema.Items.Schemas[i])
					token := strconv.FormatInt(int64(i), 10)

					vm.pushInstanceToken(token)
					vm.pushSchemaToken(token)
					vm.execSchema(itemSchema, val[i])
					vm.popInstanceToken()
					vm.popSchemaToken()
				}
				vm.popSchemaToken()
			}
		}
	case map[string]interface{}:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeObject) {
			vm.pushSchemaToken("type")
			vm.reportError()
			vm.popSchemaToken()
		}
	default:
		// TODO a better error here
		panic("unexpected non-json input")
	}
}

// psuedoExec determines whether a given schema accepts an instance, with the
// guarantee that the vm exits this function in the same state it was in when
// the function was called.
func (vm *vm) psuedoExec(schema schema, instance interface{}) bool {
	prevErrors := vm.errors
	vm.errors = vmErrors{
		hasErrors: false,
		errors:    []ValidationError{},
	}

	vm.execSchema(schema, instance)
	pseudoErrors := vm.errors
	vm.errors = prevErrors

	return pseudoErrors.hasErrors
}

func (vm *vm) pushNewSchema(id url.URL, tokens []string) {
	vm.stack.schemas = append(vm.stack.schemas, schemaStack{
		id:     id,
		tokens: tokens,
	})
}

func (vm *vm) popSchema() {
	vm.stack.schemas = vm.stack.schemas[:len(vm.stack.schemas)-1]
}

func (vm *vm) pushSchemaToken(token string) {
	s := &vm.stack.schemas[len(vm.stack.schemas)-1]
	s.tokens = append(s.tokens, token)
}

func (vm *vm) popSchemaToken() {
	s := &vm.stack.schemas[len(vm.stack.schemas)-1]
	s.tokens = s.tokens[:len(s.tokens)-1]
}

func (vm *vm) pushInstanceToken(token string) {
	vm.stack.instance = append(vm.stack.instance, token)
}

func (vm *vm) popInstanceToken() {
	vm.stack.instance = vm.stack.instance[:len(vm.stack.instance)-1]
}

func (vm *vm) reportError() {
	schemaStack := vm.stack.schemas[len(vm.stack.schemas)-1]
	instancePath := make([]string, len(vm.stack.instance))
	schemaPath := make([]string, len(schemaStack.tokens))

	copy(instancePath, vm.stack.instance)
	copy(schemaPath, schemaStack.tokens)

	vm.errors.hasErrors = true
	vm.errors.errors = append(vm.errors.errors, ValidationError{
		InstancePath: jsonpointer.Ptr{Tokens: instancePath},
		SchemaPath:   jsonpointer.Ptr{Tokens: schemaPath},
		URI:          schemaStack.id,
	})
}
