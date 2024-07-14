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
		} else {
			//	return []T{}, false
		}
	}
	return res, len(res) > 0
	//return valueToConsts[T](t.Edges[0], depth, maxDepth, flattener, mapper)
}

func ValueToInts(v ssa.Value) ([]int, bool) {
	return ValueToIntsWithMaxDepth(v, 10)
}

func ValueToIntsWithMaxDepth(v ssa.Value, maxDepth int) ([]int, bool) {
	return valueToConsts[int](v, 0, maxDepth,
		func(v ssa.Value, next func(v ssa.Value) ([]int, bool)) ([]int, bool) {
			switch t := v.(type) {
			case *ssa.BinOp:
				x, xok := next(t.X)
				y, yok := next(t.Y)
				if xok && yok && len(x) > 0 && len(y) > 0 {
					res := make([]int, 0, len(x)*len(y))
					for _, xx := range x {
						for _, yy := range y {
							switch t.Op {
							case token.ADD:
								res = append(res, xx+yy)
							case token.SUB:
								res = append(res, xx-yy)
							case token.MUL:
								res = append(res, xx*yy)
							case token.QUO:
								if yy != 0 {
									res = append(res, xx/yy)
								}
							}
						}
					}
					return res, true
				}
			}
			return []int{}, false
		}, func(t *ssa.Const) (int, bool) {
			if t.Value != nil && t.Value.Kind() == constant.Int {
				if s, err := Unquote(t.Value.ExactString()); err == nil {
					if i, err := strconv.Atoi(s); err == nil {
						return i, true
					}
				}
			}
			return 0, false
		})
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
				c := GetCallInfo(t.Common())
				if c.Match("fmt.Sprintf") {
					return fmtSprintfToStrings(t, next)
				} else if c.Match("strings.Join") {
					return stringsJoinToStrings(t, next)
				}
			case *ssa.Slice:
				// e.g.
				// s := "hello"
				// s[:len(s)-1]
				s, ok := next(t.X)
				if ok {
					res := make([]string, 0, len(s))
					for _, ss := range s {
						l, lok := []int{0}, true
						h, hok := []int{len(ss)}, true
						if t.Low != nil {
							l, lok = stringIndex(t.Low, t.X, len(ss), maxDepth)
						}
						if t.High != nil {
							h, hok = stringIndex(t.High, t.X, len(ss), maxDepth)
						}
						if lok && hok {
							for _, ll := range l {
								for _, hh := range h {
									res = append(res, ss[ll:hh])
								}
							}
						}
					}
					return res, len(res) > 0
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

func stringIndex(v ssa.Value, ref ssa.Value, strLen int, maxDepth int) ([]int, bool) {
	if i, ok := ValueToIntsWithMaxDepth(v, maxDepth); ok {
		return i, true
	}

	if binOp, ok := v.(*ssa.BinOp); ok && binOp.Op == token.SUB {
		if call, ok := binOp.X.(*ssa.Call); ok {
			c := GetCallInfo(call.Common())
			if c.Name() == "len" && c.Arg(0) == ref {
				if y, ok := ValueToIntsWithMaxDepth(binOp.Y, maxDepth); ok {
					res := make([]int, 0, len(y))
					for _, yy := range y {
						res = append(res, strLen-yy)
					}
					return res, true
				}
			}
		}
	}

	return []int{}, false
}
