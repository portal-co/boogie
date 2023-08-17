package startogo

import (
	"go.starlark.net/syntax"
)

func StarToGo(n syntax.Node) string {
	switch n := n.(type) {
	case *syntax.Ident:
		return n.Name
	case *syntax.CallExpr:

	}
}
