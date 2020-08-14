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
	// ErrRouteNotFound 表示產生反向路由的時候找不到對象路由。
	ErrRouteNotFound = errors.New("davai: the route was not found")
	// ErrHandlerNotFound 表示無法找到路由的最終處理函式，也許該函式是個 `nil` 指標。
	ErrHandlerNotFound = errors.New("davai: the handler of the route was not found, it might be a nil pointer")
	// ErrVarNotFound 表示產生反向路由時，必要的網址變數並不存在而無法反向產生該路由。
	ErrVarNotFound = errors.New("davai: cannot generate the route if the required parameter has no matched variable")
	// ErrFileNotFound 表示欲提供的靜態檔案並不存在。
	ErrFileNotFound = errors.New("davai: the file to serve was not found")
	// ErrDirectoryNotFound 表示欲提供的靜態目錄資料夾並不存在。
	ErrDirectoryNotFound = errors.New("davai: the directory to serve was not found")
)

const (
	varsKey  = "davaiVars"
	routeKey = "davaiRoute"
)

// Vars 能夠將接收到的路由變數轉換成本地的 `map[string]string` 格式來供存取使用。
// 如果路由中有選擇性路由，且請求網址中省略了該變數，取得到的變數結果則會是空字串而非 `nil` 值。
func Vars(r *http.Request) map[string]string {
	// if route := contextGet(r, routeKey); route != nil {
	// 	if v := contextGet(r, varsKey); v != nil {
	// 		vars := v.(map[string]string)
	// 		for k := range route.(*Route).defaultCaptureVars {
	// 			if _, ok := vars[k]; !ok {
	// 				vars[k] = ""
	// 			}
	// 		}
	// 		return vars
	// 	}
	// }
	if rv := contextGet(r, varsKey); rv != nil {
		return rv.(map[string]string)
	}
	return nil
}

