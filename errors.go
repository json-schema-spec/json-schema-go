package jsonschema

// Error represents a union of errors that can arise from parsing and validating
// JSON Schemas.
type Error struct {
	invalidTypeValue    bool
	schemaNotObject     bool
	idNotString         bool
	invalidURI          bool
	refNotString        bool
	uriNotDefined       bool
	invalidArrayValue   bool
	invalidNumberValue  bool
	invalidNaturalValue bool
	invalidRegexpValue  bool
	invalidBoolValue    bool
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

func uriNotDefined() *Error {
	return &Error{uriNotDefined: true}
}

func invalidArrayValue() *Error {
	return &Error{invalidArrayValue: true}
}

func invalidNumberValue() *Error {
	return &Error{invalidNumberValue: true}
}

func invalidNaturalValue() *Error {
	return &Error{invalidNaturalValue: true}
}

func invalidRegexpValue() *Error {
	return &Error{invalidRegexpValue: true}
}

func invalidBoolValue() *Error {
	return &Error{invalidBoolValue: true}
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

// URINotDefined is whether an Error indicates a schema referred to a URI
// unknown to the validator.
func (e *Error) URINotDefined() bool {
	return e.uriNotDefined
}

// InvalidArrayValue is whether an Error indicates a keyword value which was
// expected to be an array, but was not.
func (e *Error) InvalidArrayValue() bool {
	return e.invalidArrayValue
}

// InvalidNumberValue is whether an Error indicates a keyword value which was
// expected to be a number, but was not.
func (e *Error) InvalidNumberValue() bool {
	return e.invalidNumberValue
}

// InvalidNaturalValue is whether an Error indicates a keyword value which was
// expected to be a natural number, but was not.
func (e *Error) InvalidNaturalValue() bool {
	return e.invalidNaturalValue
}

// InvalidRegexpValue is whether an Error indicates a keyword value which was
// expected to be a regexp, but was not.
func (e *Error) InvalidRegexpValue() bool {
	return e.invalidRegexpValue
}

// InvalidBoolValue is whether an Error indicates a keyword value which was
// expected to be a bool, but was not.
func (e *Error) InvalidBoolValue() bool {
	return e.invalidBoolValue
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
