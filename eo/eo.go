package eo

type ErrorOr[T any] struct {
	Value T
	Err   error
}

func New[T any](v T, err error) ErrorOr[T] {
	return ErrorOr[T]{v, err}
}
func (x ErrorOr[T]) Unwrap() (T, error) {
	return x.Value, x.Err
}
