## Day 5

### 中间件设计

中间件是一组用户自定义的函数.

在处理请求时,调用用户自定义的```Handler```之后,再调用中间件函数做一些额外操作

事实上完全可以把要做的操作直接写在```Handler```中,不过因为引入了组的概念,完全可以把一个组内都需要的操作写成中间件复用,这样既简洁清晰又易于扩展

**当收到一个请求时,处理步骤大概如下**

先初始化一个```context```对象

再把用户自定义的中间件函数加入```context```对象

调用用户对此路径的```Handler```函数

依次调用中间件函数

#### Context.Next()方法

这个方法的作用是,暂时先跳过当前正在执行的中间件函数,先往后执行,执行完了之后再回来执行当前中间件函数

于是需要改动一下```context.go```函数

```go
type Context struct{
	Writer http.ResponseWriter
	Req *http.Request //请求头
	//请求参数
	Path string
	Method string
	Params map[string]string  //利用trie动态路由解析出来的参数
	//响应参数
	StatusCode int
	//中间件
	handlers []HandlerFunc
	index int //记录当前执行到第几个中间件
}

func newContext(w http.ResponseWriter,req *http.Request) *Context{
	return &Context{
		Writer: w,
		Req: req,
		Path: req.URL.Path,
		Method: req.Method,
		index: -1,
	}
}

func (c *Context) Next(){
	c.index++
	le := len( c.handlers )
	for ; c.index<le ;c.index++{  //依次调用之后的函数
		c.handlers[c.index](c)
	}
}
```

### 给组加入中间件

修改文件```gee.go```

```go
//gee实例对象,使用map存储 路由路径:处理函数 的键值对
type(
	Engine struct{
		router *router
		*RouterGroup
		groups []*RouterGroup 
	}
	RouterGroup struct{
		prefix string
		middlewares []HandlerFunc //新增属性,作用在这个组上的中间件数组
		parent *RouterGroup 
		engine *Engine
	}
)
//向group中加入中间件函数
func (group *RouterGroup) Use(middlewares ...HandlerFunc){
	group.middlewares = append( group.middlewares,middlewares... )
}
//处理请求的入口,在这里需要把请求所属的组的所有中间件函数找出来赋给context
func (engine *Engine ) ServeHTTP(w http.ResponseWriter,req *http.Request){
	var middlewares []HandlerFunc
	for _,group := range engine.groups{  //简单的判断:遍历所有组,是这个组的话就把那个组的中间件加进来
		if strings.HasPrefix( req.URL.Path,group.prefix ){
			middlewares = append( middlewares,group.middlewares... )
		}
	}

	c := newContext(w,req)
	c.handlers = middlewares
	engine.router.handle(c)
}
```

在调用完用户的```handler```之后开始调用中间件函数

修改下文件```router.go```

```go
func (r *router)handle(c *Context){
	n,params := r.getRoute(c.Method,c.Path)
	if n!=nil{
		c.Params = params
		key := c.Method + "-" +n.pattern
		r.handlers[key](c)
	}else{
		c.handlers = append(c.handlers,func(c *Context){ //报错信息加入中间件,最后输出
			c.String( http.StatusNotFound,"404 NOT FOUND: %s\n",c.Path )
		})
	}
	c.Next()  //开始处理中间件函数
}
```



---

最后测试下代码

```go
package main

import (
	"gee"
	"log"
	"net/http"
	"time"
)

func onlyForV2() gee.HandlerFunc {
	return func(c *gee.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := gee.New()
	r.Use(gee.Logger()) // global midlleware
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":9999")
}
```

