package jsonschema

import (
	"net/url"
	"testing"

	"github.com/segmentio/errors-go"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSeal(t *testing.T) {
	testCases := []struct {
		name          string
		schemas       []map[string]interface{}
		undefinedURIs []url.URL
		err           string
	}{
		{
			"empty object",
			[]map[string]interface{}{
				map[string]interface{}{},
			},
			nil,
			"",
		},
		{
			"type not string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": 3,
				},
			},
			nil,
			"InvalidTypeValue",
		},
		{
			"type not a valid string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "invalid",
				},
			},
			nil,
			"InvalidTypeValue",
		},
		{
			"items value not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": "foo",
				},
			},
			nil,
			"SchemaNotObject",
		},
		{
			"items value empty array",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": []interface{}{},
				},
			},
			nil,
			"",
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
			nil,
			"SchemaNotObject",
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
			[]url.URL{
				url.URL{Scheme: "http", Host: "example.com", Path: "/2"},
				url.URL{Scheme: "http", Host: "example.com", Path: "/3"},
				url.URL{Scheme: "http", Host: "example.com", Path: "/4"},
				url.URL{Scheme: "http", Host: "example.com", Path: "/1"},
			},
			"URINotDefined",
		},
		{
			"non-array enum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"enum": "foobar",
				},
			},
			nil,
			"InvalidArrayValue",
		},
		{
			"non-number multipleOf value",
			[]map[string]interface{}{
				map[string]interface{}{
					"multipleOf": "foobar",
				},
			},
			nil,
			"InvalidNumberValue",
		},
		{
			"non-number maximum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maximum": "foobar",
				},
			},
			nil,
			"InvalidNumberValue",
		},
		{
			"non-number minimum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minimum": "foobar",
				},
			},
			nil,
			"InvalidNumberValue",
		},
		{
			"non-number exclusiveMaximum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"exclusiveMaximum": "foobar",
				},
			},
			nil,
			"InvalidNumberValue",
		},
		{
			"non-number exclusiveMinimum value",
			[]map[string]interface{}{
				map[string]interface{}{
					"exclusiveMinimum": "foobar",
				},
			},
			nil,
			"InvalidNumberValue",
		},
		{
			"non-number maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive maxLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxLength": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-number minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive minLength value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minLength": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-string pattern value",
			[]map[string]interface{}{
				map[string]interface{}{
					"pattern": 3.14,
				},
			},
			nil,
			"InvalidRegexpValue",
		},
		{
			"non-regexp pattern value",
			[]map[string]interface{}{
				map[string]interface{}{
					"pattern": "[[[",
				},
			},
			nil,
			"InvalidRegexpValue",
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
			nil,
			"SchemaNotObject",
		},
		{
			"non-number maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive maxItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxItems": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-number minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive minItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minItems": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-boolean uniqueItems value",
			[]map[string]interface{}{
				map[string]interface{}{
					"uniqueItems": "foobar",
				},
			},
			nil,
			"InvalidBoolValue",
		},
		{
			"value of contains not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"contains": "foo",
				},
			},
			nil,
			"SchemaNotObject",
		},
		{
			"non-number maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive maxProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"maxProperties": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-number minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": "foobar",
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-int minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": 3.14,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-positive minProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"minProperties": -2.0,
				},
			},
			nil,
			"InvalidNaturalValue",
		},
		{
			"non-list required value",
			[]map[string]interface{}{
				map[string]interface{}{
					"required": "foobar",
				},
			},
			nil,
			"InvalidPropertyList",
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
			nil,
			"InvalidPropertyList",
		},
		{
			"non-object properties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"properties": "foo",
				},
			},
			nil,
			"InvalidObjectValue",
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
			nil,
			"SchemaNotObject",
		},
		{
			"non-object patternProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"patternProperties": "foobar",
				},
			},
			nil,
			"InvalidObjectValue",
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
			nil,
			"InvalidRegexpValue",
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
			nil,
			"SchemaNotObject",
		},
		{
			"non-object additionalProperties value",
			[]map[string]interface{}{
				map[string]interface{}{
					"additionalProperties": "foobar",
				},
			},
			nil,
			"SchemaNotObject",
		},
		{
			"non-object dependencies value",
			[]map[string]interface{}{
				map[string]interface{}{
					"dependencies": "foobar",
				},
			},
			nil,
			"InvalidDependenciesValue",
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
			nil,
			"InvalidDependencyValue",
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
			nil,
			"SchemaNotObject",
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
			nil,
			"InvalidPropertyList",
		},
		{
			"non-object value of propertyNames",
			[]map[string]interface{}{
				map[string]interface{}{
					"propertyNames": "foobar",
				},
			},
			nil,
			"SchemaNotObject",
		},
		{
			"non-array value of allOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"allOf": "foobar",
				},
			},
			nil,
			"InvalidArrayValue",
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
			nil,
			"SchemaNotObject",
		},
		{
			"non-array value of anyOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"anyOf": "foobar",
				},
			},
			nil,
			"InvalidArrayValue",
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
			nil,
			"SchemaNotObject",
		},
		{
			"non-array value of oneOf",
			[]map[string]interface{}{
				map[string]interface{}{
					"oneOf": "foobar",
				},
			},
			nil,
			"InvalidArrayValue",
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
			nil,
			"SchemaNotObject",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, undefinedURIs, err := NewValidator(tt.schemas)

			assert.Equal(t, tt.undefinedURIs, undefinedURIs)
			if tt.err == "" {
				assert.Equal(t, nil, err)
			} else {
				assert.True(t, errors.Is(tt.err, err), "expected %#v to be %s", err, tt.err)
			}
		})
	}
}
