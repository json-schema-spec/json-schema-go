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
				return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
		}

		uri, err := p.baseURI.Parse(refStr)
		if err != nil {
			return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
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
					return -1, ErrorInvalidSchema
				}

				jsonTyp, err := parseJSONType(t)
				if err != nil {
					return -1, err
				}

				s.Type.Types[i] = jsonTyp
			}
		default:
			return -1, ErrorInvalidSchema
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
					return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
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
			return -1, ErrorInvalidSchema
		}

		s.Enum.IsSet = true
		s.Enum.Values = enumArray
	}

	multipleOfValue, ok := input["multipleOf"]
	if ok {
		multipleOfNumber, ok := multipleOfValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.MultipleOf.IsSet = true
		s.MultipleOf.Value = multipleOfNumber
	}

	maximumValue, ok := input["maximum"]
	if ok {
		maximumNumber, ok := maximumValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.Maximum.IsSet = true
		s.Maximum.Value = maximumNumber
	}

	minimumValue, ok := input["minimum"]
	if ok {
		minimumNumber, ok := minimumValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.Minimum.IsSet = true
		s.Minimum.Value = minimumNumber
	}

	exclusiveMaximumValue, ok := input["exclusiveMaximum"]
	if ok {
		exclusiveMaximumNumber, ok := exclusiveMaximumValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.ExclusiveMaximum.IsSet = true
		s.ExclusiveMaximum.Value = exclusiveMaximumNumber
	}

	exclusiveMinimumValue, ok := input["exclusiveMinimum"]
	if ok {
		exclusiveMinimumNumber, ok := exclusiveMinimumValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.ExclusiveMinimum.IsSet = true
		s.ExclusiveMinimum.Value = exclusiveMinimumNumber
	}

	maxLengthValue, ok := input["maxLength"]
	if ok {
		maxLengthNumber, ok := maxLengthValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		maxLengthInt, rem := math.Modf(maxLengthNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if maxLengthInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MaxLength.IsSet = true
		s.MaxLength.Value = int(maxLengthInt)
	}

	minLengthValue, ok := input["minLength"]
	if ok {
		minLengthNumber, ok := minLengthValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		minLengthInt, rem := math.Modf(minLengthNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if minLengthInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MinLength.IsSet = true
		s.MinLength.Value = int(minLengthInt)
	}

	patternValue, ok := input["pattern"]
	if ok {
		patternString, ok := patternValue.(string)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		patternRegexp, err := regexp.Compile(patternString)
		if err != nil {
			return -1, ErrorInvalidSchema
		}

		s.Pattern.IsSet = true
		s.Pattern.Value = patternRegexp
	}

	additionalItemsValue, ok := input["additionalItems"]
	if ok {
		additionalItemsObject, ok := additionalItemsValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("additionalItems")

		subSchema, err := p.Parse(additionalItemsObject)
		if err != nil {
			return -1, err
		}

		s.AdditionalItems.IsSet = true
		s.AdditionalItems.Schema = subSchema

		p.Pop()
	}

	maxItemsValue, ok := input["maxItems"]
	if ok {
		maxItemsNumber, ok := maxItemsValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		maxItemsInt, rem := math.Modf(maxItemsNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if maxItemsInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MaxItems.IsSet = true
		s.MaxItems.Value = int(maxItemsInt)
	}

	minItemsValue, ok := input["minItems"]
	if ok {
		minItemsNumber, ok := minItemsValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		minItemsInt, rem := math.Modf(minItemsNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if minItemsInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MinItems.IsSet = true
		s.MinItems.Value = int(minItemsInt)
	}

	uniqueItemsValue, ok := input["uniqueItems"]
	if ok {
		uniqueItemsBool, ok := uniqueItemsValue.(bool)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		s.UniqueItems.IsSet = true
		s.UniqueItems.Value = uniqueItemsBool
	}

	containsValue, ok := input["contains"]
	if ok {
		containsObject, ok := containsValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("contains")

		subSchema, err := p.Parse(containsObject)
		if err != nil {
			return -1, err
		}

		s.Contains.IsSet = true
		s.Contains.Schema = subSchema

		p.Pop()
	}

	maxPropertiesValue, ok := input["maxProperties"]
	if ok {
		maxPropertiesNumber, ok := maxPropertiesValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		maxPropertiesInt, rem := math.Modf(maxPropertiesNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if maxPropertiesInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MaxProperties.IsSet = true
		s.MaxProperties.Value = int(maxPropertiesInt)
	}

	minPropertiesValue, ok := input["minProperties"]
	if ok {
		minPropertiesNumber, ok := minPropertiesValue.(float64)
		if !ok {
			return -1, ErrorInvalidSchema
		}

		minPropertiesInt, rem := math.Modf(minPropertiesNumber)
		if rem > epsilon {
			return -1, ErrorInvalidSchema
		}

		if minPropertiesInt < 0 {
			return -1, ErrorInvalidSchema
		}

		s.MinProperties.IsSet = true
		s.MinProperties.Value = int(minPropertiesInt)
	}

	requiredValue, ok := input["required"]
	if ok {
		requiredArray, ok := requiredValue.([]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		properties := []string{}
		for _, elem := range requiredArray {
			elemString, ok := elem.(string)
			if !ok {
				return -1, ErrorInvalidSchema
			}

			properties = append(properties, elemString)
		}

		s.Required.IsSet = true
		s.Required.Properties = properties
	}

	propertiesValue, ok := input["properties"]
	if ok {
		propertiesObject, ok := propertiesValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("properties")

		schemas := map[string]int{}
		for property, elem := range propertiesObject {
			elemObject, ok := elem.(map[string]interface{})
			if !ok {
				return -1, ErrorInvalidSchema
			}

			p.Push(property)
			subSchema, err := p.Parse(elemObject)
			if err != nil {
				return -1, err
			}

			schemas[property] = subSchema

			p.Pop()
		}

		s.Properties.IsSet = true
		s.Properties.Schemas = schemas

		p.Pop()
	}

	patternPropertiesValue, ok := input["patternProperties"]
	if ok {
		patternPropertiesObject, ok := patternPropertiesValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("patternProperties")

		schemas := map[*regexp.Regexp]int{}
		for property, elem := range patternPropertiesObject {
			elemObject, ok := elem.(map[string]interface{})
			if !ok {
				return -1, ErrorInvalidSchema
			}

			propertyRegexp, err := regexp.Compile(property)
			if err != nil {
				return -1, ErrorInvalidSchema
			}

			p.Push(property)
			subSchema, err := p.Parse(elemObject)
			if err != nil {
				return -1, err
			}

			schemas[propertyRegexp] = subSchema

			p.Pop()
		}

		s.PatternProperties.IsSet = true
		s.PatternProperties.Schemas = schemas

		p.Pop()
	}

	additionalPropertiesValue, ok := input["additionalProperties"]
	if ok {
		additionalPropertiesObject, ok := additionalPropertiesValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("additionalProperties")

		subSchema, err := p.Parse(additionalPropertiesObject)
		if err != nil {
			return -1, err
		}

		s.AdditionalProperties.IsSet = true
		s.AdditionalProperties.Schema = subSchema

		p.Pop()
	}

	dependenciesValue, ok := input["dependencies"]
	if ok {
		dependenciesObject, ok := dependenciesValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("dependencies")

		dependencies := map[string]schemaDependency{}
		for key, value := range dependenciesObject {
			p.Push(key)

			switch val := value.(type) {
			case map[string]interface{}:
				subSchema, err := p.Parse(val)
				if err != nil {
					return -1, err
				}

				dependencies[key] = schemaDependency{
					IsSchema: true,
					Schema:   subSchema,
				}
			case []interface{}:
				properties := []string{}
				for _, property := range val {
					propertyString, ok := property.(string)
					if !ok {
						return -1, ErrorInvalidSchema
					}

					properties = append(properties, propertyString)
				}

				dependencies[key] = schemaDependency{
					IsSchema:   false,
					Properties: properties,
				}
			default:
				return -1, ErrorInvalidSchema
			}

			p.Pop()
		}

		s.Dependencies.IsSet = true
		s.Dependencies.Deps = dependencies

		p.Pop()
	}

	propertyNamesValue, ok := input["propertyNames"]
	if ok {
		propertyNamesObject, ok := propertyNamesValue.(map[string]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("propertyNames")

		subSchema, err := p.Parse(propertyNamesObject)
		if err != nil {
			return -1, err
		}

		s.PropertyNames.IsSet = true
		s.PropertyNames.Schema = subSchema

		p.Pop()
	}

	allOfValue, ok := input["allOf"]
	if ok {
		allOfArray, ok := allOfValue.([]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("allOf")

		s.AllOf.IsSet = true
		s.AllOf.Schemas = make([]int, len(allOfArray))
		for i, schemaValue := range allOfArray {
			p.Push(strconv.FormatInt(int64(i), 10))

			schemaObject, ok := schemaValue.(map[string]interface{})
			if !ok {
				return -1, ErrorInvalidSchema
			}

			subSchema, err := p.Parse(schemaObject)
			if err != nil {
				return -1, err
			}

			s.AllOf.Schemas[i] = subSchema
			p.Pop()
		}

		p.Pop()
	}

	anyOfValue, ok := input["anyOf"]
	if ok {
		anyOfArray, ok := anyOfValue.([]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("anyOf")

		s.AnyOf.IsSet = true
		s.AnyOf.Schemas = make([]int, len(anyOfArray))
		for i, schemaValue := range anyOfArray {
			p.Push(strconv.FormatInt(int64(i), 10))

			schemaObject, ok := schemaValue.(map[string]interface{})
			if !ok {
				return -1, ErrorInvalidSchema
			}

			subSchema, err := p.Parse(schemaObject)
			if err != nil {
				return -1, err
			}

			s.AnyOf.Schemas[i] = subSchema
			p.Pop()
		}

		p.Pop()
	}

	oneOfValue, ok := input["oneOf"]
	if ok {
		oneOfArray, ok := oneOfValue.([]interface{})
		if !ok {
			return -1, ErrorInvalidSchema
		}

		p.Push("oneOf")

		s.OneOf.IsSet = true
		s.OneOf.Schemas = make([]int, len(oneOfArray))
		for i, schemaValue := range oneOfArray {
			p.Push(strconv.FormatInt(int64(i), 10))

			schemaObject, ok := schemaValue.(map[string]interface{})
			if !ok {
				return -1, ErrorInvalidSchema
			}

			subSchema, err := p.Parse(schemaObject)
			if err != nil {
				return -1, err
			}

			s.OneOf.Schemas[i] = subSchema
			p.Pop()
		}

		p.Pop()
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
		return 0, ErrorInvalidSchema
	}
}
