package jsonschema

import (
	"net/url"
	"regexp"

	jsonpointer "github.com/json-schema-spec/json-pointer-go"
)

type schema struct {
	Bool                 schemaBool
	ID                   url.URL
	Ref                  schemaRef
	Not                  schemaNot
	If                   schemaIf
	Then                 schemaThen
	Else                 schemaElse
	Type                 schemaType
	Items                schemaItems
	AdditionalItems      schemaAdditionalItems
	Const                schemaConst
	Enum                 schemaEnum
	MultipleOf           schemaMultipleOf
	Maximum              schemaMaximum
	Minimum              schemaMinimum
	ExclusiveMaximum     schemaExclusiveMaximum
	ExclusiveMinimum     schemaExclusiveMinimum
	MaxLength            schemaMaxLength
	MinLength            schemaMinLength
	Pattern              schemaPattern
	MaxItems             schemaMaxItems
	MinItems             schemaMinItems
	UniqueItems          schemaUniqueItems
	Contains             schemaContains
	MaxProperties        schemaMaxProperties
	MinProperties        schemaMinProperties
	Required             schemaRequired
	Properties           schemaProperties
	PatternProperties    schemaPatternProperties
	AdditionalProperties schemaAdditionalProperties
	Dependencies         schemaDependencies
	PropertyNames        schemaPropertyNames
	AllOf                schemaAllOf
	AnyOf                schemaAnyOf
	OneOf                schemaOneOf
}

type schemaBool struct {
	IsSet bool
	Value bool
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

type schemaExclusiveMinimum struct {
	IsSet bool
	Value float64
}

type schemaMaxLength struct {
	IsSet bool
	Value int
}

type schemaMinLength struct {
	IsSet bool
	Value int
}

type schemaPattern struct {
	IsSet bool
	Value *regexp.Regexp
}

type schemaAdditionalItems struct {
	IsSet  bool
	Schema int
}

type schemaMaxItems struct {
	IsSet bool
	Value int
}

type schemaMinItems struct {
	IsSet bool
	Value int
}

type schemaUniqueItems struct {
	IsSet bool
	Value bool
}

type schemaContains struct {
	IsSet  bool
	Schema int
}

type schemaMaxProperties struct {
	IsSet bool
	Value int
}

type schemaMinProperties struct {
	IsSet bool
	Value int
}

type schemaRequired struct {
	IsSet      bool
	Properties []string
}

type schemaProperties struct {
	IsSet   bool
	Schemas map[string]int
}

type schemaPatternProperties struct {
	IsSet   bool
	Schemas map[*regexp.Regexp]int
}

type schemaAdditionalProperties struct {
	IsSet  bool
	Schema int
}

type schemaDependencies struct {
	IsSet bool
	Deps  map[string]schemaDependency
}

type schemaDependency struct {
	IsSchema   bool
	Schema     int
	Properties []string
}

type schemaPropertyNames struct {
	IsSet  bool
	Schema int
}

type schemaAllOf struct {
	IsSet   bool
	Schemas []int
}

type schemaAnyOf struct {
	IsSet   bool
	Schemas []int
}

type schemaOneOf struct {
	IsSet   bool
	Schemas []int
}
