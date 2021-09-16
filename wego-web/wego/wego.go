package wego

import (
	"net/http"
)

//HandlerFunc 被引擎使用的请求处理器的类型
type HandlerFunc func(*Context)

//Engine 实现了ServeHTTP的接口
type Engine struct {
	router *router
}

//New 是wego.Engine的构造器
func New() *Engine {
	return &Engine{router: newRouter()}
}

//addRoute 内部添加Route接口,不向外暴露
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
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
	//封装后转交给router处理
	c := newContext(w, req)
	engine.router.handle(c)
}
