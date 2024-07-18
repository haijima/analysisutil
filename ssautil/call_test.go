package ssautil_test

import (
	"testing"

	"github.com/haijima/analysisutil"
	"github.com/haijima/analysisutil/ssautil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/ssa"
)

func TestGetFuncInfo(t *testing.T) {
	instrs, err := GetInstructions(t, "./testdata/src/call", "./...")
	require.NoError(t, err)

	staticMethodCalls := make([]*ssautil.StaticMethodCall, 0)
	dynamicMethodCalls := make([]*ssautil.DynamicMethodCall, 0)
	builtinDynamicMethodCalls := make([]*ssautil.BuiltinDynamicMethodCall, 0)
	staticFunctionCalls := make([]*ssautil.StaticFunctionCall, 0)
	builtinStaticFunctionCalls := make([]*ssautil.BuiltinStaticFunctionCall, 0)
	staticFunctionClosureCalls := make([]*ssautil.StaticFunctionClosureCall, 0)
	dynamicFunctionCalls := make([]*ssautil.DynamicFunctionCall, 0)

	for _, instr := range instrs {
		switch instr := instr.(type) {
		case *ssa.Call:
			switch c := ssautil.GetCallInfo(&instr.Call).(type) {
			case *ssautil.StaticMethodCall:
				staticMethodCalls = append(staticMethodCalls, c)
			case *ssautil.DynamicMethodCall:
				dynamicMethodCalls = append(dynamicMethodCalls, c)
			case *ssautil.BuiltinDynamicMethodCall:
				builtinDynamicMethodCalls = append(builtinDynamicMethodCalls, c)
			case *ssautil.StaticFunctionCall:
				staticFunctionCalls = append(staticFunctionCalls, c)
			case *ssautil.BuiltinStaticFunctionCall:
				builtinStaticFunctionCalls = append(builtinStaticFunctionCalls, c)
			case *ssautil.StaticFunctionClosureCall:
				staticFunctionClosureCalls = append(staticFunctionClosureCalls, c)
			case *ssautil.DynamicFunctionCall:
				dynamicFunctionCalls = append(dynamicFunctionCalls, c)
			default:
				t.Fatalf("unexpected call type: %T", c)
			}
		default:
			continue
		}
	}

	assert.Equal(t, 2, len(staticMethodCalls))
	assert.Equal(t, "(github.com/haijima/analysisutil/ssautil/testdata/src/call.Foo).String", staticMethodCalls[0].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticMethodCalls[0].Pkg().Path())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.Foo", staticMethodCalls[0].Recv().Type().String())
	assert.Equal(t, "String", staticMethodCalls[0].Method().Name())
	assert.Panics(t, func() { staticMethodCalls[0].Arg(0) })
	assert.Equal(t, "(*github.com/haijima/analysisutil/ssautil/testdata/src/call.Fizz).String", staticMethodCalls[1].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticMethodCalls[1].Pkg().Path())
	assert.Equal(t, "*github.com/haijima/analysisutil/ssautil/testdata/src/call.Fizz", staticMethodCalls[1].Recv().Type().String())
	assert.Equal(t, "String", staticMethodCalls[1].Method().Name())
	assert.Panics(t, func() { staticMethodCalls[1].Arg(0) })

	assert.True(t, staticMethodCalls[0].Match("(github.com/haijima/analysisutil/ssautil/testdata/src/call.Foo).String"))
	assert.True(t, staticMethodCalls[0].Match("(*github.com/haijima/analysisutil/ssautil/testdata/src/call.Foo).String"))
	assert.True(t, staticMethodCalls[0].Match("(*.Foo).String"))
	assert.True(t, staticMethodCalls[0].Match("(*github.com/haijima/analysisutil/ssautil/testdata/src/call.*).String"))
	assert.True(t, staticMethodCalls[0].Match("(*github.com/haijima/analysisutil/ssautil/testdata/src/call.Foo).*"))
	assert.True(t, staticMethodCalls[0].Match("(*.*).String"))
	assert.True(t, staticMethodCalls[0].Match("(*github.com/haijima/analysisutil/ssautil/testdata/src/call.*).*"))
	assert.True(t, staticMethodCalls[0].Match("(*.Foo).*"))
	assert.True(t, staticMethodCalls[0].Match("(*.*).*"))
	assert.True(t, staticMethodCalls[1].Match("(github.com/haijima/analysisutil/ssautil/testdata/src/call.Fizz).String"))
	assert.True(t, staticMethodCalls[1].Match("(*github.com/haijima/analysisutil/ssautil/testdata/src/call.Fizz).String"))

	assert.Equal(t, 3, len(dynamicMethodCalls))
	assert.Equal(t, "fmt.Stringer.String", dynamicMethodCalls[0].Name())
	assert.Equal(t, "fmt", dynamicMethodCalls[0].Pkg().Path())
	assert.Equal(t, "fmt.Stringer", dynamicMethodCalls[0].Recv().Type().String())
	assert.Equal(t, "String", dynamicMethodCalls[0].Method().Name())
	assert.Panics(t, func() { dynamicMethodCalls[0].Arg(0) })
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.Barer.Bar", dynamicMethodCalls[1].Name()) // foo
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", dynamicMethodCalls[1].Pkg().Path())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.Barer", dynamicMethodCalls[1].Recv().Type().String())
	assert.Equal(t, "Bar", dynamicMethodCalls[1].Method().Name())
	assert.Panics(t, func() { dynamicMethodCalls[1].Arg(0) })
	assert.Equal(t, "T.String", dynamicMethodCalls[2].Name()) // foo
	assert.Equal(t, "fmt", dynamicMethodCalls[2].Pkg().Path())
	assert.Equal(t, "T", dynamicMethodCalls[2].Recv().Type().String())
	assert.Equal(t, "String", dynamicMethodCalls[2].Method().Name())
	assert.Panics(t, func() { dynamicMethodCalls[2].Arg(0) })

	assert.True(t, dynamicMethodCalls[0].Match("fmt.Stringer.String"))
	assert.True(t, dynamicMethodCalls[0].Match("*.Stringer.String"))
	assert.True(t, dynamicMethodCalls[0].Match("fmt.*.String"))
	assert.True(t, dynamicMethodCalls[0].Match("fmt.Stringer.*"))
	assert.True(t, dynamicMethodCalls[0].Match("*.*.String"))
	assert.True(t, dynamicMethodCalls[0].Match("fmt.*.*"))
	assert.True(t, dynamicMethodCalls[0].Match("*.Stringer.*"))
	assert.True(t, dynamicMethodCalls[0].Match("*.*.*"))
	assert.True(t, dynamicMethodCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.Barer.Bar"))
	assert.True(t, dynamicMethodCalls[1].Match("*.Barer.Bar"))
	assert.True(t, dynamicMethodCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.*.Bar"))
	assert.True(t, dynamicMethodCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.Barer.*"))
	assert.True(t, dynamicMethodCalls[1].Match("*.*.Bar"))
	assert.True(t, dynamicMethodCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.*.*"))
	assert.True(t, dynamicMethodCalls[1].Match("*.Barer.*"))
	assert.True(t, dynamicMethodCalls[1].Match("*.*.*"))

	assert.Equal(t, 1, len(builtinDynamicMethodCalls))
	assert.Equal(t, "error.Error", builtinDynamicMethodCalls[0].Name())
	assert.Equal(t, "error", builtinDynamicMethodCalls[0].Recv().Type().String())
	assert.Equal(t, "Error", builtinDynamicMethodCalls[0].Method().Name())
	assert.Panics(t, func() { builtinDynamicMethodCalls[0].Arg(0) })

	assert.True(t, builtinDynamicMethodCalls[0].Match("error.Error"))
	assert.True(t, builtinDynamicMethodCalls[0].Match("*.Error"))
	assert.True(t, builtinDynamicMethodCalls[0].Match("error.*"))
	assert.True(t, builtinDynamicMethodCalls[0].Match("*.*"))

	assert.Equal(t, 4, len(staticFunctionCalls))
	assert.Equal(t, "fmt.Println", staticFunctionCalls[0].Name())
	assert.Equal(t, "fmt", staticFunctionCalls[0].Pkg().Path())
	assert.Equal(t, "Println", staticFunctionCalls[0].Func().Name())
	assert.NotNil(t, staticFunctionCalls[0].Arg(0))
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.foo", staticFunctionCalls[1].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticFunctionCalls[1].Pkg().Path())
	assert.Equal(t, "foo", staticFunctionCalls[1].Func().Name())
	assert.NotNil(t, staticFunctionCalls[1].Arg(0))
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.anonymousStaticFunc$1", staticFunctionCalls[2].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticFunctionCalls[2].Pkg().Path())
	assert.Equal(t, "anonymousStaticFunc$1", staticFunctionCalls[2].Func().Name())
	assert.Panics(t, func() { staticFunctionCalls[2].Arg(0) })
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.getCallable", staticFunctionCalls[3].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticFunctionCalls[3].Pkg().Path())
	assert.Equal(t, "getCallable", staticFunctionCalls[3].Func().Name())
	assert.Panics(t, func() { staticFunctionCalls[3].Arg(0) })

	assert.True(t, staticFunctionCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.foo"))
	assert.True(t, staticFunctionCalls[1].Match("*.foo"))
	assert.True(t, staticFunctionCalls[1].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.*"))
	assert.True(t, staticFunctionCalls[1].Match("*.*"))

	assert.Equal(t, 1, len(builtinStaticFunctionCalls))
	assert.Equal(t, "append", builtinStaticFunctionCalls[0].Name())
	assert.Equal(t, "append", builtinStaticFunctionCalls[0].Func().Name())

	assert.True(t, builtinStaticFunctionCalls[0].Match("append"))
	assert.True(t, builtinStaticFunctionCalls[0].Match("*"))

	assert.Equal(t, 1, len(staticFunctionClosureCalls))
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call.staticFuncClosure$1", staticFunctionClosureCalls[0].Name())
	assert.Equal(t, "github.com/haijima/analysisutil/ssautil/testdata/src/call", staticFunctionClosureCalls[0].Func().Pkg.Pkg.Path())
	assert.Equal(t, "staticFuncClosure$1", staticFunctionClosureCalls[0].Func().Name())

	assert.True(t, staticFunctionClosureCalls[0].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.staticFuncClosure$1"))
	assert.True(t, staticFunctionClosureCalls[0].Match("*.staticFuncClosure$1"))
	assert.True(t, staticFunctionClosureCalls[0].Match("github.com/haijima/analysisutil/ssautil/testdata/src/call.*"))
	assert.True(t, staticFunctionClosureCalls[0].Match("*.*"))

	assert.Equal(t, 3, len(dynamicFunctionCalls))
	assert.Equal(t, "fn", dynamicFunctionCalls[0].Name())
	assert.Panics(t, func() { dynamicFunctionCalls[0].Arg(0) })
	assert.Equal(t, "callableVar", dynamicFunctionCalls[1].Name())
	assert.Panics(t, func() { dynamicFunctionCalls[1].Arg(0) })
	assert.Equal(t, "getCallable", dynamicFunctionCalls[2].Name())
	assert.NotNil(t, dynamicFunctionCalls[2].Arg(0))

	assert.False(t, dynamicFunctionCalls[0].Match("fn")) // Always false
}

func GetInstructions(t *testing.T, dir string, patterns ...string) ([]ssa.Instruction, error) {
	t.Helper()

	pkgs, err := analysisutil.LoadPackages(dir, patterns...)
	if err != nil {
		return nil, err
	}

	result := make([]ssa.Instruction, 0)
	for _, pkg := range pkgs {
		ssaProg, err := ssautil.BuildSSA(pkg)
		if err != nil {
			return nil, err
		}
		for _, fn := range ssaProg.SrcFuncs {
			if fn.Name() == "main" {
				continue
			}
			for _, b := range fn.Blocks {
				for _, instr := range b.Instrs {
					result = append(result, instr)
				}
			}
		}
	}
	return result, nil
}
