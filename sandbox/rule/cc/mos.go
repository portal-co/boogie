package cc

import (
	"github.com/portal-co/boogie/sandbox"
)

type MosCC struct {
	Prefix string
	Bin    string
}

func (z MosCC) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	d := map[string]string{"src": src}
	b := "mos-" + z.Prefix + "-clang"
	if z.Bin != "" {
		o := b
		b = "./" + b
		d[o] = z.Bin
	}
	c := []string{"/usr/bin/env", b, "-x", lang, "./src", "-c", "-o", "./out.o"}
	for _, h := range hdrs {
		d["include/"+h] = h
		c = append(c, "-I", "include/"+h)
	}
	s, err := st.Run(d, c, []string{"out.o"})
	return s + "/out.o", err
}
func (z MosCC) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	d := map[string]string{}
	b := "mos-" + z.Prefix + "-clang"
	if z.Bin != "" {
		o := b
		b = "./" + b
		d[o] = z.Bin
	}
	c := []string{"/usr/bin/env", b, "-o", "./out"}
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
