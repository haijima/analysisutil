package ssautil

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

type StaticMethodCall struct {
	ssa.CallCommon
	Recv   ssa.Value
	Method *ssa.Function
	Name   string
}

func NewStaticMethodCall(common *ssa.CallCommon) *StaticMethodCall {
	fn := common.Value.(*ssa.Function)
	name := fmt.Sprintf("(%s).%s", fn.Signature.Recv().Type(), fn.Name())
	return &StaticMethodCall{CallCommon: *common, Recv: common.Args[0], Method: fn, Name: name}
}

type DynamicMethodCall struct {
	ssa.CallCommon
	Recv   ssa.Value
	Method *types.Func
	Name   string
}

func NewDynamicMethodCall(common *ssa.CallCommon) *DynamicMethodCall {
	name := fmt.Sprintf("%s.%s", common.Value.Type().String(), common.Method.Name())
	return &DynamicMethodCall{CallCommon: *common, Recv: common.Value, Method: common.Method, Name: name}
}

type BuiltinDynamicMethodCall struct {
	ssa.CallCommon
	Recv   ssa.Value
	Method *types.Func
	Name   string
}

func NewBuiltinDynamicMethodCall(common *ssa.CallCommon) *BuiltinDynamicMethodCall {
	name := fmt.Sprintf("%s.%s", common.Value.Type().String(), common.Method.Name())
	return &BuiltinDynamicMethodCall{CallCommon: *common, Recv: common.Value, Method: common.Method, Name: name}
}

type StaticFunctionCall struct {
	ssa.CallCommon
	Func *ssa.Function
	Name string
}

func NewStaticFunctionCall(common *ssa.CallCommon) *StaticFunctionCall {
	fn := common.Value.(*ssa.Function)
	name := fmt.Sprintf("%s.%s", fn.Pkg.Pkg.Name(), fn.Name())
	if fn.Package() != nil {
		return &StaticFunctionCall{CallCommon: *common, Func: fn, Name: name}
	}
	// generics static function call
	return &StaticFunctionCall{CallCommon: *common, Func: fn.Origin(), Name: name}
}

type BuiltinStaticFunctionCall struct {
	ssa.CallCommon
	Func *ssa.Builtin
	Name string
}

func NewBuiltinStaticFunctionCall(common *ssa.CallCommon) *BuiltinStaticFunctionCall {
	name := common.Value.Name()
	return &BuiltinStaticFunctionCall{CallCommon: *common, Func: common.Value.(*ssa.Builtin), Name: name}
}

type StaticFunctionClosureCall struct {
	ssa.CallCommon
	Parent *ssa.Function
	Func   *ssa.Function
}

func NewStaticFunctionClosureCall(common *ssa.CallCommon) *StaticFunctionClosureCall {
	return &StaticFunctionClosureCall{CallCommon: *common, Parent: common.Value.Parent(), Func: common.Value.(*ssa.MakeClosure).Fn.(*ssa.Function)}
}

type DynamicFunctionCall struct {
	ssa.CallCommon
}

func NewDynamicFunctionCall(common *ssa.CallCommon) *DynamicFunctionCall {
	return &DynamicFunctionCall{CallCommon: *common}
}

type CallInfo interface {
	marker()
}

func (s *StaticMethodCall) marker()          {}
func (d *DynamicMethodCall) marker()         {}
func (b *BuiltinDynamicMethodCall) marker()  {}
func (s *StaticFunctionCall) marker()        {}
func (b *BuiltinStaticFunctionCall) marker() {}
func (s *StaticFunctionClosureCall) marker() {}
func (d *DynamicFunctionCall) marker()       {}

func GetFuncInfo(common *ssa.CallCommon) CallInfo {
	if common.IsInvoke() {
		// dynamic method call
		// e.g.
		// func Something(s fmt.Stringer) {
		//     s.String() // <--- s.String() is dynamic method call
		// }
		// or
		// func another(err error) {
		//     err.Error() // <--- err.Error() is built-in dynamic method call
		// }
		if common.Method.Pkg() == nil {
			return NewBuiltinDynamicMethodCall(common)
		}
		return NewDynamicMethodCall(common)
	} else {
		switch fn := common.Value.(type) {
		case *ssa.Builtin:
			// built-in function call
			// e.g. len, append, etc.
			return NewBuiltinStaticFunctionCall(common)
		case *ssa.MakeClosure:
			// static function closure call
			// e.g. func() { ... }()
			// names are described as xxxFunc$1()
			return NewStaticFunctionClosureCall(common)
		case *ssa.Function:
			if fn.Signature.Recv() == nil {
				// static function call
				return NewStaticFunctionCall(common)
			} else {
				// static method call
				return NewStaticMethodCall(common)
			}
		default:
			// dynamic function call
			return NewDynamicFunctionCall(common)
		}
	}
}

func InstrToCallCommon(instr ssa.Instruction) (*ssa.CallCommon, bool) {
	switch i := instr.(type) {
	case ssa.CallInstruction:
		return i.Common(), true
	case *ssa.Extract:
		if call, ok := i.Tuple.(*ssa.Call); ok {
			return call.Common(), true
		}
	}
	return nil, false
}

func ValueToCallCommon(value ssa.Value) (*ssa.CallCommon, bool) {
	switch v := value.(type) {
	case *ssa.Call:
		return v.Common(), true
	case *ssa.Extract:
		if call, ok := v.Tuple.(*ssa.Call); ok {
			return call.Common(), true
		}
	}
	return nil, false
}
