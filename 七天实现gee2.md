## Day 2

本次新增内容

- 将`路由(router)`独立出来，方便之后增强。
- 设计`上下文(Context)`，封装 ```Request``` 和 ```Response``` ，提供对 ```JSON```、```HTML``` 等返回类型的支持。

本次测试代码

```go
package main

import (
	"gee"
	"net/http"
)

func main() {
	r := gee.New()
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})
	r.GET("/hello", func(c *gee.Context) {
		// expect /hello?name=geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.POST("/login", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	r.Run(":9999")
}
```

上下文```Context```结构

```go
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	// response info
	StatusCode int
}
```

- `Context`目前只包含了`http.ResponseWriter`和`*http.Request`，另外提供了对``` Method``` 和 ```Path``` 这两个常用属性的直接访问。

- 提供了访问```Query```和```PostForm```参数的方法。

- 提供了快速构造```String/Data/JSON/HTML```响应的方法。

  ```go
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
  func (c *Context) PostForm(key string) string{
  	return c.Req.FormValue(key)  //post请求获取参数
  }
  func (c *Context ) Query(key string) string{
  	return c.Req.URL.Query().Get(key) //get请求获取参数
  }
  func (c *Context) Status(code int){
  	c.StatusCode = code;
  	c.Writer.WriteHeader( code );
  }
  func (c *Context) SetHeader(key string,value string){
  	c.Writer.Header().Set(key,value)
  }
  func (c *Context) String(code int,format string,values ...interface{} ){
  	c.SetHeader("Context-Type","text/plain")
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
  ```

  再把关于路由```Router```的功能抽离出来

  ```go
  type router struct {
  	handlers map[string]HandlerFunc
  }
  
  func newRouter() *router {
  	return &router{handlers: make(map[string]HandlerFunc)}
  }
  
  func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
  	log.Printf("Route %4s - %s", method, pattern)
  	key := method + "-" + pattern
  	r.handlers[key] = handler
  }
  
  func (r *router) handle(c *Context) {
  	key := c.Method + "-" + c.Path
  	if handler, ok := r.handlers[key]; ok {
  		handler(c)
  	} else {
  		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
  	}
  }
  ```

  

最后略微修改框架主入口

```go
// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
}

// New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{router: newRouter()}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
```

