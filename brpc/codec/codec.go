package codec

import (
	"io"
	"sync"
)

const (
	TypeGob  = "application/gob"
	TypeJson = "application/json"
)

var (
	TypeGobBytes = []byte("application/gob#")
)

var defaultCodecFuncMap map[Type]NewCodecFunc
var once sync.Once

func GetCodecFuncMap() map[Type]NewCodecFunc {
	once.Do(func() {
		defaultCodecFuncMap = make(map[Type]NewCodecFunc)
		defaultCodecFuncMap[TypeGob] = NewGobCodec
	})
	return defaultCodecFuncMap
}

type Type string

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Codec interface {
	io.Closer
	// ReadHeader 从 io.Reader 中读取协议头
	ReadHeader(*Header) error
	// ReadBody 从 io.Reader 中读取请求体
	ReadBody(interface{}) error
	// Write 向 io.Writer 写入响应的 header 和 body
	Write(*Header, interface{}) error

	GetCC() io.ReadWriteCloser
}

type Header struct {
	// 请求的具体服务与方法
	ServiceMethod string
	// 请求序号
	Seq uint64
	// 被调用方报错
	Err string
}
