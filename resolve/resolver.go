package resolve

import (
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/remount"
)

type Git interface {
	Resolve(x plumbing.Hash) (string, error)
}
type GoGit struct {
	*git.Worktree
	sandbox.PFS
	mtx   sync.Mutex
	Cache map[plumbing.Hash]string
}

// Resolve implements Git.
func (g *GoGit) Resolve(x plumbing.Hash) (string, error) {
	if c, ok := g.Cache[x]; ok {
		return c, nil
	}
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.Worktree.Checkout(&git.CheckoutOptions{Hash: x})
	a, err := g.Push(remount.FB{g.Worktree.Filesystem}, "")
	if err != nil {
		return "", err
	}
	g.Cache[x] = a
	return a, nil
}

var _ Git = &GoGit{}
