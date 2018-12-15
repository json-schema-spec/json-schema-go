package jsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testDir = "tests"

type testSuite struct {
	Schema Schema     `json:"schema"`
	Tests  []testCase `json:"tests"`
}

type testCase struct {
	Data  interface{} `json:"data"`
	Valid bool        `json:"valid"`
}

func TestValidate(t *testing.T) {
	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var suites []testSuite
		err = json.Unmarshal(data, &suites)
		if err != nil {
			return err
		}

		for i, suite := range suites {
			validator := NewValidator(suite.Schema)

			for j, test := range suite.Tests {
				name := fmt.Sprintf("%s/%d/%d", path, i, j)
				t.Run(name, func(t *testing.T) {
					assert.Equal(t, test.Valid, validator.IsValid(test.Data))
				})
			}
		}

		return nil
	})

	assert.Nil(t, err)
}
