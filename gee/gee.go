package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)


//处理请求函数原型
type HandlerFunc func( *Context )

type H map[string]interface{}

//gee实例对象,使用map存储 路由路径:处理函数 的键值对
type(
	Engine struct{
		router *router
		*RouterGroup
		groups []*RouterGroup //保存所有组

		htmlTemplates *template.Template //渲染html
		funcMap template.FuncMap
	}
	RouterGroup struct{
		prefix string //前缀
		middlewares []HandlerFunc //中间件
		parent *RouterGroup //父亲指针(支持路由嵌套)
		engine *Engine
	}
)

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap){
	engine.funcMap = funcMap
}
func (engine *Engine) LoadHTMLGLOB(pattern string){
	engine.htmlTemplates =
		template.Must( template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
//初始化gee实例对象
func New() *Engine{
	engine := &Engine{ router:newRouter() }
	engine.RouterGroup = &RouterGroup{ engine:engine }
	engine.groups = []*RouterGroup{ engine.RouterGroup }
	return engine
}

//向group中加入中间件函数
func (group *RouterGroup) Use(middlewares ...HandlerFunc){
	group.middlewares = append( group.middlewares,middlewares... )
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
	var middlewares []HandlerFunc
	for _,group := range engine.groups{  //简单的判断:遍历所有组,是这个组的话就把那个组的中间件加进来
		if strings.HasPrefix( req.URL.Path,group.prefix ){
			middlewares = append( middlewares,group.middlewares... )
		}
	}

	c := newContext(w,req)
	c.handlers = middlewares
	c.engine = engine

	engine.router.handle(c)
}

/*
静态文件
 */

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath) //使用/只能连接起两个路径

	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	//其中absolutePath为访问路径,http.FileServer表示返回一个handler,以文件的形式处理请求,且这个文件的前缀为fs

	return func(c *Context) {
		file := c.Param("filepath")  //得到访问的文件名
		// 在fs映射路径下检查是否存在文件file
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		//存在,就开始处理请求
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}
// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath") //relativePath后面的参数作为文件名
	// Register GET handlers
	group.GET(urlPattern, handler)
}