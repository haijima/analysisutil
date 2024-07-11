package ssautil

import (
	"go/ast"
	"go/constant"
	"go/token"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type ToConstsFunc[T any] func(v ssa.Value, next func(v ssa.Value) ([]T, bool)) ([]T, bool)

func ValueToConsts[T any](v ssa.Value, flattener ToConstsFunc[T], mapper func(t *ssa.Const) (T, bool)) ([]T, bool) {
	return ValueToConstsWithMaxDepth[T](v, 10, flattener, mapper)
}

func ValueToConstsWithMaxDepth[T any](v ssa.Value, maxDepth int, flattener ToConstsFunc[T], mapper func(t *ssa.Const) (T, bool)) ([]T, bool) {
	return valueToConsts[T](v, 0, maxDepth, flattener, mapper)
}

func valueToConsts[T any](v ssa.Value, depth, maxDepth int, flattener ToConstsFunc[T], mapper func(t *ssa.Const) (T, bool)) ([]T, bool) {
	if depth > maxDepth {
		return []T{}, false
	}
	depth++
	switch t := v.(type) {
	case *ssa.Const:
		if t, ok := mapper(t); ok {
			return []T{t}, true
		}
	case *ssa.Phi:
		return phiToConsts[T](t, depth, maxDepth, flattener, mapper)
	default:
		if cs, ok := flattener(t, func(v ssa.Value) ([]T, bool) {
			return valueToConsts[T](v, depth, maxDepth, flattener, mapper)
		}); ok {
			return cs, true
		}
	}
	return []T{}, false
}

func phiToConsts[T any](t *ssa.Phi, depth, maxDepth int, flattener ToConstsFunc[T], mapper func(t *ssa.Const) (T, bool)) ([]T, bool) {
	res := make([]T, 0, len(t.Edges))
	for _, edge := range t.Edges {
		if c, ok := valueToConsts(edge, depth, maxDepth, flattener, mapper); ok {
			res = slices.Concat(res, c)
		}
	}
	return res, len(res) > 0
}

func ValueToStrings(v ssa.Value) ([]string, bool) {
	return ValueToStringsWithMaxDepth(v, 10)
}

func ValueToStringsWithMaxDepth(v ssa.Value, maxDepth int) ([]string, bool) {
	return valueToConsts[string](v, 0, maxDepth,
		// flattener
		func(v ssa.Value, next func(v ssa.Value) ([]string, bool)) ([]string, bool) {
			switch t := v.(type) {
			case *ssa.BinOp:
				return binOpToStrings(t, next)
			case *ssa.Call:
				switch c := GetFuncInfo(t.Common()).(type) {
				case *StaticFunctionCall:
					if c.Name() == "fmt.Sprintf" {
						return fmtSprintfToStrings(t, next)
					} else if c.Name() == "strings.Join" {
						return stringsJoinToStrings(t, next)
					}
				}
			}
			return []string{}, false
		},
		// mapper
		func(t *ssa.Const) (string, bool) {
			if t.Value != nil && t.Value.Kind() == constant.String {
				if s, err := Unquote(t.Value.ExactString()); err == nil {
					return s, true
				}
			}
			return "", false
		},
	)
}

func binOpToStrings(t *ssa.BinOp, fn func(v ssa.Value) ([]string, bool)) ([]string, bool) {
	x, xok := fn(t.X)
	y, yok := fn(t.Y)
	if xok && yok && len(x) > 0 && len(y) > 0 && t.Op == token.ADD {
		res := make([]string, 0, len(x)*len(y))
		for _, xx := range x {
			for _, yy := range y {
				res = append(res, xx+yy)
			}
		}
		return res, true
	}
	return []string{}, false
}

var fmtVerbRegexp = regexp.MustCompile(`(^|[^%]|(?:%%)+)(%(?:-?\d+|\+|#)?)(\w)`)

// fmtSprintfToStrings returns the possible string values of fmt.Sprintf.
func fmtSprintfToStrings(t *ssa.Call, fn func(v ssa.Value) ([]string, bool)) ([]string, bool) {
	fs, ok := fn(t.Call.Args[0])
	if !ok && len(fs) == 1 {
		return []string{}, false
	}
	f := fmtVerbRegexp.ReplaceAllStringFunc(fs[0], func(s string) string {
		m := fmtVerbRegexp.FindAllStringSubmatch(s, 1)
		if m == nil || len(m) < 1 || len(m[0]) < 4 {
			return s
		}
		switch m[0][3] {
		case "b":
			return m[0][1] + "01"
		case "c":
			return m[0][1] + "a"
		case "t":
			return m[0][1] + "true"
		case "T":
			return m[0][1] + "string"
		case "e":
			return m[0][1] + "1.234000e+08"
		case "E":
			return m[0][1] + "1.234000E+08"
		case "p":
			return m[0][1] + "0xc0000ba000"
		case "x":
			return m[0][1] + "1f"
		case "d":
			return m[0][1] + "1"
		case "f":
			return m[0][1] + "1.0"
		default:
			return s
		}
	})
	if !fmtVerbRegexp.MatchString(f) { // no more verbs
		return []string{f}, true
	}
	return []string{}, false
}

// stringsJoinToStrings returns the possible string values of strings.Join.
func stringsJoinToStrings(t *ssa.Call, fn func(v ssa.Value) ([]string, bool)) ([]string, bool) {
	// strings.Join
	joiner, ok := fn(t.Call.Args[1])
	if !ok || len(joiner) != 1 {
		return []string{}, false
	}
	firstArg := t.Call.Args[0]
	astArgs := make([]string, 0)
	ast.Inspect(t.Parent().Syntax(), func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if cl, ok := n.(*ast.CompositeLit); ok && n.Pos() <= firstArg.Pos() && firstArg.Pos() < n.End() {
			for _, elt := range cl.Elts {
				if bl, ok := elt.(*ast.BasicLit); ok {
					if unquoted, err := Unquote(bl.Value); err == nil {
						astArgs = append(astArgs, unquoted)
					}
				}
			}
			if len(astArgs) != len(cl.Elts) {
				// not all elements are constant or some elements are failed to unquote
				astArgs = []string{}
			}
			return false
		}
		return true
	})
	if len(astArgs) > 0 {
		return []string{strings.Join(astArgs, joiner[0])}, true
	}
	return []string{}, false
}

func Unquote(str string) (string, error) {
	for _, c := range []uint8{'`', '"', '\''} {
		if len(str) >= 2 && str[0] == c && str[len(str)-1] == c {
			return strconv.Unquote(str)
		}
	}
	return str, nil
}
