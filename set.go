package main

type Set[T comparable] map[T]struct{}

func (s Set[T]) New() map[T]struct{} {
	return make(map[T]struct{})
}

func (s Set[T]) Has(element T) bool {
	_, ok := s[element]
	return ok
}

func (s Set[T]) Remove(element T) {
	delete(s, element)
}

func (s Set[T]) Add(element T) {
	s[element] = struct{}{}
}
