package jsonschema

import (
	"errors"
	"fmt"
	"net/url"
)

// ErrStackOverflow indicates that the evaluator overflowed its internal stack
// while evaluating a schema. This can arise from schemas that have cyclical
// definitions using the "$ref" keyword.
var ErrStackOverflow = errors.New("stack overflow evaluating schema")

// ErrorInvalidSchema indicates that an inputted schema was invalid.
var ErrorInvalidSchema = errors.New("invalid schema")

// ErrMissingURIs indicates that some schemas were referred to, but were not
// known to the Validator.
type ErrMissingURIs struct {
	// URIs is a list of fragment-less URIs of schemas that are missing.
	URIs []url.URL
}

// Error fulfills the error interface.
func (e ErrMissingURIs) Error() string {
	return fmt.Sprintf("missing schemas with URIs: %v", e.URIs)
}
