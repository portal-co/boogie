package build

import (
	"context"

	"github.com/portal-co/boogie/sandbox"
)

type BuildFunc func(ctx context.Context, s sandbox.Runner) (string, error)
