package jsonschema

import (
	"errors"
	"math"
	"net/url"
	"regexp"
	"strconv"

	jsonpointer "github.com/json-schema-spec/json-pointer-go"
)

type parser struct {
	registry *registry
	baseURI  url.URL
	tokens   []string
}

func parseRootSchema(registry *registry, input interface{}) (schema, error) {
	return parseSubSchema(registry, url.URL{}, []string{}, input)
}

func parseSubSchema(registry *registry, baseURI url.URL, tokens []string, input interface{}) (schema, error) {
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

func (p *parser) Parse(input interface{}) (int, error) {
	s := schema{}

	switch input := input.(type) {
	case bool:
		s.Bool.IsSet = true
		s.Bool.Value = input
	case map[string]interface{}:
		if len(p.tokens) == 0 {
			idValue, ok := input["$id"]
			if ok {
				idStr, ok := idValue.(string)
				if !ok {
					return -1, ErrInvalidSchema
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
				return -1, ErrInvalidSchema
			}

			uri, err := p.baseURI.Parse(refStr)
			if err != nil {
				return -1, ErrInvalidSchema
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
			p.Push("not")

			subSchema, err := p.Parse(notValue)
			if err != nil {
				return -1, err
			}

			s.Not.IsSet = true
			s.Not.Schema = subSchema

			p.Pop()
		}

		ifValue, ok := input["if"]
		if ok {
			p.Push("if")

			subSchema, err := p.Parse(ifValue)
			if err != nil {
				return -1, err
			}

			s.If.IsSet = true
			s.If.Schema = subSchema

			p.Pop()
		}

		thenValue, ok := input["then"]
		if ok {
			p.Push("then")

			subSchema, err := p.Parse(thenValue)
			if err != nil {
				return -1, err
			}

			s.Then.IsSet = true
			s.Then.Schema = subSchema

			p.Pop()
		}

		elseValue, ok := input["else"]
		if ok {
			p.Push("else")

			subSchema, err := p.Parse(elseValue)
			if err != nil {
				return -1, err
			}

			s.Else.IsSet = true
			s.Else.Schema = subSchema

			p.Pop()
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
						return -1, ErrInvalidSchema
					}

					jsonTyp, err := parseJSONType(t)
					if err != nil {
						return -1, err
					}

					s.Type.Types[i] = jsonTyp
				}
			default:
				return -1, ErrInvalidSchema
			}
		}

		itemsValue, ok := input["items"]
		if ok {
			switch items := itemsValue.(type) {
			case []interface{}:
				p.Push("items")

				s.Items.IsSet = true
				s.Items.IsSingle = false
				s.Items.Schemas = make([]int, len(items))

				for i, item := range items {
					p.Push(strconv.FormatInt(int64(i), 10))

					subSchema, err := p.Parse(item)
					if err != nil {
						return -1, err
					}

					s.Items.Schemas[i] = subSchema
					p.Pop()
				}

				p.Pop()
			default:
				p.Push("items")

				subSchema, err := p.Parse(items)
				if err != nil {
					return -1, err
				}

				s.Items.IsSet = true
				s.Items.IsSingle = true
				s.Items.Schemas = []int{subSchema}

				p.Pop()
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
				return -1, ErrInvalidSchema
			}

			s.Enum.IsSet = true
			s.Enum.Values = enumArray
		}

		multipleOfValue, ok := input["multipleOf"]
		if ok {
			multipleOfNumber, ok := multipleOfValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.MultipleOf.IsSet = true
			s.MultipleOf.Value = multipleOfNumber
		}

		maximumValue, ok := input["maximum"]
		if ok {
			maximumNumber, ok := maximumValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.Maximum.IsSet = true
			s.Maximum.Value = maximumNumber
		}

		minimumValue, ok := input["minimum"]
		if ok {
			minimumNumber, ok := minimumValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.Minimum.IsSet = true
			s.Minimum.Value = minimumNumber
		}

		exclusiveMaximumValue, ok := input["exclusiveMaximum"]
		if ok {
			exclusiveMaximumNumber, ok := exclusiveMaximumValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.ExclusiveMaximum.IsSet = true
			s.ExclusiveMaximum.Value = exclusiveMaximumNumber
		}

		exclusiveMinimumValue, ok := input["exclusiveMinimum"]
		if ok {
			exclusiveMinimumNumber, ok := exclusiveMinimumValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.ExclusiveMinimum.IsSet = true
			s.ExclusiveMinimum.Value = exclusiveMinimumNumber
		}

		maxLengthValue, ok := input["maxLength"]
		if ok {
			maxLengthNumber, ok := maxLengthValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			maxLengthInt, rem := math.Modf(maxLengthNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if maxLengthInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MaxLength.IsSet = true
			s.MaxLength.Value = int(maxLengthInt)
		}

		minLengthValue, ok := input["minLength"]
		if ok {
			minLengthNumber, ok := minLengthValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			minLengthInt, rem := math.Modf(minLengthNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if minLengthInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MinLength.IsSet = true
			s.MinLength.Value = int(minLengthInt)
		}

		patternValue, ok := input["pattern"]
		if ok {
			patternString, ok := patternValue.(string)
			if !ok {
				return -1, ErrInvalidSchema
			}

			patternRegexp, err := regexp.Compile(patternString)
			if err != nil {
				return -1, ErrInvalidSchema
			}

			s.Pattern.IsSet = true
			s.Pattern.Value = patternRegexp
		}

		additionalItemsValue, ok := input["additionalItems"]
		if ok {
			p.Push("additionalItems")

			subSchema, err := p.Parse(additionalItemsValue)
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
				return -1, ErrInvalidSchema
			}

			maxItemsInt, rem := math.Modf(maxItemsNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if maxItemsInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MaxItems.IsSet = true
			s.MaxItems.Value = int(maxItemsInt)
		}

		minItemsValue, ok := input["minItems"]
		if ok {
			minItemsNumber, ok := minItemsValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			minItemsInt, rem := math.Modf(minItemsNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if minItemsInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MinItems.IsSet = true
			s.MinItems.Value = int(minItemsInt)
		}

		uniqueItemsValue, ok := input["uniqueItems"]
		if ok {
			uniqueItemsBool, ok := uniqueItemsValue.(bool)
			if !ok {
				return -1, ErrInvalidSchema
			}

			s.UniqueItems.IsSet = true
			s.UniqueItems.Value = uniqueItemsBool
		}

		containsValue, ok := input["contains"]
		if ok {
			p.Push("contains")

			subSchema, err := p.Parse(containsValue)
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
				return -1, ErrInvalidSchema
			}

			maxPropertiesInt, rem := math.Modf(maxPropertiesNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if maxPropertiesInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MaxProperties.IsSet = true
			s.MaxProperties.Value = int(maxPropertiesInt)
		}

		minPropertiesValue, ok := input["minProperties"]
		if ok {
			minPropertiesNumber, ok := minPropertiesValue.(float64)
			if !ok {
				return -1, ErrInvalidSchema
			}

			minPropertiesInt, rem := math.Modf(minPropertiesNumber)
			if rem > Epsilon {
				return -1, ErrInvalidSchema
			}

			if minPropertiesInt < 0 {
				return -1, ErrInvalidSchema
			}

			s.MinProperties.IsSet = true
			s.MinProperties.Value = int(minPropertiesInt)
		}

		requiredValue, ok := input["required"]
		if ok {
			requiredArray, ok := requiredValue.([]interface{})
			if !ok {
				return -1, ErrInvalidSchema
			}

			properties := []string{}
			for _, elem := range requiredArray {
				elemString, ok := elem.(string)
				if !ok {
					return -1, ErrInvalidSchema
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
				return -1, ErrInvalidSchema
			}

			p.Push("properties")

			schemas := map[string]int{}
			for property, elem := range propertiesObject {
				p.Push(property)
				subSchema, err := p.Parse(elem)
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
				return -1, ErrInvalidSchema
			}

			p.Push("patternProperties")

			schemas := map[*regexp.Regexp]int{}
			for property, elem := range patternPropertiesObject {
				propertyRegexp, err := regexp.Compile(property)
				if err != nil {
					return -1, ErrInvalidSchema
				}

				p.Push(property)
				subSchema, err := p.Parse(elem)
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
			p.Push("additionalProperties")

			subSchema, err := p.Parse(additionalPropertiesValue)
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
				return -1, ErrInvalidSchema
			}

			p.Push("dependencies")

			dependencies := map[string]schemaDependency{}
			for key, value := range dependenciesObject {
				p.Push(key)

				switch val := value.(type) {
				case []interface{}:
					properties := []string{}
					for _, property := range val {
						propertyString, ok := property.(string)
						if !ok {
							return -1, ErrInvalidSchema
						}

						properties = append(properties, propertyString)
					}

					dependencies[key] = schemaDependency{
						IsSchema:   false,
						Properties: properties,
					}
				default:
					subSchema, err := p.Parse(val)
					if err != nil {
						return -1, err
					}

					dependencies[key] = schemaDependency{
						IsSchema: true,
						Schema:   subSchema,
					}
				}

				p.Pop()
			}

			s.Dependencies.IsSet = true
			s.Dependencies.Deps = dependencies

			p.Pop()
		}

		propertyNamesValue, ok := input["propertyNames"]
		if ok {
			p.Push("propertyNames")

			subSchema, err := p.Parse(propertyNamesValue)
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
				return -1, ErrInvalidSchema
			}

			p.Push("allOf")

			s.AllOf.IsSet = true
			s.AllOf.Schemas = make([]int, len(allOfArray))
			for i, schemaValue := range allOfArray {
				p.Push(strconv.FormatInt(int64(i), 10))

				subSchema, err := p.Parse(schemaValue)
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
				return -1, ErrInvalidSchema
			}

			p.Push("anyOf")

			s.AnyOf.IsSet = true
			s.AnyOf.Schemas = make([]int, len(anyOfArray))
			for i, schemaValue := range anyOfArray {
				p.Push(strconv.FormatInt(int64(i), 10))

				subSchema, err := p.Parse(schemaValue)
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
				return -1, ErrInvalidSchema
			}

			p.Push("oneOf")

			s.OneOf.IsSet = true
			s.OneOf.Schemas = make([]int, len(oneOfArray))
			for i, schemaValue := range oneOfArray {
				p.Push(strconv.FormatInt(int64(i), 10))

				subSchema, err := p.Parse(schemaValue)
				if err != nil {
					return -1, err
				}

				s.OneOf.Schemas[i] = subSchema
				p.Pop()
			}

			p.Pop()
		}
	default:
		return -1, ErrInvalidSchema
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
		return 0, ErrInvalidSchema
	}
}
