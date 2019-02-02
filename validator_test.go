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
