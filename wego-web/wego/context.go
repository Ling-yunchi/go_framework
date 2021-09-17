package wego

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//H 为map[string]interface{}起的别名wego.H
//构建JSON数据时更加简洁
type H map[string]interface{}

//Context 为上下文,封装请求信息
type Context struct {
	//封装原有项目
	Writer http.ResponseWriter
	Req    *http.Request
	//请求信息
	Path   string
	Method string
	Params map[string]string
	//返回信息
	StatusCode int
	//中间件
	handlers []HandlerFunc
	index    int
}

//newContext 是 Context 的构造器
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1, //中间件执行位置,初始化为-1
	}
}

//Next 开始执行c所包含的中间件
func (c *Context) Next() {
	//关于为什么要把中间件执行的index保存在c中:
	//	1.每个中间件的格式基本上为:
	//	func A (c *Context){
	//		part1
	//      c.Next()
	//		part2
	//	}
	//	若不将中间件执行到的index存储在c中,调用c.Next()将会陷入无限嵌套
	//	2.不是所有中间件都会显式调用c.Next(),使用for循环兼容性更强!
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

//Fail 中断中间件的执行,使后面的中间件不再继续执行,并返回错误信息
func (c *Context) Fail(code int, err string) {
	log.Printf("Handler fail at %s handlers[%d] : %s", c.Path, c.index, err)
	c.JSON(code, H{"message": err})
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//Query 查询url中的参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

//Status 设定Context的状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

//SetHeader 设定Context的返回头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

//String 返回一个格式化字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

//JSON 返回json格式的对象
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	//编码器
	encoder := json.NewEncoder(c.Writer)
	defer func() {
		if err := recover(); err != nil {
			http.Error(c.Writer, "Webserver Error: "+err.(string), 500)
		}
	}()
	//Context.ResponseWriter中的Set/WriteHeader/Write这三个方法时
	//顺序必须为Set/WriteHeader/Write
	//此处错误处理很不方便,无法重新设置状态码
	//Gin中直接抛出panic
	if err := encoder.Encode(obj); err != nil {
		panic(err.Error())
	}
}

//Data 返回字符数组类型的数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
