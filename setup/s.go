package setup

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/hack-pad/hackpadfs"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/coreapi"
	"github.com/pkg/sftp"
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/remount"
	"golang.org/x/crypto/ssh"
)

type SSH interface {
	Fs() (hackpadfs.FS, error)
	Exec(string, []string, sandbox.Stdio) error
}

type NoSSH struct{}

func (n NoSSH) Fs() (hackpadfs.FS, error) {
	return nil, nil
}
func (n NoSSH) Exec(string, []string, sandbox.Stdio) error {
	return fmt.Errorf("invalid ssh")
}

type CliSSH struct {
	*ssh.Client
}

func (n CliSSH) Fs() (hackpadfs.FS, error) {
	sf, err := sftp.NewClient(n.Client)
	return remount.Sftp{sf}, err
}
func (n CliSSH) Exec(d string, r []string, i sandbox.Stdio) error {
	s, err := n.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	s.Stdin = i.Stdin
	s.Stdout = i.Stdout
	s.Stderr = i.Stderr
	c := fmt.Sprintf("cd %s;exec", d)
	for _, s := range r {
		c = fmt.Sprintf("%s $(echo %s | base64 -d)", c, base64.StdEncoding.EncodeToString([]byte(s)))
	}
	return s.Run(c)
}

func SetupRpc(s SSH) (sandbox.Runner, error) {
	f, err := s.Fs()
	if err != nil {
		return nil, err
	}
	r, err := rpc.NewLocalApi()
	if err != nil {
		return nil, err
	}
	return sandbox.State{I: remount.I{r}, Main: "/ipfs", Io: sandbox.OS(), Fs: f, Exec: s.Exec}, nil
}
func NodeRpc(x *core.IpfsNode, path string, s SSH) (func() error, sandbox.Runner, error) {
	f, err := s.Fs()
	if err != nil {
		return nil, nil, err
	}
	p := path
	i, err := coreapi.NewCoreAPI(x)
	if err != nil {
		return nil, nil, err
	}
	err = os.MkdirAll(p, 0777)
	if err != nil {
		return nil, nil, err
	}
	m, err := remount.Mount(remount.I{i}, p)
	if err != nil {
		return nil, nil, err
	}
	return m, sandbox.State{I: remount.I{i}, Main: path, Io: sandbox.OS(), Fs: f, Exec: s.Exec}, nil
}
