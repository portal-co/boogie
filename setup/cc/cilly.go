package cc

import "github.com/portal-co/boogie/sandbox"

type Cilly struct {
	Cilly string
}

// CC implements CC.
func (c Cilly) CC(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	d := map[string]string{"cilly": c.Cilly}
	// if !cpp {
	d["target.c"] = src
	// } else {
	// 	d["target.cc"] = src
	// }
	m := []string{"./cilly", "--merge"}
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

// Link implements CC.
func (c Cilly) Link(r sandbox.Runner, libs []string, objs []string) (string, error) {
	d := map[string]string{"cilly": c.Cilly}
	m := []string{"./cilly", "--merge", "--keepmerged"}
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

var _ CC = Cilly{}
