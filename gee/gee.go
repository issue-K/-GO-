package gee

import (
	"fmt"
	"net/http"
)


//处理请求函数原型
type HandlerFunc func( *Context )

type H map[string]string

//gee实例对象,使用map存储 路由路径:处理函数 的键值对

type Engine struct{
	router *router
}
//初始化gee实例对象
func New() *Engine{
	return &Engine{ router: newRouter() }
}
//向map中添加路由的函数
func (engine *Engine) addRoute(method string,pattern string,handler HandlerFunc){
	fmt.Printf("正在添加/login路由")
	engine.router.addRoute(method,pattern,handler)
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
	c := newContext(w,req)
	engine.router.handle(c)
}