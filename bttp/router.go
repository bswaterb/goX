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
	if n != nil {
		c.Params = params
	}
	key := c.Method + "-" + n.pattern
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
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
