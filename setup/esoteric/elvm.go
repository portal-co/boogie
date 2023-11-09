package esoteric

import (
	"fmt"

	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/boogie/setup/cc"
)

type ELVM struct {
	ECC, Elc, Target string
}

// CC implements cc.CC.
func (e ELVM) CC(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	d := map[string]string{"8cc": e.ECC}
	// if !cpp {
	d["target.c"] = src
	// } else {
	// 	d["target.cc"] = src
	// }
	m := []string{"./8cc"}
	// if cpp {
	// 	m = append(m, "c++", "target.cc")
	// } else {
	m = append(m, "target.c")
	// }
	m = append(m, flags...)
	m = append(m, "-o", "target.eir")
	for _, h := range hdrs {
		d[h] = h
		m = append(m, "-isystem", "./"+h)
	}
	s, err := r.Run(d, m, []string{})
	if err != nil {
		return "", err
	}
	d = map[string]string{"elc": e.Elc, "target.eir": s + "/target.eir"}
	m = []string{"./elc", "./target.eir", "-target", e.Target, "-o", "./target"}
	s, err = r.Run(d, m, []string{})
	if err != nil {
		return "", err
	}
	return s + "/target", nil
}

var ErrElvmLib = fmt.Errorf("libraries are not supported in elvm")
var ErrElvmMultiSrc = fmt.Errorf("elvm can only compile one source file")

// Link implements cc.CC.
func (ELVM) Link(r sandbox.Runner, libs []string, objs []string) (string, error) {
	if len(libs) != 0 {
		return "", ErrElvmLib
	}
	if len(objs) != 1 {
		return "", ErrElvmMultiSrc
	}
	return objs[0], nil
}

var _ cc.CC = ELVM{}
