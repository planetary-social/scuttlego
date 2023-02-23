package internal

type Queue[T any] struct {
	queue []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Enqueue(v T) {
	q.queue = append(q.queue, v)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.queue) == 0 {
		return *new(T), false
	}

	v := q.queue[0]
	q.queue = q.queue[1:]
	return v, true
}
