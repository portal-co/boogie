package sandbox

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/hack-pad/hackpadfs"
	hos "github.com/hack-pad/hackpadfs/os"

	"github.com/portal-co/remount"
)

type Stdio struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

func OS() Stdio {
	return Stdio{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

type PFS interface {
	remount.Pusher
	hackpadfs.FS
}

type State struct {
	I    PFS
	Fs   hackpadfs.FS
	Exec func(string, []string, Stdio) error
	Main string
	Io   Stdio
}

func (s State) IO() Stdio {
	return s.Io
}

func (s State) WithIO(x Stdio) Runner {
	s.Io = x
	return s
}

func (state State) IpfsPath() string {
	return state.Main
}

func (state State) Ipfs() PFS {
	return state.I
}

func (state State) Run(inputs map[string]string, cmd []string, outs []string) (string, error) {
	ix := inputs
	t, err := os.MkdirTemp("/tmp", "prtl-sbox-base-*")
	if err != nil {
		return "", err
	}
	// g = errgroup.Group{}
	l := []string{}
	if state.Fs == nil {
		for k, v := range ix {
			// err := os.Symlink(path.Join(state.Main, v), path.Join(t, k))
			// if err != nil {
			// 	return "", err
			// }
			// l = append(l, path.Join(t, k))
			// fmt.Println(v)
			err := remount.Clone(state.I, hos.NewFS(), v, path.Join(t, k)[1:])
			if err != nil {
				// fmt.Println(err)
				return "", err
				// continue
			}
			// fmt.Println("cloned", k)
			l = append(l, path.Join(t, k))
		}
		// u, err := os.MkdirTemp("/tmp", "prtl-sbox-8")
		// if err != nil {
		// 	return "", err
		// }
		// d := exec.Command("/usr/bin/env", "bindfs", "-f", "--resolve-symlinks", "-p", "a+Xx", t, u)
		// d.Stderr = os.Stderr
		// err = d.Start()
		// if err != nil {
		// 	err = fmt.Errorf("binding at %s at %w", u, err)
		// 	return "", err
		// }
		// defer func() {
		// 	err := d.Process.Signal(syscall.SIGINT)
		// 	if err != nil {
		// 		return
		// 	}
		// 	err = d.Wait()
		// 	if err != nil {
		// 		err = fmt.Errorf("binding at %s at %w", u, err)
		// 		return
		// 	}
		// 	// d := exec.Command("fusermount", "-u", u)
		// 	// d.Stderr = os.Stderr
		// 	// if err != nil {
		// 	// 	return
		// 	// }
		// 	// err = d.Run()
		// }()
		// b := false
		// for !b {
		// 	b, err = mountinfo.Mounted(u)
		// 	if err != nil {
		// 		return "", err
		// 	}
		// }
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout = state.Io.Stdout
		c.Stderr = state.Io.Stderr
		c.Stdin = state.Io.Stdin
		c.Dir = t
		err = c.Run()
		if err != nil {
			return "", err
		}
		return state.I.Push(hos.NewFS(), t[1:])
	} else {
		err = hackpadfs.MkdirAll(state.Fs, t[1:], 0777)
		if err != nil {
			return "", err
		}
		defer hackpadfs.RemoveAll(state.Fs, t[1:])
		for k, v := range ix {
			// err := os.Symlink(path.Join(state.Main, v), path.Join(t, k))
			err := remount.Clone(state.I, state.Fs, v, path.Join(t, k)[1:])
			if err != nil {
				return "", err
			}
			l = append(l, path.Join(t, k))
		}
		err = state.Exec(t, cmd, state.Io)
		if err != nil {
			return "", err
		}
		return state.I.Push(state.Fs, t[1:])
	}
}

type Runner interface {
	Run(inputs map[string]string, cmd []string, outs []string) (string, error)
	Ipfs() PFS
	IpfsPath() string
	IO() Stdio
	WithIO(x Stdio) Runner
}

type ArdRunner interface {
	Runner
	Ard() rpc.ArduinoCoreServiceClient
	Instance() *rpc.Instance
}

type Ramp[T comparable] map[T]Runner

func (r Ramp[T]) Ard() rpc.ArduinoCoreServiceClient {
	var defaul T
	return r[defaul].(ArdRunner).Ard()
}
func (r Ramp[T]) Instance() *rpc.Instance {
	var defaul T
	return r[defaul].(ArdRunner).Instance()
}

type ste struct {
	s string
	e error
}

func C[T comparable](x map[T]error) error {
	err := fmt.Errorf("")
	for k, v := range x {
		err = fmt.Errorf("%w\n%v: %w", err, k, v)
	}
	return err
}

func (r Ramp[T]) Run(inputs map[string]string, cmd []string, outs []string) (string, error) {
	c := make(chan string)
	errs := map[T]error{}
	left := 0
	errsc := make(chan struct{})
	for k, v := range r {
		v := v
		k := k
		left++
		go func() {
			a, err := v.Run(inputs, cmd, outs)
			if err != nil {
				errs[k] = err
				left--
				if left == 0 {
					errsc <- struct{}{}
				}
				return
			}
			c <- a
		}()
	}
	select {
	case ch := <-c:
		return ch, nil
	case <-errsc:
		return "", C(errs)
	}
}
func (r Ramp[T]) Ipfs() PFS {
	var defaul T
	return r[defaul].Ipfs()
}
func (r Ramp[T]) IpfsPath() string {
	var defaul T
	return r[defaul].IpfsPath()
}
func (r Ramp[T]) IO() Stdio {
	var defaul T
	return r[defaul].IO()
}
func (r Ramp[T]) WithIO(x Stdio) Runner {
	u := Ramp[T]{}
	for k, v := range r {
		u[k] = v.WithIO(x)
	}
	return u
}
