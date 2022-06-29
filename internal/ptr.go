package internal

func Ptr[T any](o T) *T {
	return &o
}
