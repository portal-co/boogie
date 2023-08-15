package cc

import (
	"fmt"
	"strings"

	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/remount"
	"golang.org/x/sync/errgroup"
)

type LLVMTarget interface {
	Target() string
}
type CC interface {
	Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error)
	Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error)
}

type Inject struct {
	C          CC
	ExtraHdrs  []string
	ExtraObjs  []string
	ExtraFlags []string
}

func (i Inject) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	return i.C.Compile(st, src, append(i.ExtraHdrs, hdrs...), append(i.ExtraFlags, flags...), lang)
}
func (i Inject) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	return i.C.Link(st, sum, append(i.ExtraObjs, objs...), combine)
}

var _ CC = Inject{}

type MapCC map[string]CC

func (m MapCC) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	n := map[string]string{}
	var g errgroup.Group
	for k, v := range m {
		k := k
		v := v
		g.Go(func() error {
			s, err := v.Compile(st, src, hdrs, flags, lang)
			if err != nil {
				return err
			}
			n[k] = s
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return "", err
	}
	return remount.NewDir(st.Ipfs(), n)
}
func (m MapCC) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	n := map[string]string{}
	var g errgroup.Group
	for k, v := range m {
		k := k
		v := v
		o := []string{}
		for _, j := range objs {
			o = append(o, j+"/"+k)
		}
		g.Go(func() error {
			s, err := v.Link(st, sum, o, combine)
			if err != nil {
				return err
			}
			n[k] = s
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		return "", err
	}
	return remount.NewDir(st.Ipfs(), n)
}

var _ CC = MapCC{}

type ZigCC struct {
	Ztarget string
	Bin     string
}

func (z ZigCC) Target() string {
	return z.Ztarget
}
func (z ZigCC) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	d := map[string]string{"src": src}
	b := "zig"
	if z.Bin != "" {
		b = "./zig"
		d["zig"] = z.Bin
	}
	c := []string{"/usr/bin/env", b, strings.TrimPrefix(lang, "objective-"), "-target", z.Ztarget, "-x", lang, "./src", "-c", "-o", "./out.o"}
	for _, h := range hdrs {
		d["include/"+h] = h
		c = append(c, "-I", "include/"+h)
	}
	s, err := st.Run(d, c, []string{"out.o"})
	return s + "/out.o", err
}
func (z ZigCC) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	d := map[string]string{}
	b := "zig"
	if z.Bin != "" {
		b = "./zig"
		d["zig"] = z.Bin
	}
	c := []string{"/usr/bin/env", b, strings.TrimPrefix(sum, "objective-"), "-target", z.Ztarget, "-o", "./out"}
	if combine {
		c = append(c, "-Wl,-Ur")
	}
	for _, h := range objs {
		d["obj/"+h+".o"] = h
		c = append(c, "obj/"+h+".o")
	}
	s, err := st.Run(d, c, []string{"out"})
	return s + "/out", err
}

var _ LLVMTarget = ZigCC{}
var _ CC = ZigCC{}

type SingleC interface {
	CToTarget(st sandbox.Runner, c string) (string, error)
}

type SingleCWrapper struct {
	CC
	Runtime []string
}

func (w SingleCWrapper) CToTarget(st sandbox.Runner, c string) (string, error) {
	c, err := w.Compile(st, c, []string{}, []string{}, "c")
	if err != nil {
		return "", err
	}
	return w.Link(st, "c", append([]string{c}, w.Runtime...), false)
}

var _ SingleC = SingleCWrapper{}

type Cilly struct {
	Bin string
	S   SingleC
}

func (c Cilly) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	if lang != "c" {
		return "", fmt.Errorf("Cilly only supports C")
	}
	d := map[string]string{}
	b := "cilly"
	if c.Bin != "" {
		b = "./cilly"
		d["cilly"] = c.Bin
	}
	m := []string{"/usr/bin/env", b, strings.TrimPrefix(lang, "objective-"), "--merge", "./src", "-c", "-o", "./out.c"}
	for _, h := range hdrs {
		d["include/"+h] = h
		m = append(m, "-I", "include/"+h)
	}
	s, err := st.Run(d, m, []string{"out.c"})
	return s + "/out.c", err
}
func (c Cilly) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	if sum != "c" {
		return "", fmt.Errorf("Cilly only supports C")
	}
	d := map[string]string{}
	b := "cilly"
	if c.Bin != "" {
		b = "./cilly"
		d["cilly"] = c.Bin
	}
	m := []string{"/usr/bin/env", b, strings.TrimPrefix(sum, "objective-"), "--merge", "-o", "./out.c"}
	for _, h := range objs {
		d["obj/"+h+".c"] = h
		m = append(m, "obj/"+h+".c")
	}
	s, err := st.Run(d, m, []string{"out.c"})
	if err != nil {
		return "", err
	}
	if c.S == nil {
		return s + "/out.c", err
	}
	return c.S.CToTarget(st, s+"/out.c")
}

var _ CC = Cilly{}
