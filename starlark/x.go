package starlark

import (
	"github.com/portal-co/boogie/cor"
	"go.starlark.net/starlark"
)

type WrapIteratorZ struct {
	Nxt func() (starlark.Value, bool)
	Don func()
}

func (w WrapIteratorZ) Next(p *starlark.Value) bool {
	x, b := w.Nxt()
	if !b {
		return b
	}
	*p = x
	return b
}
func (w WrapIteratorZ) Done() {
	w.Don()
}
func WrapIterator(fun func(func(starlark.Value) bool)) WrapIteratorZ {
	a, b := cor.Pull(fun)
	return WrapIteratorZ{a, b}
}
func UnwrapIterator(y func() starlark.Iterator) func(func(starlark.Value) bool) {
	return func(f func(starlark.Value) bool) {
		x := y()
		cor.Push[starlark.Value](func() (starlark.Value, bool) {
			var v starlark.Value
			b := x.Next(&v)
			return v, b
		}, func() {
			x.Done()
		})(f)
	}
}
