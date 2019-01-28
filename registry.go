package jsonschema

import (
	"net/url"
)

type registry struct {
	schemas map[url.URL]int
	arena   arena
}

func newRegistry(cap int) registry {
	return registry{schemas: map[url.URL]int{}, arena: newArena(cap)}
}

func (r *registry) Get(uri url.URL) (schema, bool) {
	index, ok := r.schemas[uri]
	return r.arena.Get(index), ok
}

func (r *registry) GetIndex(index int) schema {
	return r.arena.Get(index)
}

func (r *registry) Insert(uri url.URL, s schema) int {
	if index, ok := r.schemas[uri]; ok {
		return index
	}

	index := r.arena.Insert(s)
	r.schemas[uri] = index
	return index
}

func (r *registry) PopulateRefs() []url.URL {
	missing := []url.URL{}

	for index, schema := range r.arena.schemas {
		if !schema.Ref.IsSet {
			continue
		}

		if refIndex, ok := r.schemas[schema.Ref.URI]; ok {
			schema.Ref.Schema = refIndex
			r.arena.schemas[index] = schema
		} else {
			missing = append(missing, schema.Ref.URI)
		}
	}

	return missing
}
