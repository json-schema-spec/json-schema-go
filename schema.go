package jsonschema

import (
	"errors"
	"net/url"

	"github.com/ucarion/json-pointer"
)

type schema struct {
	ID    url.URL
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
	IsSet   bool
	Schema  *schema
	URI     url.URL
	BaseURI url.URL
	Ptr     jsonpointer.Ptr
}

func parseRootSchema(input map[string]interface{}) (schema, error) {
	return parseSchema(true, url.URL{}, input)
}

func parseSubSchema(baseURI url.URL, input map[string]interface{}) (schema, error) {
	return parseSchema(false, baseURI, input)
}

func parseSchema(root bool, baseURI url.URL, input map[string]interface{}) (schema, error) {
	s := schema{}

	if root {
		idValue, ok := input["$id"]
		if ok {
			idStr, ok := idValue.(string)
			if !ok {
				return s, idNotString()
			}

			uri, err := url.Parse(idStr)
			if err != nil {
				return s, errors.New("$id is not valid URI")
			}

			baseURI = *uri
		}
	}

	refValue, ok := input["$ref"]
	if ok {
		// fmt.Println("setting $ref")
		refStr, ok := refValue.(string)
		if !ok {
			return s, refNotString()
		}

		// uri, err := url.Parse(refStr)
		// fmt.Printf("parsing refStr %#v %#v\n", refStr, baseURI)
		uri, err := baseURI.Parse(refStr)
		if err != nil {
			return s, invalidURI()
		}
		// fmt.Printf("result %#v %#v\n", refStr, uri)

		refBaseURI := *uri
		refBaseURI.Fragment = ""

		ptr, err := jsonpointer.New(uri.Fragment)
		if err != nil {
			return s, errors.New("$ref fragment is not a valid JSON Pointer")
		}

		s.Ref.IsSet = true
		s.Ref.URI = *uri
		s.Ref.BaseURI = refBaseURI
		s.Ref.Ptr = ptr
	}

	typeValue, ok := input["type"]
	if ok {
		switch typ := typeValue.(type) {
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
					return s, invalidTypeValue()
				}

				jsonTyp, err := parseJSONType(t)
				if err != nil {
					return s, err
				}

				s.Type.Types[i] = jsonTyp
			}
		default:
			return s, invalidTypeValue()
		}
	}

	itemsValue, ok := input["items"]
	if ok {
		switch items := itemsValue.(type) {
		case map[string]interface{}:
			subSchema, err := parseSubSchema(baseURI, items)
			if err != nil {
				return s, err
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
					return s, schemaNotObject()
				}

				subSchema, err := parseSubSchema(baseURI, item)
				if err != nil {
					return s, err
				}

				s.Items.Schemas[i] = subSchema
			}
		default:
			return s, schemaNotObject()
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
		return 0, invalidTypeValue()
	}
}

// func parseSchema(val interface{}) (schema, error) {
// 	s := schema{}

// 	value, ok := val.(map[string]interface{})
// 	if !ok {
// 		return s, errors.New("schemas must be map[string]interface{}")
// 	}

// 	baseURI := url.URL{}
// 	if id, ok := value["$id"]; ok {
// 	}

// 	result, err := parseSchemaWithBase(value, baseURI)
// 	if err != nil {
// 		return s, err
// 	}

// 	result.ID = baseURI
// 	return result, err
// }

// func parseSchemaWithBase(val interface{}, baseURI url.URL) (schema, error) {
// 	s := schema{}

// 	value, ok := val.(map[string]interface{})
// 	if !ok {
// 		return s, errors.New("schemas must be map[string]interface{}")
// 	}

// 	switch typ := value["type"].(type) {
// 	case string:
// 		jsonTyp, err := parseJSONType(typ)
// 		if err != nil {
// 			return s, err
// 		}

// 		s.Type.IsSet = true
// 		s.Type.IsSingle = true
// 		s.Type.Types = []jsonType{jsonTyp}
// 	case []interface{}:
// 		s.Type.IsSet = true
// 		s.Type.IsSingle = false
// 		s.Type.Types = make([]jsonType, len(typ))

// 		for i, t := range typ {
// 			t, ok := t.(string)
// 			if !ok {
// 				return s, errors.New("non-string value in type member") // todo error types
// 			}

// 			jsonTyp, err := parseJSONType(t)
// 			if err != nil {
// 				return s, err
// 			}

// 			s.Type.Types[i] = jsonTyp
// 		}
// 	}

// 	switch items := value["items"].(type) {
// 	case map[string]interface{}:
// 		subSchema, err := parseSchemaWithBase(items, baseURI)
// 		if err != nil {
// 			return s, err // todo compose errors
// 		}

// 		s.Items.IsSet = true
// 		s.Items.IsSingle = true
// 		s.Items.Schemas = []schema{subSchema}
// 	case []interface{}:
// 		s.Items.IsSet = true
// 		s.Items.IsSingle = false
// 		s.Items.Schemas = make([]schema, len(items))

// 		for i, item := range items {
// 			item, ok := item.(map[string]interface{})
// 			if !ok {
// 				return s, errors.New("non-map[string]interface{} value in items member") // todo error types
// 			}

// 			subSchema, err := parseSchema(item)
// 			if err != nil {
// 				return s, err
// 			}

// 			s.Items.Schemas[i] = subSchema
// 		}
// 	}

// 	return s, nil
// }