// New 會建立一個新的路由器。
func New() *Router {
	r := &Router{
		routeNames: make(map[string]*Route),
		rules:      make(map[string]*rule),
		methodRoutes: map[string]*routes{
			"GET": {
				method:  "GET",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
			"POST": {
				method:  "POST",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
			"PUT": {
				method:  "PUT",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
			"PATCH": {
				method:  "PATCH",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
			"DELETE": {
				method:  "DELETE",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
			"OPTIONS": {
				method:  "OPTIONS",
				statics: make(map[string]*Route),
				caches:  make(map[string]*cacheRoute),
			},
		},
	}
	r.Group("")
	r.Rule("*", ".*")
	r.Rule("i", "[0-9]+")
	r.Rule("s", "[0-9A-Za-z]+")
	r.NoRoute(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found\n"))
	})
	return r
}

// routes 是單個方法的所有路由。
type routes struct {
	// method 是這個方法的名稱。
	method string
	// statics 是所有的靜態路由，這會讓路由比對率先和此切片快速比對，
	// 若無相符的路由才重新和所有動態路由比對。
	statics map[string]*Route
	// dynamics 是所有的動態路由。
	dynamics []*Route
	//
	caches map[string]*cacheRoute
}

type cacheRoute struct {
	route *Route
	vars  map[string]string
}

// Router 是路由器本體。
type Router struct {
	// server 是 HTTP 伺服器。
	server *http.Server
	// routeNames 是用來存放已命名的路由供之後取得。
	routeNames map[string]*Route
	// routes 是現有的全部路由。
	routes []*Route
	// methodRoutes 是以不同方法作為鍵名區分的所有路由。
	methodRoutes map[string]*routes
	// routeGroups 是所有的路由群組。
	routeGroups []*RouteGroup
	// middlewares 是全域中介軟體。
	middlewares []middleware
	// noRouteMiddlewares 是無路由的中介軟體。
	noRouteMiddlewares []middleware
	// noRouteHandler 是無路由時所會呼叫的處理函式。
	noRouteHandler func(w http.ResponseWriter, r *http.Request)
	// rules 用來存放所有的正規表達式規則。
	rules map[string]*rule
}

// ServeFile 能夠提供某個靜態檔案，其中可以安插中介軟體，而最後一個參數必須是字串來表示檔案的相對位置。
// 當路由器起動時，檔案不存在於硬碟上則會發生 `ErrFileNotFound` 錯誤。
func (r *Router) ServeFile(path string, handlers ...interface{}) *Route {
	for k, v := range handlers {
		switch t := v.(type) {
		case string:
			if _, err := os.Stat(t); err != nil {
				panic(ErrFileNotFound)
			}
			handlers[k] = func(w http.ResponseWriter, req *http.Request) {
				http.ServeFile(w, req, t)
			}
		}
	}
	return r.routeGroups[0].Get(path, handlers...)
}

// ServeFiles 可以提供整個靜態資料夾目錄，其中可以安插中介軟體，而最後一個參數必須是字串來表示資料夾的相對位置。
// 當路由器起動時，資料夾不存在於硬碟上則會發生 `ErrDirectoryNotFound` 錯誤。
func (r *Router) ServeFiles(path string, handlers ...interface{}) *Route {
	// 定義一個前輟路由，這樣才能接收檔案的名稱來瀏覽資料夾。
	route := r.routeGroups[0].Get(path+"/{*:file}", handlers...)
	// 遍歷傳入的處理函式，並且依照不同型態來做處置。
	for k, v := range route.rawHandlers {
		switch t := v.(type) {
		case http.Dir:
			route.rawHandlers[k] = http.StripPrefix(path, http.FileServer(t))
		case string:
			if info, err := os.Stat(t); err != nil || !info.IsDir() {
				panic(ErrDirectoryNotFound)
			}
			// 將最終的函式改為 Davai 自己的檔案處理函式，而非 `net/http` 的 `FileServer`，
			// 因為 `FileServer` 會依照尾斜線來作為是否是資料夾的判定，而 Davai 試圖寬容這個問題所以需要自幹檔案處理函式，可能會稍慢就是了。
			route.rawHandlers[k] = http.StripPrefix(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				var isDirStatus bool
				// checkIsDir 會檢查傳入的請求是否為資料夾型態。獨立成一個函式是為了讓他在最終才執行，
				// 這樣也許可以提升效能來避免每次有請求就先執行資料夾檢查。
				isDir := func() bool {
					if isDirStatus {
						return true
					}
					if info, err := os.Stat(t + req.URL.Path); err == nil && info.IsDir() {
						isDirStatus = true
						return true
					}
					return false
				}
				// 如果結尾沒有斜線，而且路徑又不是根目錄，且該目的地是個資料夾的話，
				// 那麼就透過 301 重新導向到有尾斜線的路由。
				if !strings.HasSuffix(req.URL.Path, "/") && req.URL.Path != "" {
					if isDir() {
						http.Redirect(w, req, path+req.URL.Path+"/", http.StatusPermanentRedirect)
						return
					}
				}
				// 如果是資料夾，同時這個路由又不允許資料夾索引的話則回傳 HTTP 403 錯誤。
				if isDir() && !route.DirectoryListing {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				// 其他檔案的讀取交給 FileServer 來處理。
				http.FileServer(http.Dir(t)).ServeHTTP(w, req)
			}))
		}
	}
	return route
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
// 當路由中有必要的變數但卻無傳入時會發生 `ErrVarNotFound` 錯誤，如果沒有指定的命名路由則會是 `ErrRouteNotFound` 錯誤。
func (r *Router) Generate(name string, params ...map[string]string) string {
	v, ok := r.routeNames[name]
	if !ok {
		panic(ErrRouteNotFound)
	}
	// 如果沒有傳入變數的話。
	var path string
	if len(params) == 0 {
		for _, part := range v.parts {
			if part.name != "" {
				panic(ErrVarNotFound)
			}
			path += fmt.Sprintf("/%s", part.path)
		}
		return path
	}
	// 如果有傳入變數的話就依照路由片段的要求找出對應的變數值。
	for _, part := range v.parts {
		if part.name == "" {
			path += fmt.Sprintf("/%s", part.path)
		} else {
			if v, ok := params[0][part.name]; ok {
				path += fmt.Sprintf("/%s", v)
			} else {
				panic(ErrVarNotFound)
			}
		}
	}
	return path
}

// Rule 能夠在路由器中建立一組新的正規表達式規則供在路由中使用。
func (r *Router) Rule(name string, expr string) {
	expr = fmt.Sprintf("^%s$", expr)
	r.rules[name] = &rule{
		name:   name,
		expr:   expr,
		regexp: regexp.MustCompile(expr),
	}
}

// Group 會建立新的路由群組，群組內的路由會共享前輟與中介軟體。
func (r *Router) Group(path string, middlewares ...interface{}) *RouteGroup {
	group := &RouteGroup{
		router: r,
		prefix: path,
	}
	group.Use(middlewares...)
	r.routeGroups = append(r.routeGroups, group)
	return group
}

// NoRoute 會將傳入的處理函式作為無相對路由時的執行函式。
func (r *Router) NoRoute(handlers ...interface{}) {
	for _, v := range handlers {
		switch t := v.(type) {
		// 中介軟體。
		case func(http.Handler) http.Handler:
			r.noRouteMiddlewares = append(r.noRouteMiddlewares, middlewareFunc(t))
		// 進階中介軟體。
		case middleware:
			r.noRouteMiddlewares = append(r.noRouteMiddlewares, t)
		// 處理函式。
		case func(w http.ResponseWriter, r *http.Request):
			r.noRouteHandler = t
		}
	}
}

// sortMiddlewares 會重新整理路由中的所有中介軟體並將其安插到每個路由的執行函式鏈中。
func (r *Router) sortMiddlewares() {
	for _, route := range r.routes {
		// 先套用全域中介軟體。
		route.middlewares = r.middlewares
		// 接著套用路由群組的中介軟體。
		route.middlewares = append(route.middlewares, route.routeGroup.middlewares...)
		// 然後才是本路由的中介軟體與處理函式。
		for _, v := range route.rawHandlers {
			switch t := v.(type) {
			// 中介軟體。
			case func(http.Handler) http.Handler:
				route.middlewares = append(route.middlewares, middlewareFunc(t))
			// 進階中介軟體。
			case middleware:
				route.middlewares = append(route.middlewares, t)
			// 處理函式。
			case func(http.ResponseWriter, *http.Request):
				route.handler = http.HandlerFunc(t)
			case http.Handler:
				route.handler = t
			}
		}
	}
}

// Run 能夠以 HTTP 開始執行路由器服務，若無指定的埠口則會採用預設的 `:8080` 位置。
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
	r.sortMiddlewares()
	r.sortRoutes()
	return r.server.ListenAndServe()
}

// sortRoutes 會在啟動之前重新整理路由的優先度，因為路由的優先度可能會在之前透過人工干預。
func (r *Router) sortRoutes() {
	for _, v := range r.routes {
		// 依照優先度重新排序動態路由。
		v.routeGroup.router.sort(v.method)
	}
}

// RunTLS 會依據憑證和 HTTPS 的方式開始執行路由器服務。
func (r *Router) RunTLS(addr string, certFile string, keyFile string) error {
	r.server = &http.Server{
		Addr: "0.0.0.0" + addr,
		// WriteTimeout: time.Second * 15,
		// ReadTimeout:  time.Second * 15,
		// IdleTimeout:  time.Second * 60,
		Handler: r,
	}
	r.sortMiddlewares()
	r.sortRoutes()
	return r.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown 會完好地關閉伺服器。
func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// ServeHTTP 會處理所有的請求，並分發到指定的路由。
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.dispatch(w, req)
}

// Use 能將傳入的中介軟體作為全域中介軟體在所有路由中使用。
func (r *Router) Use(middlewares ...interface{}) *Router {
	for _, v := range middlewares {
		switch t := v.(type) {
		// 中介軟體。
		case func(http.Handler) http.Handler:
			r.middlewares = append(r.middlewares, middlewareFunc(t))
		// 進階中介軟體。
		case middleware:
			r.middlewares = append(r.middlewares, t)
		}
	}
	return r
}

// call 會呼叫指定路由的處理函式，當沒有指定的處理函式時會發生 `ErrHandlerNotFound` 錯誤。
func (r *Router) call(route *Route, w http.ResponseWriter, req *http.Request) {
	if route.handler == nil {
		panic(ErrHandlerNotFound)
	}
	var handler http.Handler
	handler = route.handler

	middlewareLength := len(route.middlewares)
	for i := middlewareLength - 1; i >= 0; i-- {
		handler = route.middlewares[i].Middleware(handler)
	}
	handler.ServeHTTP(w, req)
}

// callNoRoute 會串連中介軟體並且呼叫無路由的函式。
func (r *Router) callNoRoute(w http.ResponseWriter, req *http.Request) {
	var handler http.Handler
	handler = http.HandlerFunc(r.noRouteHandler)

	middlewareLength := len(r.noRouteMiddlewares)
	for i := middlewareLength - 1; i >= 0; i-- {
		handler = r.noRouteMiddlewares[i].Middleware(handler)
	}
	handler.ServeHTTP(w, req)
}

// match 會逐一檢查路由並比對是否和請求網址相符。
func (r *Router) match(routes *routes, w http.ResponseWriter, req *http.Request) bool {
	url := req.URL.Path
	if req.URL.Path != "/" {
		url = strings.ToLower(strings.TrimRight(req.URL.Path, "/"))
	}
	if route, ok := routes.statics[url]; ok {
		r.call(route, w, req)
		return true
	}
	//if route, ok := routes.caches[url]; ok {
	//	r.call(route.route, w, contextSet(req, varsKey, route.vars))
	//	return true
	//}

	components := strings.Split(url, "/")[1:]
	componentLength := len(components)
	if componentLength == 0 {
		return false
	}

	for _, route := range routes.dynamics {
		var matched bool
		vars := make(map[string]string)
		for k, v := range route.defaultCaptureVars {
			vars[k] = v
		}
		partLength := len(route.parts)

	partScan:
		for index, part := range route.parts {
			component := components[index]
			isLastComponent := index == componentLength-1
			isLastPart := index == partLength-1

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
				vars[part.name] = component
			}
			if !isLastPart {
				if component != "" {
					nextPart := route.parts[index+1]
					if nextPart.isOptional {
						if isLastComponent {
							matched = true
							break
						}
					}
					if nextPart.isRegExp {
						if nextPart.rule.name == "*" {
							vars[nextPart.name] = strings.Join(components[index+1:], "/")
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
			//routes.caches[url] = &cacheRoute{
			//	route: route,
			//	vars:  vars,
			//}
			r.call(route, w, contextSet(req, varsKey, vars))
			return true
		}
	}
	return false
}

// disaptch 會解析接收到的請求並依照網址分發給指定的路由。
func (r *Router) dispatch(w http.ResponseWriter, req *http.Request) {
	var matched bool
	if v, ok := r.methodRoutes[req.Method]; ok {
		matched = r.match(v, w, req)
	}
	if !matched {
		r.callNoRoute(w, req)
	}
}

// sort 會依照路由群組內路由的片段數來做重新排序，用以改進比對時的優先順序。
func (r *Router) sort(method string) {
	sort.Slice(r.methodRoutes[method].dynamics, func(i, j int) bool {
		return r.methodRoutes[method].dynamics[i].priority > r.methodRoutes[method].dynamics[j].priority
	})
}
