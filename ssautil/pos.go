package ssautil

import (
	"go/token"
	"go/types"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/ssa"
)

var pathDirRegex = regexp.MustCompile(`([^/]+)/`)

type Posx struct {
	Func *ssa.Function
	Pos  []token.Pos
}

func NewPos(fn *ssa.Function, pos ...token.Pos) *Posx {
	return &Posx{Func: fn, Pos: pos}
}

func (p *Posx) Add(pos token.Pos) *Posx {
	return &Posx{Func: p.Func, Pos: append(p.Pos, pos)}
}

func (m *Posx) Package() *types.Package {
	if m.Func == nil || m.Func.Pkg == nil {
		return &types.Package{}
	}
	return m.Func.Pkg.Pkg
}

func (p *Posx) PackagePath(abbreviate bool) string {
	if abbreviate {
		return pathDirRegex.ReplaceAllStringFunc(p.Package().Path(), func(m string) string { return m[:1] + "/" })
	}
	return p.Package().Path()

}

func (m *Posx) Position() token.Position {
	if m.Func == nil {
		return token.Position{}
	}
	return GetPosition(m.Func.Pkg, m.Pos...)
}

func (p *Posx) PositionString() string {
	return filepath.Base(p.Position().String())
}

func (p *Posx) Compare(other *Posx) int {
	if p.Package().Path() != other.Package().Path() {
		return strings.Compare(p.Package().Path(), other.Package().Path())
	} else if p.Position().Filename != other.Position().Filename {
		return strings.Compare(p.Position().Filename, other.Position().Filename)
	} else if p.Position().Offset != other.Position().Offset {
		return p.Position().Offset - other.Position().Offset
	}
	return 0
}

func (p *Posx) Equal(other *Posx) bool {
	return p.Compare(other) == 0
}

func (p *Posx) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("package", p.PackagePath(true)),
		slog.String("file", p.PositionString()),
		slog.String("func", p.Func.Name()),
	)
}
