package bttp

import (
	"net/http"
	"strings"
)

type Handler func(c *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func NewEngine() *Engine {
	e := &Engine{
		router: newRouter(),
	}
	e.RouterGroup = &RouterGroup{engine: e}
	e.groups = []*RouterGroup{e.RouterGroup}

	return e
}

func (e *Engine) addRoute(method string, path string, handler Handler) {
	e.router.addRoute(method, path, handler)
}

func (e *Engine) GET(pattern string, handler Handler) {
	e.addRoute("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler Handler) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	middlewares := make([]Handler, 0, 5)
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.middlewares = middlewares
	e.router.handle(c)
}

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}
