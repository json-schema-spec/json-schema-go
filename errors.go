package jsonschema

// Error represents a union of errors that can arise from parsing and validating
// JSON Schemas.
type Error struct {
	invalidTypeValue bool
	schemaNotObject  bool
	idNotString      bool
	invalidURI       bool
	refNotString     bool
	invalidRef       bool
}

func invalidTypeValue() *Error {
	return &Error{invalidTypeValue: true}
}

func schemaNotObject() *Error {
	return &Error{schemaNotObject: true}
}

func idNotString() *Error {
	return &Error{idNotString: true}
}

func invalidURI() *Error {
	return &Error{invalidURI: true}
}

func refNotString() *Error {
	return &Error{refNotString: true}
}

func invalidRef() *Error {
	return &Error{invalidRef: true}
}

// InvalidTypeValue is whether an Error indicates a "type" keyword value wasn't
// in a valid format.
func (e *Error) InvalidTypeValue() bool {
	return e.invalidTypeValue
}

// SchemaNotObject is whether an Error indicates a schema was not an object.
func (e *Error) SchemaNotObject() bool {
	return e.schemaNotObject
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	if e.InvalidTypeValue() {
		return "invalid type value"
	}

	if e.SchemaNotObject() {
		return "schema not object"
	}

	return "unknown error"
}
