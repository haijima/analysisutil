package ssautil

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

type Namer interface {
	Name() string
}
type Packager interface {
	Pkg() *types.Package
}
type Method interface {
	Recv() ssa.Value
	Method() *types.Func
	Arg(idx int) ssa.Value
	ArgsLen() int
}
type Function interface {
	Func() *ssa.Function
	Arg(idx int) ssa.Value
	ArgsLen() int
}
type PackageMethod interface {
	Packager
	Method
}
type PackageFunction interface {
	Packager
	Function
}

type StaticMethodCall struct {
	ssa.CallCommon
}

func NewStaticMethodCall(common *ssa.CallCommon) *StaticMethodCall {
	return &StaticMethodCall{CallCommon: *common}
}

func (s *StaticMethodCall) String() string {
	return s.Signature().String()
}
func (s *StaticMethodCall) Name() string {
	return fmt.Sprintf("(%s).%s", s.Signature().Recv().Type(), s.Method().Name())
}
func (s *StaticMethodCall) Pkg() *types.Package {
	return s.Signature().Recv().Pkg()
}
func (s *StaticMethodCall) Recv() ssa.Value {
	return s.Args[0]
}
func (s *StaticMethodCall) Method() *types.Func {
	return s.Value.(*ssa.Function).Object().(*types.Func)
}
func (s *StaticMethodCall) Arg(idx int) ssa.Value {
	return s.Args[idx+1]
}
func (s *StaticMethodCall) ArgsLen() int {
	return len(s.Args) - 1
}

type DynamicMethodCall struct {
	ssa.CallCommon
}

func NewDynamicMethodCall(common *ssa.CallCommon) *DynamicMethodCall {
	return &DynamicMethodCall{CallCommon: *common}
}

func (d *DynamicMethodCall) String() string {
	return d.Signature().String()
}
func (d *DynamicMethodCall) Name() string {
	return fmt.Sprintf("%s.%s", d.Recv().Type(), d.Method().Name())
}
func (d *DynamicMethodCall) Pkg() *types.Package {
	return d.Signature().Recv().Pkg()
}
func (d *DynamicMethodCall) Recv() ssa.Value {
	return d.Value
}
func (d *DynamicMethodCall) Method() *types.Func {
	return d.CallCommon.Method
}
func (d *DynamicMethodCall) Arg(idx int) ssa.Value {
	return d.Args[idx]
}
func (d *DynamicMethodCall) ArgsLen() int {
	return len(d.Args)
}

type BuiltinDynamicMethodCall struct {
	ssa.CallCommon
}

func NewBuiltinDynamicMethodCall(common *ssa.CallCommon) *BuiltinDynamicMethodCall {
	return &BuiltinDynamicMethodCall{CallCommon: *common}
}

func (b *BuiltinDynamicMethodCall) String() string {
	return b.Signature().String()
}
func (b *BuiltinDynamicMethodCall) Name() string {
	return fmt.Sprintf("%s.%s", b.Recv().Type(), b.Method().Name())
}
func (b *BuiltinDynamicMethodCall) Recv() ssa.Value {
	return b.Value
}
func (b *BuiltinDynamicMethodCall) Method() *types.Func {
	return b.CallCommon.Method
}
func (b *BuiltinDynamicMethodCall) Arg(idx int) ssa.Value {
	return b.Args[idx]
}
func (b *BuiltinDynamicMethodCall) ArgsLen() int {
	return len(b.Args)
}

type StaticFunctionCall struct {
	ssa.CallCommon
}

func NewStaticFunctionCall(common *ssa.CallCommon) *StaticFunctionCall {
	return &StaticFunctionCall{CallCommon: *common}
}

func (s *StaticFunctionCall) String() string {
	return s.Signature().String()
}
func (s *StaticFunctionCall) Name() string {
	return fmt.Sprintf("%s.%s", s.Pkg().Path(), s.Func().Name())
}
func (s *StaticFunctionCall) Pkg() *types.Package {
	return s.Func().Package().Pkg
}
func (s *StaticFunctionCall) Func() *ssa.Function {
	fn := s.Value.(*ssa.Function)
	if fn.Package() != nil {
		return fn
	}
	return fn.Origin() // generics static function call
}
func (s *StaticFunctionCall) Arg(idx int) ssa.Value {
	return s.Args[idx]
}
func (s *StaticFunctionCall) ArgsLen() int {
	return len(s.Args)
}

type BuiltinStaticFunctionCall struct {
	ssa.CallCommon
}

func NewBuiltinStaticFunctionCall(common *ssa.CallCommon) *BuiltinStaticFunctionCall {
	return &BuiltinStaticFunctionCall{CallCommon: *common}
}

func (b *BuiltinStaticFunctionCall) String() string {
	return b.Signature().String()
}
func (b *BuiltinStaticFunctionCall) Name() string {
	return b.Value.Name()
}
func (b *BuiltinStaticFunctionCall) Func() *ssa.Builtin {
	return b.Value.(*ssa.Builtin)
}
func (b *BuiltinStaticFunctionCall) Arg(idx int) ssa.Value {
	return b.Args[idx]
}
func (b *BuiltinStaticFunctionCall) ArgsLen() int {
	return len(b.Args)
}

type StaticFunctionClosureCall struct {
	ssa.CallCommon
}

func NewStaticFunctionClosureCall(common *ssa.CallCommon) *StaticFunctionClosureCall {
	return &StaticFunctionClosureCall{CallCommon: *common}
}

func (s *StaticFunctionClosureCall) String() string {
	return s.Signature().String()
}
func (s *StaticFunctionClosureCall) Name() string {
	return fmt.Sprintf("%s.%s", s.Pkg().Path(), s.Func().Name())
}
func (s *StaticFunctionClosureCall) Pkg() *types.Package {
	return s.Func().Package().Pkg
}
func (s *StaticFunctionClosureCall) Parent() *ssa.Function {
	return s.Value.Parent()
}
func (s *StaticFunctionClosureCall) Func() *ssa.Function {
	return s.Value.(*ssa.MakeClosure).Fn.(*ssa.Function)
}
func (s *StaticFunctionClosureCall) Arg(idx int) ssa.Value {
	return s.Args[idx]
}
func (s *StaticFunctionClosureCall) ArgsLen() int {
	return len(s.Args)
}

type DynamicFunctionCall struct {
	ssa.CallCommon
}

func NewDynamicFunctionCall(common *ssa.CallCommon) *DynamicFunctionCall {
	return &DynamicFunctionCall{CallCommon: *common}
}

func (d *DynamicFunctionCall) String() string {
	return d.Signature().String()
}
func (d *DynamicFunctionCall) Name() string {
	switch fn := d.Value.(type) {
	case *ssa.Call:
		return fn.Call.Value.Name()
	case *ssa.UnOp:
		return fn.X.Name()
	default:
		return fn.Name()
	}
}
func (d *DynamicFunctionCall) Arg(idx int) ssa.Value {
	return d.Args[idx]
}
func (d *DynamicFunctionCall) ArgsLen() int {
	return len(d.Args)
}

type CallInfo interface {
	marker()
	Name() string
	Arg(idx int) ssa.Value
	ArgsLen() int
}

func (s *StaticMethodCall) marker()          {}
func (d *DynamicMethodCall) marker()         {}
func (b *BuiltinDynamicMethodCall) marker()  {}
func (s *StaticFunctionCall) marker()        {}
func (b *BuiltinStaticFunctionCall) marker() {}
func (s *StaticFunctionClosureCall) marker() {}
func (d *DynamicFunctionCall) marker()       {}

func GetCallInfo(common *ssa.CallCommon) CallInfo {
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
