package jsonschema

type arena struct {
	schemas []schema
}

func newArena(cap int) arena {
	return arena{schemas: make([]schema, 0, cap)}
}

func (a *arena) Insert(s schema) int {
	a.schemas = append(a.schemas, s)
	return len(a.schemas) - 1
}

func (a *arena) Get(i int) schema {
	return a.schemas[i]
}
