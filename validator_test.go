package jsonschema

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Name      string         `json:"name"`
	Registry  []Schema       `json:"registry"`
	Schema    Schema         `json:"schema"`
	Instances []instanceCase `json:"instances"`
}

type instanceCase struct {
	Instance interface{}       `json:"instance"`
	Errors   []ValidationError `json:"errors"`
}

func TestValidator(t *testing.T) {
	err := filepath.Walk("tests", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			data, err := ioutil.ReadFile(path)
			assert.Nil(t, err)

			var testCases []testCase
			err = json.Unmarshal(data, &testCases)
			assert.Nil(t, err)

			for _, tt := range testCases {
				t.Run(tt.Name, func(t *testing.T) {
					validator := NewValidator(tt.Schema)

					for _, instance := range tt.Instances {
						result, err := validator.Validate(instance.Instance)
						assert.Nil(t, err)
						assert.Equal(t, instance.Errors, result.Errors)
					}
				})
			}
		})

		return nil
	})

	assert.Nil(t, err)
}
