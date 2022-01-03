## Day 3

**实现动态路由.**

考虑形如```/p/:name```的这样一条路由,其中```:name```表示通配符,接受任何字段并作为参数```name```的值

比如能匹配上```/p/cl```,此时```name="cl"```

如果还采用之前的结构```map[string]HandlerFunc```就无法完成匹配

所以还需要改进一下我们的匹配路由机制,使用```trie```树来完成更优秀的匹配.

```go
package gee

import "strings"

/*
存储路由的结构.普通的map[string]handlefunc只能查找静态路由
这里使用trie实现一个动态路由,支持:name 和 *name 两种匹配规则
*/
type node struct{
	pattern string //待匹配路由(从根节点到本节点的路由路径,形如/p/:name)
	part string //本节点中存储路由的一部分,如:name
	children []*node //子节点
	isWild bool //part中是否含有:和*,有则为true
}

//查找子节点中第一个匹配成功的节点
func (n *node) matchChild(part string) *node{
	for _,child := range n.children{
		if child.part == part || child.isWild{  //路由完全吻合或路由中存在通配符:,*
			return child
		}
	}
	return nil
}
//查找所有匹配成功的节点,返回节点数组
func (n *node) matchChildren(part string) []*node{
	nodes := make([]*node,0)
	for _,child:= range n.children{
		if child.part == part || child.isWild{
			nodes = append( nodes,child )
		}
	}
	return nodes
}
/* insert
pattern表示需要插入的路由路径,形如/p/:name
parts就是把pattern根据/拆开来形成的数组
dep表示当前在trie树的深度
*/
func (n *node) insert( pattern string,parts []string,dep int){
	if len(parts) == dep{ ///匹配成功
		n.pattern = pattern
		return
	}
	part := parts[dep] //在下一层需要匹配的字符串
	child := n.matchChild(part)
	if child == nil{ //没有该子节点,就新建一个
		child = &node{ part:part,isWild:part[0]==':' || part[0]=='*' }
		n.children = append( n.children,child )
	}
	child.insert( pattern,parts,dep+1 )
}
func (n *node) search( parts []string,dep int) * node{
	if len(parts)==dep || strings.HasPrefix(n.part,"*"){
		if n.pattern==""{
			return nil
		}
		return n
	}
	part := parts[dep]
	children := n.matchChildren(part)

	for _,child := range children{
		result := child.search( parts,dep+1 )
		if result!=nil{
			return result
		}
	}
	return nil
}
```

### Context.go

略微修改我们的```Context```结构,因为现在是动态路由,设置一个保留参数的```map```进去

```go
type Context struct{
	Writer http.ResponseWriter
	Req *http.Request

	Path string
	Method string
	Params map[string]string  //利用trie动态路由解析出来的参数

	StatusCode int
}
func (c *Context) Param(key string) string{  //获取参数的方法
	value,_ := c.Params[key]
	return value
}
```

### Route.go

之前直接使用的```map[string]HandlerFunc```查找路由

现在加入了有参数的路由,所以需要使用```trie```树来查找,所以新建一个```map[string]*node```用来保存各个请求方法($get,post$等)对应```trie```树的根节点

先从入口函数看起

```go
func (r *router)handle(c *Context){  //需要处理请求了
	n,params := r.getRoute(c.Method,c.Path)  //得到原生url后给getRoute进行解析
    //返回值n表示匹配到的对应trie树节点,params是参数列表
	if n!=nil{
		c.Params = params 
		key := c.Method + "-" +n.pattern
		r.handlers[key](c) //调用对应处理函数
	}else{
		c.String( http.StatusNotFound,"404 NOT FOUND: %s\n",c.Path ) //报404错
	}
}
```

再看整个代码就很清晰了

```go
package gee

import (
	"net/http"
	"strings"
)

type router struct{
	roots map[string]*node //trie树,roots["GET"]表示GET请求的trie树
	handlers map[string]HandlerFunc
}
func newRouter() *router{
	return &router{
		roots: make( map[string]*node ),
		handlers: make( map[string]HandlerFunc ) ,
	}
}

func parsePattern(pattern string) []string { //拿到原生url后对/进行分割返回数组
	 vs := strings.Split( pattern,"/" )
	 parts := make([]string,0)
	 for _,item := range vs{
	 	if item !=""{
	 		parts = append( parts,item )
	 		if item[0] == '*'{
	 			break
			}
		}
	 }
	 return parts
}

//addRoute 添加路由
func (r *router) addRoute(method string,pattern string,handler HandlerFunc){
	parts := parsePattern( pattern )
	key := method + "-" + pattern
	_,ok := r.roots[method]
	if !ok{ //没有这个方法对应的trie树
		r.roots[method] = &node{}
	}
	r.roots[method].insert( pattern,parts,0 )
	r.handlers[key] = handler
}
//根据请求路径拿到对应的
func(r *router)getRoute(method string,path string)(*node,map[string]string ){
	searchParts := parsePattern( path )  //用户的请求路径
	params := make( map[string]string )  //把路由中对应的参数解析出来
	root,ok := r.roots[method]

	if !ok{
		return nil,nil
	}
	n := root.search( searchParts,0 )
	if n!=nil{
		parts := parsePattern(n.pattern)  //得到匹配成功的路由
		for index,part := range parts{
			if part[0] == ':'{
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part)>1{
				params[part[1:]] = strings.Join( searchParts[index:],"/" )
				break
			}
		}
		return n,params
	}
	return nil,nil
}
func (r *router)handle(c *Context){
	n,params := r.getRoute(c.Method,c.Path)
	if n!=nil{
		c.Params = params
		key := c.Method + "-" +n.pattern
		r.handlers[key](c)
	}else{
		c.String( http.StatusNotFound,"404 NOT FOUND: %s\n",c.Path )
	}
}
```

最后来测试一下吧

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

	r.GET("/hello/:name", func(c *gee.Context) {
		// expect /hello/geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}
```

