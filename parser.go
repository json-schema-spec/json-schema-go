package jsonschema

import (
	"errors"
	"math"
	"net/url"
	"regexp"
	"strconv"

	"github.com/ucarion/json-pointer"
)

type parser struct {
	registry *registry
	baseURI  url.URL
	tokens   []string
}

func parseRootSchema(registry *registry, input map[string]interface{}) (schema, error) {
	return parseSubSchema(registry, url.URL{}, []string{}, input)
}

func parseSubSchema(registry *registry, baseURI url.URL, tokens []string, input map[string]interface{}) (schema, error) {
	p := parser{
		registry: registry,
		tokens:   tokens,
		baseURI:  baseURI,
	}

	index, err := p.Parse(input)
	if err != nil {
		return schema{}, err
	}

	return registry.GetIndex(index), nil
}

func (p *parser) Push(token string) {
	p.tokens = append(p.tokens, token)
}

func (p *parser) Pop() {
	p.tokens = p.tokens[:len(p.tokens)-1]
}

func (p *parser) URI() url.URL {
	ptr := jsonpointer.Ptr{Tokens: p.tokens}

	url := p.baseURI
	url.Fragment = ptr.String()
	return url
}

func (p *parser) Parse(input map[string]interface{}) (int, error) {
	s := schema{}

	if len(p.tokens) == 0 {
		idValue, ok := input["$id"]
		if ok {
			idStr, ok := idValue.(string)
			if !ok {
				return -1, idNotString()
			}

			uri, err := url.Parse(idStr)
			if err != nil {
				return -1, errors.New("$id is not valid URI")
			}

			p.baseURI = *uri
			s.ID = *uri
		}
	}

	refValue, ok := input["$ref"]
	if ok {
		refStr, ok := refValue.(string)
		if !ok {
			return -1, refNotString()
		}

		uri, err := p.baseURI.Parse(refStr)
		if err != nil {
			return -1, invalidURI()
		}

		refBaseURI := *uri
		refBaseURI.Fragment = ""

		ptr, err := jsonpointer.New(uri.Fragment)
		if err != nil {
			return -1, errors.New("$ref fragment is not a valid JSON Pointer")
		}

		s.Ref.IsSet = true
		s.Ref.URI = *uri
		s.Ref.BaseURI = refBaseURI
		s.Ref.Ptr = ptr
	}

	notValue, ok := input["not"]
	if ok {
		switch not := notValue.(type) {
		case map[string]interface{}:
			p.Push("not")

			subSchema, err := p.Parse(not)
			if err != nil {
				return -1, err
			}

			s.Not.IsSet = true
			s.Not.Schema = subSchema

			p.Pop()
		default:
			return -1, schemaNotObject()
		}
	}

	ifValue, ok := input["if"]
	if ok {
		switch ifx := ifValue.(type) {
		case map[string]interface{}:
			p.Push("if")

			subSchema, err := p.Parse(ifx)
			if err != nil {
				return -1, err
			}

			s.If.IsSet = true
			s.If.Schema = subSchema

			p.Pop()
		default:
			return -1, schemaNotObject()
		}
	}

	thenValue, ok := input["then"]
	if ok {
		switch then := thenValue.(type) {
		case map[string]interface{}:
			p.Push("then")

			subSchema, err := p.Parse(then)
			if err != nil {
				return -1, err
			}

			s.Then.IsSet = true
			s.Then.Schema = subSchema

			p.Pop()
		default:
			return -1, schemaNotObject()
		}
	}

	elseValue, ok := input["else"]
	if ok {
		switch elsex := elseValue.(type) {
		case map[string]interface{}:
			p.Push("else")

			subSchema, err := p.Parse(elsex)
			if err != nil {
				return -1, err
			}

			s.Else.IsSet = true
			s.Else.Schema = subSchema

			p.Pop()
		default:
			return -1, schemaNotObject()
		}
	}

	typeValue, ok := input["type"]
	if ok {
		switch typ := typeValue.(type) {
		case string:
			jsonTyp, err := parseJSONType(typ)
			if err != nil {
				return -1, err
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
					return -1, invalidTypeValue()
				}

				jsonTyp, err := parseJSONType(t)
				if err != nil {
					return -1, err
				}

				s.Type.Types[i] = jsonTyp
			}
		default:
			return -1, invalidTypeValue()
		}
	}

	itemsValue, ok := input["items"]
	if ok {
		switch items := itemsValue.(type) {
		case map[string]interface{}:
			p.Push("items")

			subSchema, err := p.Parse(items)
			if err != nil {
				return -1, err
			}

			s.Items.IsSet = true
			s.Items.IsSingle = true
			s.Items.Schemas = []int{subSchema}

			p.Pop()
		case []interface{}:
			p.Push("items")

			s.Items.IsSet = true
			s.Items.IsSingle = false
			s.Items.Schemas = make([]int, len(items))

			for i, item := range items {
				p.Push(strconv.FormatInt(int64(i), 10))

				item, ok := item.(map[string]interface{})
				if !ok {
					return -1, schemaNotObject()
				}

				subSchema, err := p.Parse(item)
				if err != nil {
					return -1, err
				}

				s.Items.Schemas[i] = subSchema
				p.Pop()
			}

			p.Pop()
		default:
			return -1, schemaNotObject()
		}
	}

	constValue, ok := input["const"]
	if ok {
		s.Const.IsSet = true
		s.Const.Value = constValue
	}

	enumValue, ok := input["enum"]
	if ok {
		enumArray, ok := enumValue.([]interface{})
		if !ok {
			return -1, invalidArrayValue()
		}

		s.Enum.IsSet = true
		s.Enum.Values = enumArray
	}

	multipleOfValue, ok := input["multipleOf"]
	if ok {
		multipleOfNumber, ok := multipleOfValue.(float64)
		if !ok {
			return -1, invalidNumberValue()
		}

		s.MultipleOf.IsSet = true
		s.MultipleOf.Value = multipleOfNumber
	}

	maximumValue, ok := input["maximum"]
	if ok {
		maximumNumber, ok := maximumValue.(float64)
		if !ok {
			return -1, invalidNumberValue()
		}

		s.Maximum.IsSet = true
		s.Maximum.Value = maximumNumber
	}

	minimumValue, ok := input["minimum"]
	if ok {
		minimumNumber, ok := minimumValue.(float64)
		if !ok {
			return -1, invalidNumberValue()
		}

		s.Minimum.IsSet = true
		s.Minimum.Value = minimumNumber
	}

	exclusiveMaximumValue, ok := input["exclusiveMaximum"]
	if ok {
		exclusiveMaximumNumber, ok := exclusiveMaximumValue.(float64)
		if !ok {
			return -1, invalidNumberValue()
		}

		s.ExclusiveMaximum.IsSet = true
		s.ExclusiveMaximum.Value = exclusiveMaximumNumber
	}

	exclusiveMinimumValue, ok := input["exclusiveMinimum"]
	if ok {
		exclusiveMinimumNumber, ok := exclusiveMinimumValue.(float64)
		if !ok {
			return -1, invalidNumberValue()
		}

		s.ExclusiveMinimum.IsSet = true
		s.ExclusiveMinimum.Value = exclusiveMinimumNumber
	}

	maxLengthValue, ok := input["maxLength"]
	if ok {
		maxLengthNumber, ok := maxLengthValue.(float64)
		if !ok {
			return -1, invalidNaturalValue()
		}

		maxLengthInt, rem := math.Modf(maxLengthNumber)
		if rem > epsilon {
			return -1, invalidNaturalValue()
		}

		if maxLengthInt < 0 {
			return -1, invalidNaturalValue()
		}

		s.MaxLength.IsSet = true
		s.MaxLength.Value = int(maxLengthInt)
	}

	minLengthValue, ok := input["minLength"]
	if ok {
		minLengthNumber, ok := minLengthValue.(float64)
		if !ok {
			return -1, invalidNaturalValue()
		}

		minLengthInt, rem := math.Modf(minLengthNumber)
		if rem > epsilon {
			return -1, invalidNaturalValue()
		}

		if minLengthInt < 0 {
			return -1, invalidNaturalValue()
		}

		s.MinLength.IsSet = true
		s.MinLength.Value = int(minLengthInt)
	}

	patternValue, ok := input["pattern"]
	if ok {
		patternString, ok := patternValue.(string)
		if !ok {
			return -1, invalidRegexpValue()
		}

		patternRegexp, err := regexp.Compile(patternString)
		if err != nil {
			return -1, invalidRegexpValue()
		}

		s.Pattern.IsSet = true
		s.Pattern.Value = patternRegexp
	}

	index := p.registry.Insert(p.URI(), s)
	return index, nil
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
