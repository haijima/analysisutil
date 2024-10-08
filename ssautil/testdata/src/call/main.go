package main

import (
	"errors"
	"fmt"
)

func main() {
	staticMethod(Foo{name: "foo"})
	staticMethod2(Fizz{name: "fizz"})
	dynamicMethod(Foo{name: "foo"})
	dynamicMethod2(B{})
	builtinDynamicMethod(errors.New("error"))
	staticFunc()
	genericsStaticFunc()
	anonymousStaticFunc()
	builtinStaticFunc()
	staticFuncClosure()
	dynamicFunc(func() string { return "foo" })
	dynamicFunc2()
	dynamicFunc3()
}

type Foo struct {
	name string
}

func (f Foo) String() string {
	return f.name
}

func staticMethod(f Foo) {
	_ = f.String()
}

type Fizz struct {
	name string
}

func (f *Fizz) String() string {
	return f.name
}

func staticMethod2(f Fizz) {
	_ = f.String()
}

func dynamicMethod(f fmt.Stringer) {
	_ = f.String()
}

type Barer interface {
	Bar() string
}

type B struct{}

func (b B) Bar() string { return "B" }

func dynamicMethod2(f Barer) {
	_ = f.Bar()
}

func builtinDynamicMethod(err error) {
	_ = err.Error()
}

func staticFunc() {
	fmt.Println("staticFunc")
}

type Stringer struct {
	name string
}

func (s Stringer) String() string {
	return s.name
}

func foo[T fmt.Stringer](t T) string {
	return t.String()
}

func genericsStaticFunc() {
	_ = foo[Stringer](Stringer{name: "foo"})
}

func anonymousStaticFunc() {
	func() {
		// do nothing
	}()
}

func builtinStaticFunc() {
	_ = append([]int{1, 2, 3}, 4)
}

func staticFuncClosure() {
	count := 0
	func() int {
		count++
		return count
	}()
}

func dynamicFunc(fn func() string) {
	_ = fn()
}

var callableVar = func() string { return "foo" }

func dynamicFunc2() {
	_ = callableVar()
}

func getCallable() func(num int) int {
	return func(num int) int { return num }
}

func dynamicFunc3() {
	c := getCallable() // static function call
	_ = c(0)           // dynamic function call
}
