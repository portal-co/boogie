package cc

import "github.com/portal-co/boogie/sandbox"

type Linker interface {
	Link(r sandbox.Runner, libs []string, objs []string) (string, error)
}
type CC interface {
	Linker
	CC(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error)
}
type CXX interface {
	CC
	CXX(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error)
}
