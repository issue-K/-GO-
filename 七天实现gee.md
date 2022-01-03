## Day 1

```net/http```标准库启动```web```服务

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

// handler echoes r.URL.Path
func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

// handler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}
```

使用```http.ListenAndServe```启动服务,这个函数的原型是

```go
func ListenAndServe(addr string, handler Handler) error
```

第一个参数是```ip+端口```,第二个参数是```Handler```类型,其实这个类型是一个接口

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

所有的请求最后都会交付给这个接口处理,所以第一步我们就可以先实现```Handler```接口

- 自定义一个结构体名为```engine```实现```Handler```接口

- 标准库中绑定路由的方法形如

  ```go
  http.HandleFunc("/hello", helloHandler)
  ```

  所以我们也需要提供绑定路由的接口```GET```和```POST```

  在```engine```类型中设置一个存储键为路由路径,值为处理函数的数据结构(这里用```map```实现)

- 实现```ServeHTTP```函数。由于所有请求都会发送到这个接口,于是可以取出路径,找到对应的处理函数,调用对应的处理函数

- 最后使```engine```生效,也就是

  ```go
  func (engine *Engine ) Run(addr string) (err error){
  	return http.ListenAndServe( addr,engine )
  }
  ```

  ```gee```框架如下

```go
package gee

import(
	"fmt"
	"net/http"
)

//处理请求函数原型
type HandlerFunc func( http.ResponseWriter,*http.Request )
//gee实例对象,使用map存储 路由路径:处理函数 的键值对
type Engine struct{
	router map[string]HandlerFunc
}
//初始化gee实例对象
func New() *Engine{
	return &Engine{ router: make( map[string]HandlerFunc) }
}
//向map中添加路由的函数
func (engine *Engine) addRoute(method string,pattern string,handler HandlerFunc){
	key := method + "-" + pattern
	engine.router[key] = handler
}
func (engine *Engine) GET(pattern string,handler HandlerFunc){
	engine.addRoute("GET",pattern,handler )
}
func (engine *Engine) POST(pattern string,handler HandlerFunc){
	engine.addRoute("POST",pattern,handler )
}

func (engine *Engine ) Run(addr string) (err error){
	return http.ListenAndServe( addr,engine )
}
func (engine *Engine ) ServeHTTP(w http.ResponseWriter,req *http.Request){
	key := req.Method + "-" + req.URL.Path
	if handler,ok := engine.router[key]; ok{
		handler(w,req ) //存在此条路由,调用对应的handler函数
	}else{
		fmt.Fprintf(w,"404 NOT FOUND: %s\n",req.URL )//不存在此条路由
	}
}
```

测试代码

```go
package main

import (
	"fmt"
	"net/http"

	"gee"
)

func main() {
	r := gee.New()
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	r.Run(":9999")
}
```

