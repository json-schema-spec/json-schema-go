package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := stack{elems: []string{}}
	s.Push("foo")
	s.Push("bar")
	assert.Equal(t, "bar", s.Pop())
	s.Push("baz")
	assert.Equal(t, "baz", s.Pop())
	assert.Equal(t, "foo", s.Pop())
}
