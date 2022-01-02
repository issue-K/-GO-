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