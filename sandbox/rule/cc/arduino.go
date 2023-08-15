package cc

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	hos "github.com/hack-pad/hackpadfs/os"
	"github.com/portal-co/boogie/sandbox"
	"github.com/portal-co/remount"
)

type Arduino struct {
	Board string
}

func (a Arduino) Compile(st sandbox.Runner, src string, hdrs []string, flags []string, lang string) (string, error) {
	if strings.HasPrefix(lang, "objective") {
		return "", fmt.Errorf("Arduino does not support objective-c")
	}
	var err error
	d := map[string]string{}
	d[src+"."+lang] = src
	for _, h := range hdrs {
		x, ok := d["include"]
		if !ok {
			d["include"] = h
			continue
		}
		d["include"], err = remount.Meld(st.Ipfs(), x, h)
		if err != nil {
			return "", err
		}
	}
	return remount.NewDir(st.Ipfs(), d)
}
func (a Arduino) Link(st sandbox.Runner, sum string, objs []string, combine bool) (string, error) {
	var err error
	r := ""
	for _, h := range objs {
		if r == "" {
			r = h
			continue
		}
		r, err = remount.Meld(st.Ipfs(), r, h)
		if err != nil {
			return "", err
		}
	}
	if combine {
		return r, nil
	}
	d, err := os.MkdirTemp("/tmp", "ard-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(d)
	compRespStream, err := st.(sandbox.ArdRunner).Ard().Compile(context.Background(), &rpc.CompileRequest{
		Instance:   st.(sandbox.ArdRunner).Instance(),
		Fqbn:       a.Board,
		SketchPath: st.IpfsPath() + "/" + r,
		ExportDir:  d,
	})
	if err != nil {
		return "", err
	}
	for {
		compResp, err := compRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Print("Compilation done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Compile error: %s\n", err)
		}

		// When an operation is ongoing you can get its output
		if resp := compResp.GetOutStream(); resp != nil {
			log.Printf("STDOUT: %s", resp)
		}
		if resperr := compResp.GetErrStream(); resperr != nil {
			log.Printf("STDERR: %s", resperr)
		}
	}
	return remount.Push(st.Ipfs(), hos.NewFS(), d[1:])
}
