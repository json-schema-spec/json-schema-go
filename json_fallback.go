package jsonschema

import (
	"encoding/json"
)

// unmarshalWithFallback attempts to unmarshal data into primary, falling back
// to fallback if that fails due to a type error.
//
// If no error is retuned, then the returned boolean indicates whether
// marshaling into primary succeeded.
func unmarshalWithFallback(data []byte, primary, fallback interface{}) (bool, error) {
	// Attempt to marshal into primary.
	err := json.Unmarshal(data, primary)
	if err != nil {
		// The primary marshal failed. Detect if this was due to a type error.
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			// The primary marshal failed due to a type error. Try again on the
			// fallback.
			return false, json.Unmarshal(data, fallback)
		}

		// The primary marshal failed, but not due to a type error.
		return false, err
	}

	// The primary marshal succeeded.
	return true, nil
}
