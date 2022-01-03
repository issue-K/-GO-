package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Context struct{
	Writer http.ResponseWriter
	Req *http.Request //请求头
	//请求参数
	Path string
	Method string
	Params map[string]string  //利用trie动态路由解析出来的参数
	//响应参数
	StatusCode int
}
func newContext(w http.ResponseWriter,req *http.Request) *Context{
	return &Context{
		Writer: w,
		Req: req,
		Path: req.URL.Path,
		Method: req.Method,
	}
}
/*
下面的方法是获取参数的方法
 */
func (c *Context) PostForm(key string) string{
	return c.Req.FormValue(key)  //post请求获取参数
}
func (c *Context ) Query(key string) string{
	return c.Req.URL.Query().Get(key) //get请求获取参数
}
func (c *Context) Param(key string) string{
	value,_ := c.Params[key]
	return value
}
func (c *Context) Status(code int){
	c.StatusCode = code;
	c.Writer.WriteHeader( code );
}
func (c *Context) SetHeader(key string,value string){
	c.Writer.Header().Set(key,value)
}
func (c *Context) String(code int,format string,values ...interface{} ){
	c.SetHeader("Content-Type","text/plain")
	c.Status(code)
	c.Writer.Write( []byte( fmt.Sprintf(format,values...)) )
}

func (c *Context) JSON(code int,obj interface{}){
	c.SetHeader("Content-Type","application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer) //使用c.Writer输出流初始化json构造器
	if err := encoder.Encode(obj); err!=nil{
		http.Error(c.Writer,err.Error(),500 )
	}
}
func (c *Context) Data(code int,data []byte){
	c.Status(code)
	c.Writer.Write(data)
}
func (c *Context) HTML(code int,html string){
	c.SetHeader("Content-Type","text/html")
	c.Status( code )
	c.Writer.Write([]byte(html) )
}
