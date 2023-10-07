package cell

import (
	"context"

	"github.com/portal-co/boogie/eo"
)

type Cell[T any] chan T

func (c Cell[T]) Put(v func() T) {
	go func() { c <- v() }()
}
func (c Cell[T]) Get() T {
	x := <-c
	go func() { c <- x }()
	return x
}
func (c Cell[T]) GetWithCtx(ctx context.Context) (T, error) {
	select {
	case x := <-c:
		go func() { c <- x }()
		return x, nil
	case <-ctx.Done():
		var n T
		return n, ctx.Err()
	}
}
func GetWithErrorContext[T any](ctx context.Context, x Cell[eo.ErrorOr[T]]) (T, error) {
	c, err := x.GetWithCtx(ctx)
	if err != nil {
		return c.Value, err
	}
	return c.Unwrap()
}
