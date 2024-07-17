package ssautil

import (
	"go/token"
	"slices"

	"golang.org/x/tools/go/ssa"
)

func GetPosition(pkg *ssa.Package, pos ...token.Pos) token.Position {
	if pkg == nil || pkg.Prog == nil || pkg.Prog.Fset == nil {
		return token.Position{}
	}
	if i := slices.IndexFunc(pos, func(p token.Pos) bool { return p.IsValid() }); i > -1 {
		return pkg.Prog.Fset.Position(pos[i])
	}
	return token.Position{}
}
