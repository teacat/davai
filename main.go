package davai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrRouteNotFound     = errors.New("davai: the route was not found")
	ErrHandlerNotFound   = errors.New("davai: the handler of the route was not found, it might be a nil pointer")
	ErrVarNotFound       = errors.New("davai: cannot generate the route if the required parameter has no matched variable")
	ErrFileNotFound      = errors.New("davai: the file to serve was not found")
	ErrDirectoryNotFound = errors.New("davai: the directory to serve was not found")
)

func main() {

}

// New 會建立一個新的路由器。
func New() *Router {
	r := &Router{
		routeNames: make(map[string]*Route),
		rules:      make(map[string]*Rule),
		routes: map[string]*routes{
			"GET": {
				method:  "GET",
				statics: make(map[string]*Route),
			},
			"POST": {
				method:  "POST",
				statics: make(map[string]*Route),
			},
			"PUT": {
				method:  "PUT",
				statics: make(map[string]*Route),
			},
			"PATCH": {
				method:  "PATCH",
				statics: make(map[string]*Route),
			},
			"DELETE": {
				method:  "DELETE",
				statics: make(map[string]*Route),
			},
			"OPTIONS": {
				method:  "OPTIONS",
				statics: make(map[string]*Route),
			},
		},
	}
	// 初始化一個 `根` 群組。
	r.Group("")
	// 初始化預設的正規表達式規則。
	r.Rule("*", ".*")
	r.Rule("i", "[0-9]+")
	r.Rule("s", "[0-9A-Za-z]+")
	//
	r.NoRoute(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	return r
}

const (
	varsKey = "davaiVars"
)

// 重新檢查 empty 的 vars 該不該納入 map
//
//
//
//
//

// Vars 能夠將接收到的路由變數轉換成本地的 `map[string]string` 格式來供存取使用。
func Vars(r *http.Request) map[string]string {
	if rv := contextGet(r, varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

// routes 是單個方法的所有路由。
type routes struct {
	// method 是這個方法的名稱。
	method string
	// statics 是所有的靜態路由，這會讓路由比對率先和此切片快速比對，
	// 若無相符的路由才重新和所有動態路由比對。
	statics map[string]*Route
	// dymanics 是所有的動態路由。
	dymanics []*Route
}

// Router 是路由器本體。
type Router struct {
	// server 是 HTTP 伺服器。
	server *http.Server
	// routeNames 是用來存放已命名的路由供之後取得。
	routeNames map[string]*Route
	// routes 是現有的全部路由，以不同方法作為鍵名區分。
	routes map[string]*routes
	// routeGroups 是所有的路由群組。
	routeGroups []*RouteGroup
	// middlewares 是全域中介軟體。
	middlewares []middleware
	// noRouteMiddlewares 是無路由的中介軟體。
	noRouteMiddlewares []middleware
	// noRouteHandler 是無路由時所會呼叫的處理函式。
	noRouteHandler func(w http.ResponseWriter, r *http.Request)
	// rules 用來存放所有的正規表達式規則。
	rules map[string]*Rule
}

func (r *Router) ServeFile(path string, handlers ...interface{}) *Route {
	var serveHandler interface{}

	for _, v := range handlers {
		switch a := v.(type) {
		case string:
			if _, err := os.Stat(a); err != nil {
				panic(ErrFileNotFound)
			}
			serveHandler = func(w http.ResponseWriter, req *http.Request) {
				http.ServeFile(w, req, a)
			}
		}
	}

	if serveHandler != nil {
		handlers[len(handlers)-1] = serveHandler
	}

	return r.routeGroups[0].Get(path, handlers...)
}

//
func (r *Router) ServeFiles(path string, handlers ...interface{}) *Route {
	strSlice := strings.Split(path, "/{*:")
	prefix := strSlice[0]

	var serveHandler interface{}

	for k, v := range handlers {
		switch a := v.(type) {

		case http.Dir:
			handlers[k] = http.StripPrefix(prefix, http.FileServer(a))

		case string:
			if _, err := os.Stat(a); err != nil {
				panic(ErrDirectoryNotFound)
			}

			if info, err := os.Stat(a); err == nil && info.IsDir() {
				serveHandler = http.StripPrefix(prefix, http.FileServer(http.Dir(a)))
			} else {
				serveHandler = func(w http.ResponseWriter, req *http.Request) {
					http.ServeFile(w, req, a)
				}
			}
		}
	}

	if serveHandler != nil {
		handlers[len(handlers)-1] = serveHandler
	}

	return r.routeGroups[0].Get(path, handlers...)
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
	//
	var path string
	if len(params) == 0 {
		for _, part := range v.parts {
			if part.name != "" {
				return path
			}
			path += fmt.Sprintf("/%s", part.path)
		}
		return path
	}
	//
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
func (r *Router) Rule(name string, expr string) {
	expr = fmt.Sprintf("^%s$", expr)
	r.rules[name] = &Rule{
		name:   name,
		expr:   expr,
		regexp: regexp.MustCompile(expr),
	}
}

// Group 會建立新的路由群組，群組內的路由會共享前輟與中介軟體。
func (r *Router) Group(path string, middlewares ...interface{}) *RouteGroup {
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
	for _, v := range handlers {
		switch t := v.(type) {
		// 中介軟體。
		case func(http.Handler) http.Handler:
			r.noRouteMiddlewares = append(r.noRouteMiddlewares, MiddlewareFunc(t))
		// 進階中介軟體。
		case middleware:
			r.noRouteMiddlewares = append(r.noRouteMiddlewares, t)
		// 處理函式。
		case func(w http.ResponseWriter, r *http.Request):
			r.noRouteHandler = t
		}
	}
}

// Run 能夠以 HTTP 開始執行路由器服務。
func (r *Router) Run(addr ...string) error {
	var a string
	if len(addr) == 0 {
		a = ":8080"
	} else {
		a = addr[0]
	}
	r.server = &http.Server{
		Addr: "0.0.0.0" + a,
		// WriteTimeout: time.Second * 15,
		// ReadTimeout:  time.Second * 15,
		// IdleTimeout:  time.Second * 60,
		Handler: r,
	}
	return r.server.ListenAndServe()
}

// RunTLS 會依據憑證和 HTTPS 的方式開始執行路由器服務。
func (r *Router) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, r)
}

// Shutdown 會關閉伺服器。
func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// ServeHTTP 會處理所有的請求，並分發到指定的路由。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.dispatch(w, req)
}

//
//
//
//
//
//
///
func (r *Router) Use(middlewares ...interface{}) {

}

// Strict 能夠更改路由器的嚴格設定，當設置為 `true` 的時候會嚴格比對路由的結尾斜線，預設為 `false`。
// func (r *Router) Strict(value bool) *Router {
//
// }
//
// CaseSensitive 會更改路由器的大小寫敏感設定，當設置為 `true` 的時候 `/foo` 不會和 `/FOO` 相符，預設為 `false`。
// func (r *Router) CaseSensitive(value bool) *Router {
//
// }
//
// RegexCache 能啟用路由器的正規表達式快取，如果路由中有正規表達式規則且內容通常是固定的，那麼開啟此功能可以增進效能。
// func (r *Router) RegexCache(value bool) *Router {
//
// }
//
// SkipClean 會略過網址清理，當設置為 `true` 的時候並不會移除網址多餘的雙斜線，這讓你能用上類似 `/fetch/http://www.example.com/` 的路由。
// func (r *Router) SkipClean(value bool) *Router {
//
// }

// call 會呼叫指定路由的處理函式。
func (r *Router) call(route *Route, w http.ResponseWriter, req *http.Request) {
	if route.handler == nil {
		panic(ErrHandlerNotFound)
	}

	middlewareLength := len(route.middlewares)

	var handler http.Handler
	handler = route.handler

	for i := middlewareLength - 1; i >= 0; i-- {
		handler = route.middlewares[i].Middleware(handler)
	}

	handler.ServeHTTP(w, req)
}

// callNoRoute 會串連中介軟體並且呼叫無路由的函式。
func (r *Router) callNoRoute(w http.ResponseWriter, req *http.Request) {
	middlewareLength := len(r.noRouteMiddlewares)

	var handler http.Handler
	handler = http.HandlerFunc(r.noRouteHandler)

	for i := middlewareLength - 1; i >= 0; i-- {
		handler = r.noRouteMiddlewares[i].Middleware(handler)
	}

	handler.ServeHTTP(w, req)
}

// 移除尾部 「/」
// 移除尾部 「/」
// 移除尾部 「/」
// 移除尾部 「/」
// 移除尾部 「/」
func (r *Router) match(routes *routes, w http.ResponseWriter, req *http.Request) bool {
	var url string
	url = req.URL.Path
	if req.URL.Path != "/" {
		fmt.Println(req.URL.Path)
		url = strings.ToLower(strings.TrimRight(req.URL.Path, "/"))
		fmt.Println(url)
	}

	if route, ok := routes.statics[url]; ok {
		r.call(route, w, req)
		return true
	}

	components := strings.Split(url, "/")[1:]
	fmt.Println(components)
	componentLength := len(components)

	for _, route := range routes.dymanics {
		var matched bool
		var vars map[string]string
		partLength := len(route.parts)

	partScan:
		for index, part := range route.parts {
			component := components[index]
			isLastComponent := index+1 > componentLength-1
			isLastPart := index+1 > partLength-1

			switch {
			case part.isStatic:
				if part.path != component {
					break partScan
				}
			case part.isCaptureGroup:
				if part.prefix != "" {
					if !strings.HasPrefix(component, part.prefix) {
						break partScan
					}
					component = strings.TrimPrefix(component, part.prefix)
				}
				if part.suffix != "" {
					if !strings.HasSuffix(component, part.suffix) {
						break partScan
					}
					component = strings.TrimSuffix(component, part.suffix)
				}
				if part.prefix != "" || part.suffix != "" {
					if !part.isOptional && !part.isRegExp && component == "" {
						break partScan
					}
				}
				if part.isRegExp {
					if part.rule.name == "*" {
						if isLastPart {
							if vars == nil {
								vars = make(map[string]string)
							}
							vars[part.name] = strings.Join(components[index:], "/")

							matched = true
							break partScan
						}
					}
					if (part.isOptional && component != "") || !part.isOptional {
						if !part.rule.regexp.MatchString(component) {
							break partScan
						}
					}
				}
				if vars == nil {
					vars = make(map[string]string)
				}
				vars[part.name] = component
			}
			if !isLastPart {
				if component != "" {
					if route.parts[index+1].isOptional {
						if isLastComponent {
							matched = true
							break
						}
					}
					if route.parts[index+1].rule.name == "*" {
						if index+1 == partLength-1 {
							if vars == nil {
								vars = make(map[string]string)
							}
							vars[part.name] = strings.Join(components[index:], "/")

							matched = true
							break partScan
						}

					}
				}
			}
			if isLastPart && isLastComponent {
				matched = true
				break
			}
			if isLastComponent {
				break
			}
		}
		if matched {
			if vars == nil {
				r.call(route, w, req)
			} else {
				r.call(route, w, contextSet(req, varsKey, vars))
			}
			return true
		}
	}
	return false
}

// disaptch 會解析接收到的請求並依照網址分發給指定的路由。
func (r *Router) dispatch(w http.ResponseWriter, req *http.Request) {
	var matched bool
	switch req.Method {
	case "GET":
		matched = r.match(r.routes["GET"], w, req)
	case "POST":
		matched = r.match(r.routes["POST"], w, req)
	case "PUT":
		matched = r.match(r.routes["PUT"], w, req)
	case "PATCH":
		matched = r.match(r.routes["PATCH"], w, req)
	case "DELETE":
		matched = r.match(r.routes["DELETE"], w, req)
	case "OPTIONS":
		matched = r.match(r.routes["OPTIONS"], w, req)
	}
	if !matched {
		r.callNoRoute(w, req)
	}
}

// sort 會依照路由群組內路由的片段數來做重新排序，用以改進比對時的優先順序。
func (r *Router) sort(method string) {
	switch method {
	case "GET":
		sort.Slice(r.routes["GET"].dymanics, func(i, j int) bool {
			return r.routes["GET"].dymanics[i].priority > r.routes["GET"].dymanics[j].priority
		})
	case "POST":
		sort.Slice(r.routes["POST"].dymanics, func(i, j int) bool {
			return r.routes["POST"].dymanics[i].priority > r.routes["POST"].dymanics[j].priority
		})
	case "PUT":
		sort.Slice(r.routes["PUT"].dymanics, func(i, j int) bool {
			return r.routes["PUT"].dymanics[i].priority > r.routes["PUT"].dymanics[j].priority
		})
	case "PATCH":
		sort.Slice(r.routes["PATCH"].dymanics, func(i, j int) bool {
			return r.routes["PATCH"].dymanics[i].priority > r.routes["PATCH"].dymanics[j].priority
		})
	case "DELETE":
		sort.Slice(r.routes["DELETE"].dymanics, func(i, j int) bool {
			return r.routes["DELETE"].dymanics[i].priority > r.routes["DELETE"].dymanics[j].priority
		})
	case "OPTIONS":
		sort.Slice(r.routes["OPTIONS"].dymanics, func(i, j int) bool {
			return r.routes["OPTIONS"].dymanics[i].priority > r.routes["OPTIONS"].dymanics[j].priority
		})
	}

}
