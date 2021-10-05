package wego

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
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
		*RouterGroup  //继承RouterGroup,将Engine抽象为最高层的RouterGroup
		router        *router
		groups        []*RouterGroup     //存储所有的groups
		htmlTemplates *template.Template //http模板
		funcMap       template.FuncMap   //html模板渲染函数
	}
)

//New 是wego.Engine的构造器
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}  //新建引擎所在的group
	engine.groups = []*RouterGroup{engine.RouterGroup} //将引擎所在的group加入groups中
	return engine
}

//Default 构造的engine使用默认的Logger与Recovery中间件
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
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

//Use 为当前组添加需要使用的中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		//只要存在组对应的前缀,则将组对应的中间件加入该上下文需要使用的中间件
		//组的嵌套使用中间件在此处实现
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	//封装后转交给router处理
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	//解析请求的地址，映射到服务器上文件的真实地址，交给http.FileServer处理
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filePath")
		//检查文件是否存在或是否有权限读取
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

//Static 将服务器上的静态资源映射到url中
//	Static("/assets", "/static")
//	访问/asserts/xxxfile 即可返回 /static/xxxfile
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	//注册GET处理器
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
