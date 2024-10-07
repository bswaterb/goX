package brpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type service struct {
	name    string
	objType reflect.Type
	regObj  reflect.Value
	method  map[string]*methodType
}

func newService(regObj any) *service {
	s := new(service)
	s.regObj = reflect.ValueOf(regObj)
	s.name = reflect.Indirect(s.regObj).Type().Name()
	s.objType = reflect.TypeOf(regObj)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.objType.NumMethod(); i++ {
		method := s.objType.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, respType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(respType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:   method,
			ArgType:  argType,
			RespType: respType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv, respV reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.regObj, argv, respV})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

type methodType struct {
	method   reflect.Method
	ArgType  reflect.Type
	RespType reflect.Type
	// 统计该方法的累计调用次数
	numCalls uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgV() reflect.Value {
	var argv reflect.Value
	// arg may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newRespV() reflect.Value {
	// reply must be a pointer type
	respV := reflect.New(m.RespType.Elem())
	switch m.RespType.Elem().Kind() {
	case reflect.Map:
		respV.Elem().Set(reflect.MakeMap(m.RespType.Elem()))
	case reflect.Slice:
		respV.Elem().Set(reflect.MakeSlice(m.RespType.Elem(), 0, 0))
	default:
		// do nothing

	}
	return respV
}
