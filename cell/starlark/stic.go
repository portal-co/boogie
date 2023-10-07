package starlark

import (
	"fmt"

	"github.com/portal-co/boogie/cell"
	"github.com/portal-co/boogie/eo"
	"go.starlark.net/starlark"
)

type Stic struct {
	cell.Cell[eo.ErrorOr[starlark.Value]]
}

// Attr implements starlark.HasAttrs.
func (s Stic) Attr(name string) (starlark.Value, error) {
	switch name {
	case "get":
		return s.Cell.Get().Unwrap()
	case "put":
		return starlark.NewBuiltin("put", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var c starlark.Callable
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "c", &c); err != nil {
				return nil, err
			}
			s.Cell.Put(func() eo.ErrorOr[starlark.Value] {
				x, err := starlark.Call(thread, c, []starlark.Value{}, []starlark.Tuple{})
				return eo.New[starlark.Value](x, err)
			})
			return starlark.None, nil
		}), nil
	}
	return nil, fmt.Errorf("not supported")
}

// AttrNames implements starlark.HasAttrs.
func (Stic) AttrNames() []string {
	return []string{"get", "put"}
}

// Freeze implements starlark.Value.
func (Stic) Freeze() {
	// panic("unimplemented")
}

// Hash implements starlark.Value.
func (Stic) Hash() (uint32, error) {
	// panic("unimplemented")
	return 0, fmt.Errorf("not supported")
}

// String implements starlark.Value.
func (Stic) String() string {
	return "cell"
}

// Truth implements starlark.Value.
func (Stic) Truth() starlark.Bool {
	return true
}

// Type implements starlark.Value.
func (Stic) Type() string {
	return "cell"
}

var _ starlark.Value = Stic{}
var _ starlark.HasAttrs = Stic{}
