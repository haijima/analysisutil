package ssautil

import (
	"go/ast"
	"go/types"

	"github.com/haijima/analysisutil"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

func LoadBuildSSAs(dir string, patterns ...string) ([]*buildssa.SSA, error) {
	pkgs, err := analysisutil.LoadPackages(dir, patterns...)
	if err != nil {
		return nil, err
	}

	ssaProgs := make([]*buildssa.SSA, 0, len(pkgs))
	for _, pkg := range pkgs {
		ssaProg, err := BuildSSA(pkg)
		if err != nil {
			return nil, err
		}
		ssaProgs = append(ssaProgs, ssaProg)
	}

	return ssaProgs, nil
}

func LoadInstrs(dir string, patterns ...string) ([]ssa.Instruction, error) {
	pkgs, err := analysisutil.LoadPackages(dir, patterns...)
	if err != nil {
		return nil, err
	}

	instrs := make([]ssa.Instruction, 0)
	for _, pkg := range pkgs {
		ssaProg, err := BuildSSA(pkg)
		if err != nil {
			return nil, err
		}
		for _, fn := range ssaProg.SrcFuncs {
			for _, b := range fn.Blocks {
				for _, instr := range b.Instrs {
					instrs = append(instrs, instr)
				}
			}
		}
	}
	return instrs, nil
}

// BuildSSA See: buildssa.Analyzer.
func BuildSSA(pkg *packages.Package) (*buildssa.SSA, error) {
	prog := ssa.NewProgram(pkg.Fset, ssa.BuilderMode(0))

	// Create SSA packages for direct imports.
	for _, p := range pkg.Types.Imports() {
		prog.CreatePackage(p, nil, nil, true)
	}

	// Create and build the primary package.
	ssapkg := prog.CreatePackage(pkg.Types, pkg.Syntax, pkg.TypesInfo, false)
	ssapkg.Build()

	// Compute list of source functions, including literals,
	// in source order.
	var funcs []*ssa.Function
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			if fdecl, ok := decl.(*ast.FuncDecl); ok {
				if fn, ok := pkg.TypesInfo.Defs[fdecl.Name].(*types.Func); ok {
					f := ssapkg.Prog.FuncValue(fn)
					if f == nil {
						panic(fn)
					}

					var addAnons func(f *ssa.Function)
					addAnons = func(f *ssa.Function) {
						funcs = append(funcs, f)
						for _, anon := range f.AnonFuncs {
							addAnons(anon)
						}
					}
					addAnons(f)
				}
			}
		}
	}

	return &buildssa.SSA{Pkg: ssapkg, SrcFuncs: funcs}, nil
}
