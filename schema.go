package jsonschema

// Schema is a JSON Schema schema. Instances of Schema can be unmarshalled from
// JSON.
type Schema struct {
	Type  *SchemaType  `json:"type"`
	Items *SchemaItems `json:"items"`
}

// SchemaType is either one JSONType or a nonempty list of JSONType.
type SchemaType struct {
	IsSingle bool
	Single   JSONType
	List     []JSONType
}

func (t *SchemaType) contains(typ JSONType) bool {
	if t.IsSingle {
		return t.Single == typ
	}

	for _, elem := range t.List {
		if elem == typ {
			return true
		}
	}

	return false
}

func (t *SchemaType) UnmarshalJSON(data []byte) error {
	var single JSONType
	var list []JSONType

	isSingle, err := unmarshalWithFallback(data, &single, &list)
	if err != nil {
		return err
	}

	t.IsSingle = isSingle
	t.Single = single
	t.List = list
	return nil
}

// SchemaItems is either one Schema or a nonempty list of Schemas.
type SchemaItems struct {
	IsSingle bool
	Single   Schema
	List     []Schema
}

func (t *SchemaItems) UnmarshalJSON(data []byte) error {
	var single Schema
	var list []Schema

	isSingle, err := unmarshalWithFallback(data, &single, &list)
	if err != nil {
		return err
	}

	t.IsSingle = isSingle
	t.Single = single
	t.List = list
	return nil
}

type JSONType string

const (
	JSONTypeNull    JSONType = "null"
	JSONTypeBoolean          = "boolean"
	JSONTypeNumber           = "number"
	JSONTypeInteger          = "integer"
	JSONTypeString           = "string"
	JSONTypeArray            = "array"
	JSONTypeObject           = "object"
)
