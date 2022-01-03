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

func parsePattern(pattern string) []string {
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