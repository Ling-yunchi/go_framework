package wego

import (
	"log"
	"net/http"
)

//HandlerFunc 被引擎使用的请求处理器的类型
type HandlerFunc func(*Context)

//Engine 实现了ServeHTTP的接口
type (
	RouterGroup struct {
		prefix      string        //支持嵌套
		middlewares []HandlerFunc //支持中间件
		engine      *Engine       //所有的组使用同一个Engine实例
	}

	Engine struct {
		*RouterGroup //继承RouterGroup,将Engine抽象为最高层的RouterGroup
		router       *router
		groups       []*RouterGroup //存储所有的groups
	}
)

//New 是wego.Engine的构造器
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}  //新建引擎所在的group
	engine.groups = []*RouterGroup{engine.RouterGroup} //将引擎所在的group加入groups中
	return engine
}

//Group 用于在当前group下创建一个新的子RouterGroup
//所有的group共享同一个 Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

//addRoute 内部添加Route接口,不向外暴露
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

//GET 定义了添加GET请求的方法
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

//POST 定义了添加POST请求的方法
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//封装后转交给router处理
	c := newContext(w, req)
	engine.router.handle(c)
}
