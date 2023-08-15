package cc

import (
	"github.com/portal-co/boogie/sandbox"
)

type W2c2 struct {
	Bin string
}
type CompilerWrapper interface {
	Wrap(st sandbox.Runner, cc CC, f func(CC) (string, error), name string) (string, string, error)
}

func (c W2c2) Wrap(st sandbox.Runner, cc CC, f func(CC) (string, error), name string) (string, string, error) {
	d := map[string]string{}
	b := "w2c2"
	if c.Bin != "" {
		b = "./w2c2"
		d["wc2"] = c.Bin
	}
	l, err := f(ZigCC{Ztarget: "wasm32-wasi"})
	if err != nil {
		return "", "", err
	}
	d["input.wasm"] = l
	m := []string{b, "./input.wasm", name + ".c"}
	x, err := st.Run(d, m, []string{name + ".c"})
	if err != nil {
		return "", "", err
	}
	r, err := cc.Compile(st, x+"/"+name+".c", []string{x}, []string{}, "c")
	return r, x, err
}
