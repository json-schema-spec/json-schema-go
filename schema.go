package jsonschema

type Schema struct {
	Type SchemaType `json:"type"`
}

// SchemaType is either one JSONType or a nonempty list of JSONType.
type SchemaType struct {
	IsSingle bool
	Single   JSONType
	List     []JSONType
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
