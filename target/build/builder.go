package build

import (
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/hack-pad/hackpadfs"
	"github.com/portal-co/boogie/hashmap"
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/boogie/stache"
	"github.com/portal-co/boogie/target"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type Builder struct {
	stache.Cache
	sandbox.Runner
	Targets hashmap.HashMap[target.ConfiguredLabel, map[string]func(target.ConfiguredLabel) (*Target, error)]
}

type Target struct {
	Default string
	In      target.ConfiguredLabel
}

func (t *Target) String() string {
	return fmt.Sprintf("%s{default: %s}", t.In, t.Default)
}

func (t *Target) Type() string {
	return "target"
}

func (t *Target) Freeze() {

}

func (t *Target) Truth() starlark.Bool {
	return true
}

func (t *Target) Hash() (uint32, error) {
	s := sha256.Sum256([]byte(t.String()))
	f := fnv.New32()
	f.Write(s[:])
	return f.Sum32(), nil
}

func (b *Builder) Build(c target.ConfiguredLabel) (*Target, error) {
	switch l := c.Name.(type) {
	case target.BakedLabel:
		return &Target{string(l), c}, nil
	case target.IpfsLabel:
		return &Target{string(l), c}, nil
	case target.DelveLabel:
		_, err := b.Load(l.Internal.String() + "/BUILD")
		if err != nil {
			return nil, err
		}
		t, ok := b.Targets.Get(l.Internal)
		if !ok {
			return nil, fmt.Errorf("not found")
		}
		var v *Target
		p := ""
		for _, a := range strings.Split(l.Path, "/") {
			p += a
			u, ok := t[p]
			if !ok {
				return nil, fmt.Errorf("not found")
			}
			v, err = u(c)
			if err != nil {
				return nil, err
			}
			p += "/"
		}
		return v, nil
	default:
	}
	return nil, fmt.Errorf("invalid label")
}

func (b *Builder) File(x string) (string, error) {
	var c target.ConfiguredLabel
	c.Parse(&x)
	bd, err := b.Build(c)
	if err != nil {
		return "", err
	}
	y, err := hackpadfs.ReadFile(b.Ipfs(), bd.Default)
	if err != nil {
		return "", err
	}
	return string(y), nil
}
func (b *Builder) Global(x string) starlark.StringDict {
	d := starlark.StringDict{}
	d["struct"] = starlark.NewBuiltin("struct", starlarkstruct.Make)
	d["label"] = starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"cfg": starlark.NewBuiltin("cfg", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var t, c string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &t, "cfg_key", &c); err != nil {
				return nil, err
			}
			var tc target.ConfiguredLabel
			var cc target.ConfiguredLabel
			tc.Parse(&t)
			cc.Parse(&c)
			x, ok := hashmap.HashMap[target.ConfiguredLabel, target.CfgEntry](tc.Cfg).Get(cc)
			if !ok {
				return nil, fmt.Errorf("not found")
			}
			return starlark.String(x.String()), nil
		}),
		"resolve": starlark.NewBuiltin("resolve", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var t string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &t); err != nil {
				return nil, err
			}
			var tc target.ConfiguredLabel
			tc.Parse(&t)
			return starlark.String(tc.String()), nil
		}),
		"build": starlark.NewBuiltin("build", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var t string
			if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &t); err != nil {
				return nil, err
			}
			var c target.ConfiguredLabel
			c.Parse(&t)
			bd, err := b.Build(c)
			if err != nil {
				return nil, err
			}
			return starlark.String(bd.Default), nil
		}),
	})
	d["path"] = starlark.String(x)
	d["emit"] = starlark.NewBuiltin("emit", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var c target.ConfiguredLabel
		c.Parse(&x)
		var t string
		var impl starlark.Value
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "target", &t, "impl", &impl); err != nil {
			return nil, err
		}
		tm, ok := b.Targets.Get(c)
		if !ok {
			tm = map[string]func(target.ConfiguredLabel) (*Target, error){}
			b.Targets.Put(c, tm)
		}
		_, ok = tm[t]
		if ok {
			return starlark.None, nil
		}
		tm[t] = func(cl target.ConfiguredLabel) (*Target, error) {
			x, err := starlark.Call(thread, impl, starlark.Tuple{starlark.String(cl.String())}, []starlark.Tuple{})
			if err != nil {
				return nil, err
			}
			d, err := x.(starlark.HasAttrs).Attr("default")
			if err != nil {
				return nil, err
			}
			return &Target{
				In:      cl,
				Default: string(d.(starlark.String)),
			}, nil
		}
		return starlark.None, nil
	})
	return d
}
