package jsonschema

type stack struct {
	elems []string
}

func (s *stack) Push(token string) {
	s.elems = append(s.elems, token)
}

func (s *stack) Pop() string {
	if len(s.elems) == 0 {
		panic("stack underflow")
	}

	var last string
	s.elems, last = s.elems[:len(s.elems)-1], s.elems[len(s.elems)-1]
	return last
}
