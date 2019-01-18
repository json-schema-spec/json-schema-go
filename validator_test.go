package jsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/ucarion/json-pointer"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	Name      string                   `json:"name"`
	Registry  []map[string]interface{} `json:"registry"`
	Schema    map[string]interface{}   `json:"schema"`
	Instances []instanceCase           `json:"instances"`
}

type instanceCase struct {
	Instance interface{}     `json:"instance"`
	Errors   []instanceError `json:"errors"`
}

type instanceError struct {
	InstancePath string `json:"instancePath"`
	SchemaPath   string `json:"schemaPath"`
	URI          string `json:"uri"`
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
					validator := NewValidator()
					for _, schema := range tt.Registry {
						err := validator.Register(schema)
						assert.Nil(t, err)
					}

					err := validator.Register(tt.Schema)
					assert.Nil(t, err)

					err = validator.Seal()
					assert.Nil(t, err)

					for i, instance := range tt.Instances {
						t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
							result, err := validator.Validate(instance.Instance)
							assert.Nil(t, err)

							assert.Equal(t, len(instance.Errors), len(result.Errors))
							for i := 0; i < len(instance.Errors) && i < len(result.Errors); i++ {
								expected := instance.Errors[i]
								instancePath, _ := jsonpointer.New(expected.InstancePath)
								schemaPath, _ := jsonpointer.New(expected.SchemaPath)
								uri, _ := url.Parse(expected.URI)

								assert.Equal(t, instancePath, result.Errors[i].InstancePath)
								assert.Equal(t, schemaPath, result.Errors[i].SchemaPath)
								assert.Equal(t, *uri, result.Errors[i].URI)
							}
						})
					}
				})
			}
		})

		return nil
	})

	assert.Nil(t, err)
}
