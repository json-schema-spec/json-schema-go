package jsonschema

import (
	"net/url"
	"testing"

	jsonpointer "github.com/json-schema-spec/json-pointer-go"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSeal(t *testing.T) {
	testCases := []struct {
		name    string
		schemas []interface{}
		err     error
	}{
		{
			"empty object",
			[]interface{}{
				map[string]interface{}{},
			},
			nil,
		},
		{
			"type not string",
			[]interface{}{
				map[string]interface{}{
					"type": 3,
				},
			},
			ErrInvalidSchema,
		},
		{
			"type not a valid string",
			[]interface{}{
				map[string]interface{}{
					"type": "invalid",
				},
			},
			ErrInvalidSchema,
		},
		{
			"items value not object",
			[]interface{}{
				map[string]interface{}{
					"items": "foo",
				},
			},
			ErrInvalidSchema,
		},
		{
			"items value empty array",
			[]interface{}{
				map[string]interface{}{
					"items": []interface{}{},
				},
			},
			nil,
		},
		{
			"element of items not object",
			[]interface{}{
				map[string]interface{}{
					"items": []interface{}{
						"foo",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"references to non-existent URIs",
			[]interface{}{
				map[string]interface{}{
					"$ref": "http://example.com/1",
					"items": []interface{}{
						map[string]interface{}{
							"$ref": "http://example.com/2",
						},
						map[string]interface{}{
							"$ref": "http://example.com/3",
						},
						map[string]interface{}{
							"$ref": "http://example.com/4#/fragment",
						},
					},
				},
			},
			ErrMissingURIs{
				URIs: []url.URL{
					url.URL{Scheme: "http", Host: "example.com", Path: "/2"},
					url.URL{Scheme: "http", Host: "example.com", Path: "/3"},
					url.URL{Scheme: "http", Host: "example.com", Path: "/4"},
					url.URL{Scheme: "http", Host: "example.com", Path: "/1"},
				},
			},
		},
		{
			"non-array enum value",
			[]interface{}{
				map[string]interface{}{
					"enum": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number multipleOf value",
			[]interface{}{
				map[string]interface{}{
					"multipleOf": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number maximum value",
			[]interface{}{
				map[string]interface{}{
					"maximum": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number minimum value",
			[]interface{}{
				map[string]interface{}{
					"minimum": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number exclusiveMaximum value",
			[]interface{}{
				map[string]interface{}{
					"exclusiveMaximum": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number exclusiveMinimum value",
			[]interface{}{
				map[string]interface{}{
					"exclusiveMinimum": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number maxLength value",
			[]interface{}{
				map[string]interface{}{
					"maxLength": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int maxLength value",
			[]interface{}{
				map[string]interface{}{
					"maxLength": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive maxLength value",
			[]interface{}{
				map[string]interface{}{
					"maxLength": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number minLength value",
			[]interface{}{
				map[string]interface{}{
					"minLength": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int minLength value",
			[]interface{}{
				map[string]interface{}{
					"minLength": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive minLength value",
			[]interface{}{
				map[string]interface{}{
					"minLength": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-string pattern value",
			[]interface{}{
				map[string]interface{}{
					"pattern": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-regexp pattern value",
			[]interface{}{
				map[string]interface{}{
					"pattern": "[[[",
				},
			},
			ErrInvalidSchema,
		},
		{
			"element of additionalItems not object",
			[]interface{}{
				map[string]interface{}{
					"additionalItems": []interface{}{
						"foo",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number maxItems value",
			[]interface{}{
				map[string]interface{}{
					"maxItems": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int maxItems value",
			[]interface{}{
				map[string]interface{}{
					"maxItems": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive maxItems value",
			[]interface{}{
				map[string]interface{}{
					"maxItems": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number minItems value",
			[]interface{}{
				map[string]interface{}{
					"minItems": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int minItems value",
			[]interface{}{
				map[string]interface{}{
					"minItems": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive minItems value",
			[]interface{}{
				map[string]interface{}{
					"minItems": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-boolean uniqueItems value",
			[]interface{}{
				map[string]interface{}{
					"uniqueItems": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"value of contains not object",
			[]interface{}{
				map[string]interface{}{
					"contains": "foo",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number maxProperties value",
			[]interface{}{
				map[string]interface{}{
					"maxProperties": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int maxProperties value",
			[]interface{}{
				map[string]interface{}{
					"maxProperties": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive maxProperties value",
			[]interface{}{
				map[string]interface{}{
					"maxProperties": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-number minProperties value",
			[]interface{}{
				map[string]interface{}{
					"minProperties": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-int minProperties value",
			[]interface{}{
				map[string]interface{}{
					"minProperties": 3.14,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-positive minProperties value",
			[]interface{}{
				map[string]interface{}{
					"minProperties": -2.0,
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-list required value",
			[]interface{}{
				map[string]interface{}{
					"required": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-string-containing list required value",
			[]interface{}{
				map[string]interface{}{
					"required": []interface{}{
						"foo",
						3,
						"baz",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object properties value",
			[]interface{}{
				map[string]interface{}{
					"properties": "foo",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object value of properties value",
			[]interface{}{
				map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object patternProperties value",
			[]interface{}{
				map[string]interface{}{
					"patternProperties": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-regexp patternProperties key",
			[]interface{}{
				map[string]interface{}{
					"patternProperties": map[string]interface{}{
						"[[[": map[string]interface{}{},
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object patternProperties value",
			[]interface{}{
				map[string]interface{}{
					"patternProperties": map[string]interface{}{
						"[[[": "foobar",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object additionalProperties value",
			[]interface{}{
				map[string]interface{}{
					"additionalProperties": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object dependencies value",
			[]interface{}{
				map[string]interface{}{
					"dependencies": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-array and non-object dependencies property value",
			[]interface{}{
				map[string]interface{}{
					"dependencies": map[string]interface{}{
						"foo": 3,
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"invalid schema dependencies property value",
			[]interface{}{
				map[string]interface{}{
					"dependencies": map[string]interface{}{
						"foo": map[string]interface{}{
							"items": 3,
						},
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-string element of dependencies property",
			[]interface{}{
				map[string]interface{}{
					"dependencies": map[string]interface{}{
						"foo": []interface{}{
							"bar",
							"baz",
							3,
							"quux",
						},
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-object value of propertyNames",
			[]interface{}{
				map[string]interface{}{
					"propertyNames": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-array value of allOf",
			[]interface{}{
				map[string]interface{}{
					"allOf": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-schema element of allOf",
			[]interface{}{
				map[string]interface{}{
					"allOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-array value of anyOf",
			[]interface{}{
				map[string]interface{}{
					"anyOf": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-schema element of anyOf",
			[]interface{}{
				map[string]interface{}{
					"anyOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-array value of oneOf",
			[]interface{}{
				map[string]interface{}{
					"oneOf": "foobar",
				},
			},
			ErrInvalidSchema,
		},
		{
			"non-schema element of oneOf",
			[]interface{}{
				map[string]interface{}{
					"oneOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrInvalidSchema,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewValidator(tt.schemas)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestValidatorOverflow(t *testing.T) {
	schemas := []interface{}{
		map[string]interface{}{
			"$ref": "#",
		},
	}

	validator, err := NewValidator(schemas)
	assert.NoError(t, err)

	_, err = validator.Validate(nil)
	assert.Equal(t, ErrStackOverflow, err)
}

func TestValidatorMaxErrors(t *testing.T) {
	schemas := []interface{}{
		map[string]interface{}{
			"allOf": []interface{}{
				map[string]interface{}{
					"type": "null",
				},
				map[string]interface{}{
					"$ref": "#",
				},
			},
		},
	}

	validationError := ValidationError{
		InstancePath: jsonpointer.Ptr{Tokens: []string{}},
		SchemaPath:   jsonpointer.Ptr{Tokens: []string{"allOf", "0", "type"}},
	}

	expectedResult := []ValidationError{}
	for i := 0; i < 5; i++ {
		expectedResult = append(expectedResult, validationError)
	}

	validator, err := NewValidatorWithConfig(schemas, ValidatorConfig{
		MaxErrors:     5,
		MaxStackDepth: 10, // max depth > errors, so no stack overflow should occur
	})

	assert.NoError(t, err)

	result, err := validator.Validate(true)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result.Errors)
}

func TestValidatorIsValid(t *testing.T) {
	valid := ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	invalid := ValidationResult{
		Errors: make([]ValidationError, 1),
	}

	assert.True(t, valid.IsValid())
	assert.False(t, invalid.IsValid())
}

func TestValidatorValidateURI(t *testing.T) {
	schemas := []interface{}{
		map[string]interface{}{
			"$id":  "http://example.com/foo",
			"type": "null",
		},
		map[string]interface{}{
			"$id":  "http://example.com/bar",
			"type": "integer",
		},
	}

	validator, err := NewValidator(schemas)
	assert.NoError(t, err)

	uriFoo, err := url.Parse("http://example.com/foo")
	assert.NoError(t, err)

	result, err := validator.ValidateURI(*uriFoo, nil)
	assert.NoError(t, err)
	assert.True(t, result.IsValid())

	uriBar, err := url.Parse("http://example.com/bar")
	assert.NoError(t, err)

	result, err = validator.ValidateURI(*uriBar, nil)
	assert.NoError(t, err)
	assert.False(t, result.IsValid())

	uriBaz, err := url.Parse("http://example.com/baz")
	assert.NoError(t, err)

	_, err = validator.ValidateURI(*uriBaz, nil)
	assert.Equal(t, ErrNoSuchSchema, err)
}
