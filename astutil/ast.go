package astutil

import (
	"go/ast"
)

func Include(u, v ast.Node) bool {
	return u.Pos() <= v.Pos() && v.End() <= u.End()
}
