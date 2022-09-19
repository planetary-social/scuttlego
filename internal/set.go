package internal

type Set[T comparable] struct {
	values map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{
		values: make(map[T]struct{}),
	}
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s.values[v]
	return ok
}

func (s *Set[T]) Put(v T) {
	s.values[v] = struct{}{}
}

func (s Set[T]) List() []T {
	var result []T
	for v := range s.values {
		result = append(result, v)
	}
	return result
}

func (s Set[T]) Len() int {
	return len(s.values)
}
