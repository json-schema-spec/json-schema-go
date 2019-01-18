package jsonschema

import (
	"net/url"

	"github.com/segmentio/errors-go"
	"github.com/ucarion/json-pointer"
)

type schema struct {
	Ref   schemaRef
	Type  schemaType
	Items schemaItems
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
	Schemas  []schema
}

type schemaRef struct {
	IsSet  bool
	Schema *schema
	URI    url.URL
	Ptr    jsonpointer.Ptr
}

func parseSchema(val interface{}) (schema, error) {
	s := schema{}

	value, ok := val.(map[string]interface{})
	if !ok {
		return s, errors.New("schemas must be map[string]interface{}")
	}

	if ref, ok := value["$ref"]; ok {
		refStr, ok := ref.(string)
		if !ok {
			return s, errors.New("$ref values must be a string")
		}

		uri, err := url.Parse(refStr)
		if err != nil {
			return s, errors.New("$ref is not valid URI")
		}

		ptr, err := jsonpointer.New(uri.Fragment)
		if err != nil {
			return s, errors.New("$ref fragment is not a valid JSON Pointer")
		}

		s.Ref.IsSet = true
		s.Ref.URI = *uri
		s.Ref.Ptr = ptr
	}

	switch typ := value["type"].(type) {
	case string:
		jsonTyp, err := parseJSONType(typ)
		if err != nil {
			return s, err
		}

		s.Type.IsSet = true
		s.Type.IsSingle = true
		s.Type.Types = []jsonType{jsonTyp}
	case []interface{}:
		s.Type.IsSet = true
		s.Type.IsSingle = false
		s.Type.Types = make([]jsonType, len(typ))

		for i, t := range typ {
			t, ok := t.(string)
			if !ok {
				return s, errors.New("non-string value in type member") // todo error types
			}

			jsonTyp, err := parseJSONType(t)
			if err != nil {
				return s, err
			}

			s.Type.Types[i] = jsonTyp
		}
	}

	switch items := value["items"].(type) {
	case map[string]interface{}:
		subSchema, err := parseSchema(items)
		if err != nil {
			return s, err // todo compose errors
		}

		s.Items.IsSet = true
		s.Items.IsSingle = true
		s.Items.Schemas = []schema{subSchema}
	case []interface{}:
		s.Items.IsSet = true
		s.Items.IsSingle = false
		s.Items.Schemas = make([]schema, len(items))

		for i, item := range items {
			item, ok := item.(map[string]interface{})
			if !ok {
				return s, errors.New("non-map[string]interface{} value in items member") // todo error types
			}

			subSchema, err := parseSchema(item)
			if err != nil {
				return s, err
			}

			s.Items.Schemas[i] = subSchema
		}
	}

	return s, nil
}

func parseJSONType(typ string) (jsonType, error) {
	switch typ {
	case "null":
		return jsonTypeNull, nil
	case "boolean":
		return jsonTypeBoolean, nil
	case "number":
		return jsonTypeNumber, nil
	case "integer":
		return jsonTypeInteger, nil
	case "string":
		return jsonTypeString, nil
	case "array":
		return jsonTypeArray, nil
	case "object":
		return jsonTypeObject, nil
	default:
		return 0, errors.New("bad json type") // todo error types
	}
}
