## Day 4

实现```gin```中的路由分组功能

**分组**: 若一组路由需要经过相似的处理,可以把这些路由分为一组,给这个组加上对应的中间件函数,这样就可以达到复用代码的效果,容易扩展.

通常采用相同前缀来区分 分组

使用方法如下(最后的测试代码)

```go
func main() {
	r := gee.New()
	r.GET("/index", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *gee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *gee.Context) {
			// expect /hello?name=geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *gee.Context) {
			c.JSON(http.StatusOK, gee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
}
```

- 可以发现,这个```Group```对象拥有```.GET,.POST```等方法,所以```Group```对象应该具有访问```Route```的能力

我们可以在```Group```中保存一个```Engine```指针,就可以通过这个指针调用对应方法了

- 也需要拥有前缀,比如```/```,```/api```

- 支持分组嵌套,那么需要知道当前分组的父亲是谁,于是还需要有一个指向父亲的指针

- [中间件函数]数组。

  于是```Group```对象大概长这个样子

  ```go
  RouterGroup struct {
  	prefix      string
  	middlewares []HandlerFunc // support middleware
  	parent      *RouterGroup  // support nesting
  	engine      *Engine       // all groups share a Engine instance
  }
  ```

  观察以下路由嵌套代码

  ```go
  r := gee.New()
  v1 := r.Group("/v1")
  {
      v2 := v1.Group("/v2")
      {
          //此处代码省略
      }
  }
  ```

  考虑到$v_1$和$r$具有相似的函数和功能,不妨让```engine​```继承自```RouterGroup```

  ```go
  Engine struct{
      router *router
      *RouterGroup //继承RouterGroup
      groups []*RouterGroup //保存所有组
  }
  ```

  让我们来实现```Group```函数

  ```go
  func (group *RouterGroup) Group (prefix string) *RouterGroup{
  	engine := group.engine //获取engine
  	newGroup := &RouterGroup{ 
  		prefix: group.prefix+prefix, //父亲前缀+自己本身的路径
  		parent:group, //父亲
  		engine: engine,  //engine
  	}
  	engine.groups = append( engine.groups,newGroup ) //把所有组保存起来
  	return newGroup  //返回
  }
  ```

  然后```RouterGroup```还支持```GET,POST```方法,于是改写一下之前的代码

  ```go
  func (group *RouterGroup) addRoute(method string,path string,handler HandlerFunc){
  	pattern := group.prefix + path  //计算真实的路径
  	log.Printf("Route %4s - %s", method, pattern)
  	group.engine.router.addRoute(method,pattern,handler)
  }
  func (group *RouterGroup) GET(pattern string,handler HandlerFunc){
  	group.addRoute("GET",pattern,handler )
  }
  func (group *RouterGroup) POST(pattern string,handler HandlerFunc){
  	group.addRoute("POST",pattern,handler )
  }
  ```

  