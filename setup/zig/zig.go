package zig

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hack-pad/hackpadfs/tar"
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/boogie/setup/cc"
	"github.com/portal-co/remount"
	"github.com/ulikunitz/xz"
)

var ZPath = "zig-linux-x86_64-" + ZVer

var ZVer = "0.11.0"

var ZCache = "QmXvLZoRPZd76YUVzbijrxZmw5MGHt8n1sNkV329s9FAH1"

func GetZigTarball() (*http.Response, error) {
	return http.Get("https://ziglang.org/download/" + ZVer + "/" + ZPath + ".tar.xz")
}
func GetZig(x remount.Pusher) (string, error) {
	z, err := GetZigTarball()
	if err != nil {
		return "", err
	}
	defer z.Body.Close()
	xzs, err := xz.NewReader(z.Body)
	if err != nil {
		return "", err
	}
	t, err := tar.NewReaderFS(context.Background(), xzs, tar.ReaderFSOptions{})
	if err != nil {
		return "", err
	}
	return x.Push(t, ZPath)
}

type ZigPkg struct {
	Deps []string
	Src  string
	Path string
}
type ZigCompiler interface {
	cc.CXX
	ZigCompile(r sandbox.Runner, mods map[string]ZigPkg, root ZigPkg) (string, error)
}
type CanonZig struct {
	Target, Zig string
}

// CXX implements ZigCompiler.
func (c CanonZig) CXX(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	return c.ccBase(r, true, src, hdrs, flags)
}

// CC implements ZigCompiler.
func (c CanonZig) CC(r sandbox.Runner, src string, hdrs []string, flags []string) (string, error) {
	return c.ccBase(r, false, src, hdrs, flags)
}

// Link implements ZigCompiler.
func (c CanonZig) Link(r sandbox.Runner, libs []string, objs []string) (string, error) {
	d := map[string]string{"zig": c.Zig}
	m := []string{"./zig/zig"}
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

func (c CanonZig) ccBase(r sandbox.Runner, cpp bool, src string, hdrs []string, flags []string) (string, error) {
	d := map[string]string{"zig": c.Zig}
	if !cpp {
		d["target.c"] = src
	} else {
		d["target.cc"] = src
	}
	m := []string{"./zig/zig"}
	if cpp {
		m = append(m, "c++", "target.cc")
	} else {
		m = append(m, "cc", "target.c")
	}
	if c.Target != "" {
		m = append(m, "-target", c.Target)
	}
	m = append(m, flags...)
	m = append(m, "-o", "target.o", "-c")
	for _, h := range hdrs {
		d[h] = h
		m = append(m, "-isystem", "./"+h)
	}
	s, err := r.Run(d, m, []string{}, "/target.o")
	return s, err
}

// ZigCompile implements ZigCompiler.
func (c CanonZig) ZigCompile(r sandbox.Runner, mods map[string]ZigPkg, root ZigPkg) (string, error) {
	d := map[string]string{"zig": c.Zig, "root": root.Src}
	m := []string{"./zig/zig", "build-obj", "-o", "target.o", "--main-pkg-path", "./root/" + root.Path, "--deps", strings.Join(root.Deps, ",")}
	if c.Target != "" {
		m = append(m, "-target", c.Target)
	}
	for mn, n := range mods {
		d["$"+mn] = n.Src
		m = append(m, "--mod", fmt.Sprintf("%s:%s:./$%s/%s", mn, strings.Join(n.Deps, ","), mn, n.Path))
	}
	s, err := r.Run(d, m, []string{}, "/target.o")
	return s, err
}

var _ ZigCompiler = CanonZig{}
