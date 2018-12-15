package jsonschema

// Schema is a representation of a JSON Schema.
type Schema struct {
	IsTrivial    bool
	TrivialValue bool
	Document     Document
}

// Document is a representation of a nontrivial JSON Schema.
//
// Note that all fields of Document are pointers. This is because the JSON
// Schema spec does not require any particular fields to be present in a JSON
// Schema document.
//
// Document is meant to be unmarshalled from JSON. Where the JSON Schema spec
// allows for multiple types of data in the same field, Document uses a special
// struct to handle these polymorphic cases.
//
// Document does not implement the default values of JSON Schema keywords as
// part of its unmarshalling logic.
type Document struct {
	ID                   *string                      `json:"$id"`
	Schema               *string                      `json:"$schema"`
	Ref                  *string                      `json:"$ref"`
	Comment              *string                      `json:"$comment"`
	Title                *string                      `json:"title"`
	Description          *string                      `json:"description"`
	Default              *interface{}                 `json:"default"`
	ReadOnly             *bool                        `json:"readOnly"`
	Examples             *[]interface{}               `json:"examples"`
	MultipleOf           *float64                     `json:"multipleOf"`
	Maximum              *float64                     `json:"maximum"`
	ExclusiveMaximum     *float64                     `json:"exclusiveMaximum"`
	Minimum              *float64                     `json:"minimum"`
	ExclusiveMinimum     *float64                     `json:"exclusiveMinimum"`
	MaxLength            *int                         `json:"maxLength"`
	MinLength            *int                         `json:"minLength"`
	Pattern              *string                      `json:"pattern"`
	AdditionalItems      *Schema                      `json:"additionalItems"`
	Items                *SchemaItems                 `json:"items"`
	MaxItems             *int                         `json:"maxItems"`
	MinItems             *int                         `json:"minItems"`
	UniqueItems          *bool                        `json:"uniqueItems"`
	Contains             *Schema                      `json:"contains"`
	MaxProperties        *int                         `json:"maxProperties"`
	MinProperties        *int                         `json:"minProperties"`
	Required             *[]string                    `json:"required"`
	AdditionalProperties *Schema                      `json:"additionalProperties"`
	Definitions          *map[string]Schema           `json:"definitions"`
	Properties           *map[string]Schema           `json:"properties"`
	PatternProperties    *map[string]Schema           `json:"patternProperties"`
	Dependencies         *map[string]SchemaDependency `json:"dependencies"`
	PropertyNames        *Schema                      `json:"propertyNames"`
	Const                *interface{}                 `json:"const"`
	Enum                 *[]interface{}               `json:"enum"`
	Type                 *SchemaType                  `json:"type"`
	Format               *string                      `json:"format"`
	ContentMediaType     *string                      `json:"contentMediaType"`
	ContentEncoding      *string                      `json:"contentEncoding"`
	If                   *Schema                      `json:"if"`
	Then                 *Schema                      `json:"then"`
	Else                 *Schema                      `json:"else"`
	AllOf                *[]Schema                    `json:"allOf"`
	AnyOf                *[]Schema                    `json:"anyOf"`
	OneOf                *[]Schema                    `json:"oneOf"`
	Not                  *Schema                      `json:"not"`
}

// SchemaItems is either one Schema, or a nonempty list of schemas.
type SchemaItems struct {
	IsSingle bool
	Single   Schema
	List     []Schema
}

func (i *SchemaItems) UnmarshalJSON(data []byte) error {
	var single Schema
	var list []Schema

	isSingle, err := unmarshalWithFallback(data, &single, &list)
	if err != nil {
		return err
	}

	i.IsSingle = isSingle
	i.Single = single
	i.List = list
	return nil
}

// SchemaDependency is either a Schema or a list of strings.
type SchemaDependency struct {
	IsSchema bool
	Schema   Schema
	Strings  []string
}

func (d *SchemaDependency) UnmarshalJSON(data []byte) error {
	var schema Schema
	var strings []string

	isSchema, err := unmarshalWithFallback(data, &schema, &strings)
	if err != nil {
		return err
	}

	d.IsSchema = isSchema
	d.Schema = schema
	d.Strings = strings
	return nil
}

// SchemaType is either one SimpleType or a nonempty list of SimpleTypes.
type SchemaType struct {
	IsSingle bool
	Single   SimpleType
	List     []SimpleType
}

func (t *SchemaType) UnmarshalJSON(data []byte) error {
	var single SimpleType
	var list []SimpleType

	isSingle, err := unmarshalWithFallback(data, &single, &list)
	if err != nil {
		return err
	}

	t.IsSingle = isSingle
	t.Single = single
	t.List = list
	return nil
}

type SimpleType string

const (
	ArraySimpleType   SimpleType = "array"
	BooleanSimpleType            = "boolean"
	IntegerSimpleType            = "integer"
	NullSimpleType               = "null"
	NumberSimpleType             = "number"
	ObjectSimpleType             = "object"
	StringSimpleType             = "string"
)

// UnmarshalJSON implements unmarshalling a Schema from raw JSON bytes.
//
// Schema cannot be unmarshalled using Golang's default JSON unmarshalling
// logic. This is because JSON Schema supports "trivial" documents which are
// represented as JSON booleans, but Golang does not by default support
// unmarshalling booleans into Golang structs, such as Schema.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var trivialValue bool
	var document Document
	isTrivial, err := unmarshalWithFallback(data, &trivialValue, &document)
	if err != nil {
		return err
	}

	s.IsTrivial = isTrivial
	s.TrivialValue = trivialValue
	s.Document = document
	return nil
}
