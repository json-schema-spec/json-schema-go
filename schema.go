package jsonschema

import (
	"net/url"

	"github.com/ucarion/json-pointer"
)

type schema struct {
	ID               url.URL
	Ref              schemaRef
	Not              schemaNot
	If               schemaIf
	Then             schemaThen
	Else             schemaElse
	Type             schemaType
	Items            schemaItems
	Const            schemaConst
	Enum             schemaEnum
	MultipleOf       schemaMultipleOf
	Maximum          schemaMaximum
	Minimum          schemaMinimum
	ExclusiveMaximum schemaExclusiveMaximum
}

type schemaNot struct {
	IsSet  bool
	Schema int
}

type schemaIf struct {
	IsSet  bool
	Schema int
}

type schemaThen struct {
	IsSet  bool
	Schema int
}

type schemaElse struct {
	IsSet  bool
	Schema int
}

type schemaRef struct {
	IsSet   bool
	Schema  int
	URI     url.URL
	BaseURI url.URL
	Ptr     jsonpointer.Ptr
}

type schemaType struct {
	IsSet    bool
	IsSingle bool
	Types    []jsonType
}

type jsonType int

const (
	jsonTypeNull jsonType = iota + 1
	jsonTypeBoolean
	jsonTypeNumber
	jsonTypeInteger
	jsonTypeString
	jsonTypeArray
	jsonTypeObject
)

func (t schemaType) contains(typ jsonType) bool {
	for _, t := range t.Types {
		if t == typ {
			return true
		}
	}

	return false
}

type schemaItems struct {
	IsSet    bool
	IsSingle bool
	Schemas  []int
}

type schemaConst struct {
	IsSet bool
	Value interface{}
}

type schemaEnum struct {
	IsSet  bool
	Values []interface{}
}

type schemaMultipleOf struct {
	IsSet bool
	Value float64
}

type schemaMaximum struct {
	IsSet bool
	Value float64
}

type schemaMinimum struct {
	IsSet bool
	Value float64
}

type schemaExclusiveMaximum struct {
	IsSet bool
	Value float64
}
