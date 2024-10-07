package brpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Student struct {
	Name string
	Age  int
}

func (s *Student) GetName(args Args, resp *string) error {
	*resp = s.Name
	return nil
}

func (s *Student) getAge(args Args, resp *int) error {
	*resp = s.Age
	return nil
}

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// it's not an exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	stu := Student{
		Name: "test",
		Age:  18,
	}
	s := newService(&stu)
	_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["GetName"]
	_assert(mType != nil, "wrong Method, Sum shouldn't nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgV()
	respV := mType.newRespV()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, respV)
	_assert(err == nil && *respV.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}
