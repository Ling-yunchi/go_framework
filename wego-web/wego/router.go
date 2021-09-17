package wego

import (
	"net/http"
	"strings"
)

type router struct {
	//roots 存储每种请求方式的Trie树根节点
	roots map[string]*node
	//handlers 存储每种请求方式的 HandlerFunc
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

//parsePattern 解析url的模式
func parsePattern(pattern string) []string {
	//将url以 / 为分隔符分割
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	//将分割完的非空字符串保存
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			//只允许有一个*
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

//addRoute 添加路由规则
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	if _, ok := r.roots[method]; !ok {
		//检查是否有method对应的子节点,若没有就创建一个
		r.roots[method] = &node{}
	}
	//在子节点上插入
	r.roots[method].insert(pattern, parts, 0)
	//设定处理器
	r.handlers[key] = handler
}

//getRoute 获取路由规则
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
	}
	return n, params
}

func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		//将最终处理请求的handler加入c的handler列表中
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(context *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	//开始执行handlers
	c.Next()
}
