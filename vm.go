package jsonschema

import (
	"errors"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/ucarion/json-pointer"
)

const epsilon = 1e-3

var errMaxErrors = errors.New("internal error for maximum errors")

type vm struct {
	// registry holds an arena of schemas
	registry registry

	// stack holds state used for error-message generation
	stack stack

	// errors holds all the errors to be produced
	errors vmErrors

	// maxStackDepth is the most number of $ref-s that can be followed at once
	maxStackDepth int

	// maxErrors is the most number of errors that can be reported
	maxErrors int
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

func newVM(registry registry, maxStackDepth, maxErrors int) vm {
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
		maxStackDepth: maxStackDepth,
		maxErrors:     maxErrors,
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
		return ErrNoSuchSchema
	}

	fragPtr, err := jsonpointer.New(uri.Fragment)
	if err != nil {
		return err
	}

	vm.pushNewSchema(uri, fragPtr.Tokens)
	err = vm.execSchema(schema, instance)
	if err == errMaxErrors {
		// not a real error -- just an internal flag to quit early
		return nil
	}

	return err
}

func (vm *vm) execSchema(schema schema, instance interface{}) error {
	if schema.Bool.IsSet {
		if !schema.Bool.Value {
			if err := vm.reportError(); err != nil {
				return err
			}
		}

		return nil
	}

	if schema.Ref.IsSet {
		if len(vm.stack.schemas) == vm.maxStackDepth {
			return ErrStackOverflow
		}

		refSchema := vm.registry.GetIndex(schema.Ref.Schema)

		schemaTokens := make([]string, len(schema.Ref.Ptr.Tokens))
		copy(schemaTokens, schema.Ref.Ptr.Tokens)

		vm.pushNewSchema(schema.Ref.BaseURI, schemaTokens)
		if err := vm.execSchema(refSchema, instance); err != nil {
			return err
		}
		vm.popSchema()
	}

	if schema.Not.IsSet {
		notSchema := vm.registry.GetIndex(schema.Not.Schema)
		notErrors, err := vm.pseudoExec(notSchema, instance)
		if err != nil {
			return err
		}

		if !notErrors {
			vm.pushSchemaToken("not")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}
	}

	if schema.If.IsSet {
		ifSchema := vm.registry.GetIndex(schema.If.Schema)
		ifErrors, err := vm.pseudoExec(ifSchema, instance)
		if err != nil {
			return err
		}

		if !ifErrors {
			if schema.Then.IsSet {
				thenSchema := vm.registry.GetIndex(schema.Then.Schema)

				vm.pushSchemaToken("then")
				if err := vm.execSchema(thenSchema, instance); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		} else {
			if schema.Else.IsSet {
				elseSchema := vm.registry.GetIndex(schema.Else.Schema)

				vm.pushSchemaToken("else")
				if err := vm.execSchema(elseSchema, instance); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}
	}

	if schema.Const.IsSet {
		if !reflect.DeepEqual(instance, schema.Const.Value) {
			vm.pushSchemaToken("const")
			if err := vm.reportError(); err != nil {
				return err
			}
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
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}
	}

	if schema.AllOf.IsSet {
		vm.pushSchemaToken("allOf")

		for i, index := range schema.AllOf.Schemas {
			allOfSchema := vm.registry.GetIndex(index)
			token := strconv.FormatInt(int64(i), 10)

			vm.pushSchemaToken(token)
			if err := vm.execSchema(allOfSchema, instance); err != nil {
				return err
			}
			vm.popSchemaToken()
		}

		vm.popSchemaToken()
	}

	if schema.AnyOf.IsSet {
		anyOfOk := false
		for _, index := range schema.AnyOf.Schemas {
			anyOfSchema := vm.registry.GetIndex(index)
			anyOfErrors, err := vm.pseudoExec(anyOfSchema, instance)
			if err != nil {
				return err
			}

			if !anyOfErrors {
				anyOfOk = true
				break
			}
		}

		if !anyOfOk {
			vm.pushSchemaToken("anyOf")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}
	}

	if schema.OneOf.IsSet {
		oneOfOk := false
		for _, index := range schema.OneOf.Schemas {
			oneOfSchema := vm.registry.GetIndex(index)
			oneOfErrors, err := vm.pseudoExec(oneOfSchema, instance)
			if err != nil {
				return err
			}

			if !oneOfErrors {
				if oneOfOk {
					oneOfOk = false
					break
				} else {
					oneOfOk = true
				}
			}
		}

		if !oneOfOk {
			vm.pushSchemaToken("oneOf")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}
	}

	switch val := instance.(type) {
	case nil:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeNull) {
			vm.pushSchemaToken("type")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}
	case bool:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeBoolean) {
			vm.pushSchemaToken("type")
			if err := vm.reportError(); err != nil {
				return err
			}
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
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.MultipleOf.IsSet {
			if math.Abs(math.Mod(val, schema.MultipleOf.Value)) > epsilon {
				vm.pushSchemaToken("multipleOf")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.Maximum.IsSet {
			if val > schema.Maximum.Value {
				vm.pushSchemaToken("maximum")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.Minimum.IsSet {
			if val < schema.Minimum.Value {
				vm.pushSchemaToken("minimum")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.ExclusiveMaximum.IsSet {
			if val > schema.ExclusiveMaximum.Value-epsilon {
				vm.pushSchemaToken("exclusiveMaximum")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.ExclusiveMinimum.IsSet {
			if val < schema.ExclusiveMinimum.Value+epsilon {
				vm.pushSchemaToken("exclusiveMinimum")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}
	case string:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeString) {
			vm.pushSchemaToken("type")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}

		if schema.MaxLength.IsSet {
			if utf8.RuneCountInString(val) > schema.MaxLength.Value {
				vm.pushSchemaToken("maxLength")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.MinLength.IsSet {
			if utf8.RuneCountInString(val) < schema.MinLength.Value {
				vm.pushSchemaToken("minLength")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.Pattern.IsSet {
			if !schema.Pattern.Value.MatchString(val) {
				vm.pushSchemaToken("pattern")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}
	case []interface{}:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeArray) {
			vm.pushSchemaToken("type")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}

		if schema.MaxItems.IsSet {
			if len(val) > schema.MaxItems.Value {
				vm.pushSchemaToken("maxItems")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.MinItems.IsSet {
			if len(val) < schema.MinItems.Value {
				vm.pushSchemaToken("minItems")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.UniqueItems.IsSet && schema.UniqueItems.Value {
		loop:
			for i := 0; i < len(val); i++ {
				for j := i + 1; j < len(val); j++ {
					if reflect.DeepEqual(val[i], val[j]) {
						vm.pushSchemaToken("uniqueItems")
						if err := vm.reportError(); err != nil {
							return err
						}
						vm.popSchemaToken()

						break loop
					}
				}
			}
		}

		if schema.Contains.IsSet {
			containsOk := false
			for _, elem := range val {
				containsSchema := vm.registry.GetIndex(schema.Contains.Schema)
				containsErrors, err := vm.pseudoExec(containsSchema, elem)
				if err != nil {
					return err
				}

				if !containsErrors {
					containsOk = true
					break
				}
			}

			if !containsOk {
				vm.pushSchemaToken("contains")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.Items.IsSet {
			if schema.Items.IsSingle {
				vm.pushSchemaToken("items")

				itemSchema := vm.registry.GetIndex(schema.Items.Schemas[0])
				for i, elem := range val {
					vm.pushInstanceToken(strconv.FormatInt(int64(i), 10))
					if err := vm.execSchema(itemSchema, elem); err != nil {
						return err
					}
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
					if err := vm.execSchema(itemSchema, val[i]); err != nil {
						return err
					}
					vm.popInstanceToken()
					vm.popSchemaToken()
				}
				vm.popSchemaToken()

				if schema.AdditionalItems.IsSet {
					vm.pushSchemaToken("additionalItems")

					additionalItemSchema := vm.registry.GetIndex(schema.AdditionalItems.Schema)
					for i := len(schema.Items.Schemas); i < len(val); i++ {
						token := strconv.FormatInt(int64(i), 10)

						vm.pushInstanceToken(token)
						if err := vm.execSchema(additionalItemSchema, val[i]); err != nil {
							return err
						}
						vm.popInstanceToken()
					}
					vm.popSchemaToken()
				}
			}
		}
	case map[string]interface{}:
		if schema.Type.IsSet && !schema.Type.contains(jsonTypeObject) {
			vm.pushSchemaToken("type")
			if err := vm.reportError(); err != nil {
				return err
			}
			vm.popSchemaToken()
		}

		if schema.MaxProperties.IsSet {
			if len(val) > schema.MaxProperties.Value {
				vm.pushSchemaToken("maxProperties")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.MinProperties.IsSet {
			if len(val) < schema.MinProperties.Value {
				vm.pushSchemaToken("minProperties")
				if err := vm.reportError(); err != nil {
					return err
				}
				vm.popSchemaToken()
			}
		}

		if schema.Required.IsSet {
			vm.pushSchemaToken("required")

			for i, property := range schema.Required.Properties {
				if _, ok := val[property]; !ok {
					vm.pushSchemaToken(strconv.FormatInt(int64(i), 10))
					if err := vm.reportError(); err != nil {
						return err
					}
					vm.popSchemaToken()
				}
			}

			vm.popSchemaToken()
		}

		for key, value := range val {
			isAdditional := true

			if schema.Properties.IsSet {
				if index, ok := schema.Properties.Schemas[key]; ok {
					isAdditional = false
					propertySchema := vm.registry.GetIndex(index)

					vm.pushSchemaToken("properties")
					vm.pushSchemaToken(key)
					vm.pushInstanceToken(key)
					if err := vm.execSchema(propertySchema, value); err != nil {
						return err
					}
					vm.popInstanceToken()
					vm.popSchemaToken()
					vm.popSchemaToken()
				}
			}

			if schema.PatternProperties.IsSet {
				for pattern, index := range schema.PatternProperties.Schemas {
					if pattern.MatchString(key) {
						isAdditional = false
						propertySchema := vm.registry.GetIndex(index)

						vm.pushSchemaToken("patternProperties")
						vm.pushSchemaToken(pattern.String())
						vm.pushInstanceToken(key)
						if err := vm.execSchema(propertySchema, value); err != nil {
							return err
						}
						vm.popInstanceToken()
						vm.popSchemaToken()
						vm.popSchemaToken()
					}
				}
			}

			if schema.AdditionalProperties.IsSet && isAdditional {
				propertySchema := vm.registry.GetIndex(schema.AdditionalProperties.Schema)

				vm.pushSchemaToken("additionalProperties")
				vm.pushInstanceToken(key)
				if err := vm.execSchema(propertySchema, value); err != nil {
					return err
				}
				vm.popInstanceToken()
				vm.popSchemaToken()
			}
		}

		if schema.Dependencies.IsSet {
			vm.pushSchemaToken("dependencies")

			for key, dep := range schema.Dependencies.Deps {
				vm.pushSchemaToken(key)

				if _, ok := val[key]; ok {
					if dep.IsSchema {
						propertySchema := vm.registry.GetIndex(dep.Schema)

						if err := vm.execSchema(propertySchema, val); err != nil {
							return err
						}
					} else {
						for i, property := range dep.Properties {
							if _, ok := val[property]; !ok {
								vm.pushSchemaToken(strconv.FormatInt(int64(i), 10))
								if err := vm.reportError(); err != nil {
									return err
								}
								vm.popSchemaToken()
							}
						}
					}
				}

				vm.popSchemaToken()
			}

			vm.popSchemaToken()
		}

		if schema.PropertyNames.IsSet {
			vm.pushSchemaToken("propertyNames")

			propertyNameSchema := vm.registry.GetIndex(schema.PropertyNames.Schema)
			for key := range val {
				vm.pushInstanceToken(key)
				if err := vm.execSchema(propertyNameSchema, key); err != nil {
					return err
				}
				vm.popInstanceToken()
			}

			vm.popSchemaToken()
		}
	default:
		// TODO a better error here
		panic("unexpected non-json input")
	}

	return nil
}

// pseudoExec determines whether a given schema accepts an instance, with the
// guarantee that the vm exits this function in the same state it was in when
// the function was called.
func (vm *vm) pseudoExec(schema schema, instance interface{}) (bool, error) {
	prevErrors := vm.errors
	vm.errors = vmErrors{
		hasErrors: false,
		errors:    []ValidationError{},
	}

	if err := vm.execSchema(schema, instance); err != nil {
		return false, err
	}

	pseudoErrors := vm.errors
	vm.errors = prevErrors

	return pseudoErrors.hasErrors, nil
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

func (vm *vm) reportError() error {
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

	if len(vm.errors.errors) == vm.maxErrors {
		return errMaxErrors
	}

	return nil
}
