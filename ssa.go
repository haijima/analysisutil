package analysisutil

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

func ValueToStrings(v ssa.Value) ([]string, bool) {
	return valueToStrings(v, 0)
}

func valueToStrings(v ssa.Value, depth int) ([]string, bool) {
	if depth > 10 {
		return []string{}, false
	}
	depth++
	switch t := v.(type) {
	case *ssa.Const:
		return constToStrings(t)
	case *ssa.BinOp:
		return binOpToStrings(t, depth)
	case *ssa.Phi:
		return phiToStrings(t, depth)
	default:
		return []string{}, false
	}
}

func constToStrings(t *ssa.Const) ([]string, bool) {
	if t.Value != nil && t.Value.Kind() == constant.String {
		if s, err := Unquote(t.Value.ExactString()); err == nil {
			return []string{s}, true
		}
	}
	return []string{}, false
}

func binOpToStrings(t *ssa.BinOp, depth int) ([]string, bool) {
	x, xok := valueToStrings(t.X, depth)
	y, yok := valueToStrings(t.Y, depth)
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

func phiToStrings(t *ssa.Phi, depth int) ([]string, bool) {
	res := make([]string, 0, len(t.Edges))
	for _, edge := range t.Edges {
		if s, ok := valueToStrings(edge, depth); ok {
			res = slices.Concat(res, s)
		}
	}
	return res, len(res) > 0
}

var fmtVerbRegexp = regexp.MustCompile(`(^|[^%]|(?:%%)+)(%(?:-?\d+|\+|#)?)(\w)`)

// fmtSprintfToStrings returns the possible string values of fmt.Sprintf.
func fmtSprintfToStrings(t *ssa.Call, depth int) ([]string, bool) {
	fs, ok := valueToStrings(t.Call.Args[0], depth)
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
func stringsJoinToStrings(t *ssa.Call, depth int) ([]string, bool) {
	// strings.Join
	joiner, ok := valueToStrings(t.Call.Args[1], depth)
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
