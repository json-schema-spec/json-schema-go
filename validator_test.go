package jsonschema

import (
	"testing"

	"github.com/segmentio/errors-go"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSeal(t *testing.T) {
	testCases := []struct {
		name    string
		schemas []map[string]interface{}
		err     string
	}{
		{
			"empty object",
			[]map[string]interface{}{
				map[string]interface{}{},
			},
			"",
		},
		{
			"type not string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": 3,
				},
			},
			"InvalidTypeValue",
		},
		{
			"type not a valid string",
			[]map[string]interface{}{
				map[string]interface{}{
					"type": "invalid",
				},
			},
			"InvalidTypeValue",
		},
		{
			"items value not object",
			[]map[string]interface{}{
				map[string]interface{}{
					"items": "foo",
				},
			},
			"SchemaNotObject",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator()
			for _, s := range tt.schemas {
				validator.Register(s)
			}

			err := validator.Seal()
			if tt.err == "" {
				assert.Equal(t, nil, err)
			} else {
				assert.True(t, errors.Is(tt.err, err), "expected %#v to be %s", err, tt.err)
			}
		})
	}
}
