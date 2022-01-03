package gee

import (
	"log"
	"net/http"
)


//处理请求函数原型
type HandlerFunc func( *Context )

type H map[string]string

//gee实例对象,使用map存储 路由路径:处理函数 的键值对
type(
	Engine struct{
		router *router
		*RouterGroup
		groups []*RouterGroup //保存所有组
	}
	RouterGroup struct{
		prefix string //前缀
		middlewares []HandlerFunc //中间件
		parent *RouterGroup //父亲指针(支持路由嵌套)
		engine *Engine
	}
)
//初始化gee实例对象
func New() *Engine{
	engine := &Engine{ router:newRouter() }
	engine.RouterGroup = &RouterGroup{ engine:engine }
	engine.groups = []*RouterGroup{ engine.RouterGroup }
	return engine
}
/*Group
根据父亲的信息构造一个新的 RouterGroup
*/
func (group *RouterGroup) Group (prefix string) *RouterGroup{
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix+prefix,
		parent:group,
		engine: engine,
	}
	engine.groups = append( engine.groups,newGroup )
	return newGroup
}
//向map中添加路由的函数
func (group *RouterGroup) addRoute(method string,path string,handler HandlerFunc){
	pattern := group.prefix + path
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method,pattern,handler)
}
func (group *RouterGroup) GET(pattern string,handler HandlerFunc){
	group.addRoute("GET",pattern,handler )
}
func (group *RouterGroup) POST(pattern string,handler HandlerFunc){
	group.addRoute("POST",pattern,handler )
}

func (engine *Engine ) Run(addr string) (err error){
	return http.ListenAndServe( addr,engine )
}
func (engine *Engine ) ServeHTTP(w http.ResponseWriter,req *http.Request){
	c := newContext(w,req)
	engine.router.handle(c)
}