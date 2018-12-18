package jsonschema

import (
	"errors"
	"fmt"
)

// Schema represents a JSON Schema document.
type Schema struct {
	True   bool
	False  bool
	Bool   BoolAssertions
	Number NumberAssertions
	String StringAssertions
	Array  ArrayAssertions
	Object ObjectAssertions
	Null   NullAssertions
}

type BoolAssertions struct {
	Reject bool
}

type NumberAssertions struct {
	Reject  bool
	Integer bool
}

type StringAssertions struct {
	Reject bool
}

type ArrayAssertions struct {
	Reject bool
}

type ObjectAssertions struct {
	Reject bool
}

type NullAssertions struct {
	Reject bool
}

const (
	ArraySimpleType   = "array"
	BooleanSimpleType = "boolean"
	IntegerSimpleType = "integer"
	NullSimpleType    = "null"
	NumberSimpleType  = "number"
	ObjectSimpleType  = "object"
	StringSimpleType  = "string"
)

func NewSchema(doc interface{}) (Schema, error) {
	switch val := doc.(type) {
	case bool:
		return Schema{
			True:  val,
			False: !val,
		}, nil
	case map[string]interface{}:
		return parseDoc(val)
	default:
		return Schema{}, errors.New("`doc` must be bool or map[string]interface{}")
	}
}

func parseDoc(doc map[string]interface{}) (Schema, error) {
	var s Schema

	if typeVal, ok := doc["type"]; ok {
		// If `type` is specified, then a type assertion will occur, and all types
		// are rejected by default.
		s.Bool.Reject = true
		s.Number.Reject = true
		s.String.Reject = true
		s.Array.Reject = true
		s.Object.Reject = true
		s.Null.Reject = true

		switch typedType := typeVal.(type) {
		case string:
			err := acceptType(&s, typedType)
			if err != nil {
				return s, err
			}
		case []interface{}:
			for _, t := range typedType {
				t, ok := t.(string)
				if !ok {
					return s, errors.New("when `type` is []interface{}, members must be string")
				}

				err := acceptType(&s, t)
				if err != nil {
					return s, err
				}
			}
		default:
			return s, errors.New("`type` must be a string or []interface{}")
		}
	}

	return s, nil
}

func acceptType(s *Schema, t string) error {
	switch t {
	case BooleanSimpleType:
		s.Bool.Reject = false
	case NumberSimpleType:
		s.Number.Reject = false
	case IntegerSimpleType:
		s.Number.Reject = false
		s.Number.Integer = true
	case StringSimpleType:
		s.String.Reject = false
	case ArraySimpleType:
		s.Array.Reject = false
	case ObjectSimpleType:
		s.Object.Reject = false
	case NullSimpleType:
		s.Null.Reject = false
	default:
		return fmt.Errorf("`type` is invalid: %s", t)
	}

	return nil
}
