package brpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bswaterb/goX/brpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	MagicNumber = 0x123abc
)

// Option
// this filed is the start in conn msg
// [<- Option ->][<- Header ->][<- Body ->][<- Header ->][<- Body ->]......
// part of [<- Option ->] is codec by json
// part of [<- Header ->] and [<- Body ->] is codec by Option.CodecType
type Option struct {
	MagicNumber int
	CodecType   codec.Type

	ConnectTimeout time.Duration // 0 means no limit
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.TypeGob,
	ConnectTimeout: time.Second * 10,
}

var DefaultServer = NewServer()

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

// Register publishes in the server the set of methods of the
func (server *Server) Register(regObj any) error {
	s := newService(regObj)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(regObj any) error { return DefaultServer.Register(regObj) }

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

// Accept accepts connections on the listener and serves requests
func (server *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func Accept(listener net.Listener) { DefaultServer.Accept(listener) }

// ServeConn  handle the incoming conn
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var opt Option

	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		fmt.Println("brpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		fmt.Printf("brpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	codecFuncMap := codec.GetCodecFuncMap()
	f := codecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	var sending sync.Mutex // make sure to send a complete response
	var wg sync.WaitGroup  // wait until all request are handled

	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Err = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, &sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, &sending, &wg, time.Second*10)
	}
	wg.Wait()
	_ = cc.Close()
}

// request stores all information of a call
type request struct {
	h           *codec.Header // header of request
	argV, respV reflect.Value // argV and respV of request
	mtype       *methodType
	svc         *service
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Printf("rpc server: read header error:%s, header:%#v\n", err, h)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argV = req.mtype.newArgV()
	req.respV = req.mtype.newRespV()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argV.Interface()
	if req.argV.Type().Kind() != reflect.Ptr {
		argvi = req.argV.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		fmt.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argV, req.respV)
		called <- struct{}{}
		if err != nil {
			req.h.Err = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.respV.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	ddl := time.After(timeout)
timeoutCheck:
	select {
	case <-ddl:
		req.h.Err = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		select {
		case <-ddl:
			req.h.Err = fmt.Sprintf("rpc server: called successfully but sent timeout: expect within %s", timeout)
			server.sendResponse(cc, req.h, invalidRequest, sending)
		case <-sent:
			break timeoutCheck
		}
	}
}
