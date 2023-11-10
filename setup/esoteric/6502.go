package esoteric

import (
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/boogie/setup/cc"
)

type Mos struct {
	Type string
	SDK  string
}

// CXX implements cc.CXX.
func (mos Mos) CXX(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	d := map[string]string{"sdk": mos.SDK}
	// if !cpp {
	d["target.cc"] = src
	// } else {
	// 	d["target.cc"] = src
	// }
	m := []string{"./sdk/bin/mos-" + mos.Type + "-clang++"}
	// if cpp {
	// 	m = append(m, "c++", "target.cc")
	// } else {
	m = append(m, "target.cc")
	// }
	m = append(m, flags...)
	m = append(m, "-o", "target.o", "-c")
	for _, h := range hdrs {
		d[h] = h
		m = append(m, "-isystem", "./"+h)
	}
	s, err := r.Run(d, m, []string{}, "/target.o")
	return s, err
}

// CC implements cc.CC.
func (mos Mos) CC(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	d := map[string]string{"sdk": mos.SDK}
	// if !cpp {
	d["target.c"] = src
	// } else {
	// 	d["target.cc"] = src
	// }
	m := []string{"./sdk/bin/mos-" + mos.Type + "-clang"}
	// if cpp {
	// 	m = append(m, "c++", "target.cc")
	// } else {
	m = append(m, "target.c")
	// }
	m = append(m, flags...)
	m = append(m, "-o", "target.o", "-c")
	for _, h := range hdrs {
		d[h] = h
		m = append(m, "-isystem", "./"+h)
	}
	s, err := r.Run(d, m, []string{}, "/target.o")
	return s, err
}

// Link implements cc.CC.
func (mos Mos) Link(r sandbox.Runner, libs []string, objs []string) (string, error) {
	d := map[string]string{"sdk": mos.SDK}
	m := []string{"./sdk/bin/mos-" + mos.Type + "-clang"}
	for _, l := range libs {
		m = append(m, "-l"+l)
	}
	for _, o := range objs {
		d[o+".o"] = o
		m = append(m, "./"+o+".o")
	}
	s, err := r.Run(d, m, []string{}, "/target")
	return s, err
}

var _ cc.CXX = Mos{}
