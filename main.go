package davai

import (
	"fmt"
	"net/http"
	"sort"
)

func main() {

}

// New 會建立一個新的路由器。
func New() *Router {
	r := &Router{
		routeNames: make(map[string]*Route),
		rules:      make(map[string]*Rule),
	}
	// 初始化一個 `根` 群組。
	r.Group("/")
	// 初始化預設的正規表達式規則。
	r.Rule("*", ".+?")
	r.Rule("i", "[0-9]++")
	r.Rule("s", "[0-9A-Za-z]++")
	return r
}

// Vars 能夠將接收到的路由變數轉換成本地的 `map[string]string` 格式來供存取使用。
func Vars(r *http.Request) map[string]string {
	vars := make(map[string]string)
	return vars
}

// Router 是路由器本體。
type Router struct {
	routeNames      map[string]*Route
	routes          []*Route
	routeGroups     []*RouteGroup
	noRouteHandlers []interface{}
	rules           map[string]*Rule
}

// Get 會依照 GET 方法建立相對應的路由。
func (r *Router) Get(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Get(path, handlers...)
}

// Post 會依照 POST 方法建立相對應的路由。
func (r *Router) Post(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Post(path, handlers...)
}

// Put 會依照 PUT 方法建立相對應的路由。
func (r *Router) Put(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Put(path, handlers...)
}

// Patch 會依照 PATCH 方法建立相對應的路由。
func (r *Router) Patch(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Patch(path, handlers...)
}

// Delete 會依照 DELETE 方法建立相對應的路由。
func (r *Router) Delete(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Delete(path, handlers...)
}

// Options 會依照 OPTIONS 方法建立相對應的路由。
func (r *Router) Options(path string, handlers ...interface{}) *Route {
	return r.routeGroups[0].Options(path, handlers...)
}

// Generate 可以依照傳入的路由名稱與變數來反向產生定義好的路由，這在用於產生模板連結上非常有用。
func (r *Router) Generate(name string, params ...map[string]string) string {
	v, ok := r.routeNames[name]
	if !ok {
		return ""
	}
	var path string
	if len(params) == 0 {
		for _, part := range v.parts {
			path += fmt.Sprintf("/%s", part.path)
		}
		return path
	}
	for _, part := range v.parts {
		if part.name == "" {
			path += fmt.Sprintf("/%s", part.path)
		} else {
			if v, ok := params[0][part.name]; ok {
				path += fmt.Sprintf("/%s", v)
			} else {
				return path
			}
		}
	}
	return path
}

// Rule 能夠在路由器中建立一組新的正規表達式規則供在路由中使用。
func (r *Router) Rule(name string, regexp string) {
	r.rules[name] = &Rule{
		name:   name,
		regExp: regexp,
	}
}

// Group 會建立新的路由群組，群組內的路由會共享前輟與中介軟體。
func (r *Router) Group(path string, middlewares ...func(http.Handler) http.Handler) *RouteGroup {
	group := &RouteGroup{
		router:      r,
		prefix:      path,
		middlewares: middlewares,
	}
	r.routeGroups = append(r.routeGroups, group)
	return group
}

// NoRoute 會將傳入的處理函式作為無相對路由時的執行函式。
func (r *Router) NoRoute(handlers ...interface{}) {
	r.noRouteHandlers = handlers
}

// Run 能夠以 HTTP 開始執行路由器服務。
func (r *Router) Run(addr ...string) error {
	var a string
	if len(addr) == 0 {
		a = ":8080"
	} else {
		a = addr[0]
	}
	return http.ListenAndServe(a, r)
}

// RunTLS 會依據憑證和 HTTPS 的方式開始執行路由器服務。
func (r *Router) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, r)
}

// Shutdown 能夠結束路由器的服務。
func (r *Router) Shutdown() error {
	return nil
}

// ServeHTTP 會處理所有的請求，並分發到指定的路由。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.dispatch(w, req)
}

//
func (r *Router) dispatch(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
}

// sort 會依照路由群組內路由的片段數來做重新排序，用以改進比對時的優先順序。
func (r *Router) sort() {
	sort.Slice(r.routes, func(i, j int) bool {
		return r.routes[i].len > r.routes[j].len
	})
}
