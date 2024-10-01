package bttp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]any

type Context struct {
	// original objs
	Writer http.ResponseWriter
	Req    *http.Request
	// filled by request
	Path   string
	Method string
	Params map[string]string
	// filled by resp
	StatusCode int
	// middleware related
	idx         int
	middlewares []Handler
}

func (c *Context) Next() {
	c.idx++
	for c.idx < len(c.middlewares) {
		c.middlewares[c.idx](c)
		c.idx++
	}
}

func (c *Context) Abort() {
	c.idx = len(c.middlewares)
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		idx:    -1,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		// 这里实际并不会生效，header 和 statusCode 只能被修改一次
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) InternalError(msg string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(http.StatusInternalServerError)
	c.Writer.Write([]byte(msg))

}
