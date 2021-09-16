package wego

import (
	"fmt"
	"net/http"
)

//HandlerFunc 被引擎使用的请求处理器的类型
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

//Engine 实现了ServeHTTP的接口
type Engine struct {
	router map[string]HandlerFunc
}

//New 是wego.Engine的构造器
func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	//将请求方法与路径合并为选定路由的键
	key := method + "-" + pattern
	engine.router[key] = handler
}

//GET 定义了添加GET请求的方法
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

//POST 定义了添加POST请求的方法
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//解析请求方法与路径
	key := req.Method + "-" + req.URL.Path

	//查找路由映射表
	//handler, ok := engine.router[key];
	//if ok { ... }
	if handler, ok := engine.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
