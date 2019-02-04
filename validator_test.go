package jsonschema

import (
	"net/url"
	"testing"

	"github.com/ucarion/json-pointer"

	"github.com/stretchr/testify/assert"
)

func TestValidatorSeal(t *testing.T) {
	testCases := []struct {
		name    string
		schemas []map[string]interface{}
		err     error
	}{
		{
			"empty object",
			[]map[string]interface{}{
				map[string]interface{}{},
			},
			nil,
		},
		{
			"type not string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": 3,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"type not a valid string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "invalid",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"items value not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": "foo",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"items value empty array",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": []interface{}{},
				},
			},
			nil,
		},
		{
			"element of items not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": []interface{}{
						"foo",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"references to non-existent URIs",
			[]map[string]interface{}{
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
			[]map[string]interface{}{
				map[string]interface{}{
					"enum": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number multipleOf value",
			[]map[string]interface{}{
				map[string]interface{}{
					"multipleOf": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number maximum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maximum": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number minimum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minimum": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number exclusiveMaximum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"exclusiveMaximum": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number exclusiveMinimum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"exclusiveMinimum": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-string pattern value",
			[]map[string]interface{}{
				map[string]interface{}{
					"pattern": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-regexp pattern value",
			[]map[string]interface{}{
				map[string]interface{}{
					"pattern": "[[[",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"element of additionalItems not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"additionalItems": []interface{}{
						"foo",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-boolean uniqueItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"uniqueItems": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"value of contains not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"contains": "foo",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-number minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-int minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": 3.14,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-positive minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": -2.0,
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-list required value",
			[]map[string]interface{}{
				map[string]interface{}{
					"required": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-string-containing list required value",
			[]map[string]interface{}{
				map[string]interface{}{
					"required": []interface{}{
						"foo",
						3,
						"baz",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object properties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"properties": "foo",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object value of properties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object patternProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"patternProperties": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-regexp patternProperties key",
			[]map[string]interface{}{
				map[string]interface{}{
					"patternProperties": map[string]interface{}{
						"[[[": map[string]interface{}{},
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object patternProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"patternProperties": map[string]interface{}{
						"[[[": "foobar",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object additionalProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"additionalProperties": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-object dependencies value",
			[]map[string]interface{}{
				map[string]interface{}{
					"dependencies": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-array and non-object dependencies property value",
			[]map[string]interface{}{
				map[string]interface{}{
					"dependencies": map[string]interface{}{
						"foo": 3,
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"invalid schema dependencies property value",
			[]map[string]interface{}{
				map[string]interface{}{
					"dependencies": map[string]interface{}{
						"foo": map[string]interface{}{
							"items": 3,
						},
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-string element of dependencies property",
			[]map[string]interface{}{
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
			ErrorInvalidSchema,
		},
		{
			"non-object value of propertyNames",
			[]map[string]interface{}{
				map[string]interface{}{
					"propertyNames": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-array value of allOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"allOf": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-schema element of allOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"allOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-array value of anyOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"anyOf": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-schema element of anyOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"anyOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-array value of oneOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"oneOf": "foobar",
				},
			},
			ErrorInvalidSchema,
		},
		{
			"non-schema element of oneOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"oneOf": []interface{}{
						"foobar",
					},
				},
			},
			ErrorInvalidSchema,
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
	schemas := []map[string]interface{}{
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
	schemas := []map[string]interface{}{
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
