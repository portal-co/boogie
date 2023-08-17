package cor

import (
	"errors"
	"fmt"
)

var ErrCanceled = errors.New("coroutine canceled")

func New[In, Out any](f func(in In, yield func(Out) In) Out) (resume func(In) (Out, bool), cancel func()) {
	cin := make(chan msg[In])
	cout := make(chan msg[Out])
	running := true
	resume = func(in In) (out Out, ok bool) {
		if !running {
			return
		}
		cin <- msg[In]{val: in}
		m := <-cout
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val, running
	}
	cancel = func() {
		e := fmt.Errorf("%w", ErrCanceled) // unique wrapper
		cin <- msg[In]{panic: e}
		m := <-cout
		if m.panic != nil && m.panic != e {
			panic(m.panic)
		}
	}
	yield := func(out Out) In {
		cout <- msg[Out]{val: out}
		m := <-cin
		if m.panic != nil {
			panic(m.panic)
		}
		return m.val
	}
	go func() {
		defer func() {
			if running {
				running = false
				cout <- msg[Out]{panic: recover()}
			}
		}()
		var out Out
		m := <-cin
		if m.panic == nil {
			out = f(m.val, yield)
		}
		running = false
		cout <- msg[Out]{val: out}
	}()
	return resume, cancel
}

type msg[T any] struct {
	panic any
	val   T
}

func Pull[V any](push func(yield func(V) bool)) (pull func() (V, bool), stop func()) {
	copush := func(more bool, yield func(V) bool) V {
		if more {
			push(yield)
		}
		var zero V
		return zero
	}
	resume, _ := New(copush)
	pull = func() (V, bool) {
		return resume(true)
	}
	stop = func() {
		resume(false)
	}
	return pull, stop
}
func Push[V any](get func() (V, bool), stop func()) func(yield func(V) bool) {
	return func(yield func(V) bool) {
		defer stop()
		b := true
		for b {
			var g V
			g, b = get()
			if b {
				y := yield(g)
				if !y {
					return
				}
			}
		}
	}
}
