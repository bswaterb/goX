package bttp

import (
	"net/http"
	"strings"
)

type router struct {
	// rootsMap key: http-method, value: trie-root-node
	rootsMap map[string]*treeNode
	handlers map[string]Handler
}

func newRouter() *router {
	return &router{
		rootsMap: make(map[string]*treeNode),
		handlers: make(map[string]Handler),
	}
}

func (r *router) addRoute(method string, pattern string, handler Handler) {
	subPath := parsePattern(pattern)
	root := r.rootsMap[method]
	if root == nil {
		root = &treeNode{}
		root.insert(pattern, subPath, 0)
		r.rootsMap[method] = root
	} else {
		root.insert(pattern, subPath, 0)
	}
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	var key string
	if n != nil {
		c.Params = params
		key = c.Method + "-" + n.pattern
	}

	if handler, ok := r.handlers[key]; ok && n != nil {
		c.middlewares = append(c.middlewares, handler)
	} else {
		c.middlewares = append(c.middlewares, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}

func (r *router) getRoute(method string, pattern string) (*treeNode, map[string]string) {
	searchParts := parsePattern(pattern)
	params := make(map[string]string)
	root, ok := r.rootsMap[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if len(part) > 0 && part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if len(part) > 0 && part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func parsePattern(pattern string) []string {
	//if len(pattern) > 0 && strings.HasPrefix(pattern, "/") {
	//	pattern = pattern[1:]
	//}
	sub := strings.Split(pattern, "/")
	subPath := make([]string, 0, len(sub))
	for _, item := range sub {
		if len(item) == 0 {
			continue
		}
		subPath = append(subPath, item)
		if len(item) > 0 && item[0] == '*' {
			break
		}
	}
	return subPath
}

type RouterGroup struct {
	prefix      string
	middlewares []Handler
	parent      *RouterGroup
	engine      *Engine
}

func (rg *RouterGroup) Group(prefix string) *RouterGroup {
	e := rg.engine
	newGroup := &RouterGroup{
		prefix:      e.prefix + prefix,
		middlewares: make([]Handler, 0),
		parent:      rg,
		engine:      e,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

func (rg *RouterGroup) Use(middlewares ...Handler) {
	rg.middlewares = append(rg.middlewares, middlewares...)
}

func (rg *RouterGroup) addRoute(method string, pattern string, handler Handler) {
	fullPattern := rg.prefix + pattern
	rg.engine.router.addRoute(method, fullPattern, handler)
}

func (rg *RouterGroup) GET(pattern string, handler Handler) {
	rg.addRoute("GET", pattern, handler)
}

func (rg *RouterGroup) POST(pattern string, handler Handler) {
	rg.addRoute("POST", pattern, handler)
}
