package wego

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//H 为map[string]interface{}起的别名wego.H
//构建JSON数据时更加简洁
type H map[string]interface{}

//Context 为上下文,封装请求信息
type Context struct {
	//封装原有项目
	Witer http.ResponseWriter
	Req   *http.Request
	//请求信息
	Path   string
	Method string
	//返回信息
	StatusCode int
}

//NewContext 是 Context 的构造器
func NewContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Witer:  w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

//Status 设定Context的状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Witer.WriteHeader(code)
}

//SetHeader 设定Context的返回头
func (c *Context) SetHeader(key string, value string) {
	c.Witer.Header().Set(key, value)
}

//String 返回一个格式化字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Witer.Write([]byte(fmt.Sprintf(format, values...)))
}

//JSON 返回json格式的对象
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)

	//编码器
	encoder := json.NewEncoder(c.Witer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Witer, err.Error(), 500)
	}
}

//Data 返回字符数组类型的数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Witer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Witer.Write([]byte(html))
}
