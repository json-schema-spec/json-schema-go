package jsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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

func TestValidatorSpec(t *testing.T) {
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
					schemas := []map[string]interface{}{tt.Schema}
					schemas = append(schemas, tt.Registry...)
					validator, undefinedURIs, err := NewValidator(schemas)

					assert.Empty(t, undefinedURIs)
					assert.Nil(t, err)
					if err != nil {
						return
					}

					for i, instance := range tt.Instances {
						t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
							result, err := validator.Validate(instance.Instance)
							assert.Nil(t, err)

							expected := make([]ValidationError, len(instance.Errors))
							for i, e := range instance.Errors {
								instancePath, _ := jsonpointer.New(e.InstancePath)
								schemaPath, _ := jsonpointer.New(e.SchemaPath)
								uri, _ := url.Parse(e.URI)

								expected[i] = ValidationError{
									InstancePath: instancePath,
									SchemaPath:   schemaPath,
									URI:          *uri,
								}
							}

							sort.Slice(expected, func(i, j int) bool {
								return expected[i].SchemaPath.String() < expected[j].SchemaPath.String()
							})

							sort.Slice(result.Errors, func(i, j int) bool {
								return result.Errors[i].SchemaPath.String() < result.Errors[j].SchemaPath.String()
							})

							assert.Equal(t, expected, result.Errors)
						})
					}
				})
			}
		})

		return nil
	})

	assert.Nil(t, err)
}
