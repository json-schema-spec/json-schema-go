package jsonschema

import (
	"encoding/json"
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

		return nil
	})

	assert.NotNil(t, err)
	assert.Nil(t, err)
}
