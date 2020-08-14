package davai

import (
	"net/http"
	"strings"
)

// RouteGroup 是單個路由群組。
type RouteGroup struct {
	// router 是這個路由群組所屬的路由器。
	router *Router
	// prefix 是這個路由群組的前輟路徑。
	prefix string
	// routes 表示這個群組內的路由。
	routes []*Route
	// middlewares 是這個路由群組的共享中介軟體。
	middlewares []middleware
}

// newRoute 會在目前的路由群組中依指定的方法、路徑、處理函式來插入新的路由。
func (r *RouteGroup) newRoute(method string, path string, handlers ...interface{}) *Route {
	if path == "/" {
		if r.prefix != "" {
			path = ""
		}
	}
	if path != "/" {
		path = strings.TrimRight(r.prefix+path, "/")
	}
	route := &Route{
		routeGroup:  r,
		path:        path,
		rawHandlers: handlers,
		method:      method,
	}
	// 初始化路由。
	route.init()
	// 保存路由至此群組。
	r.routes = append(r.routes, route)
	// 保存路由至此路由器。
	r.router.routes = append(r.router.routes, route)
	// 將路由依照動態和靜態保存到不同的路由樹中。
	if route.isStatic {
		r.router.methodRoutes[route.method].statics[route.path] = route
	} else {
		r.router.methodRoutes[route.method].dynamics = append(r.router.methodRoutes[route.method].dynamics, route)
	}
	return route
}

// Use 能將傳入的中介軟體作為群組中介軟體在群組內的所有路由中使用。
func (r *RouteGroup) Use(middlewares ...interface{}) *RouteGroup {
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

// Get 會依照 GET 方法建立相對應的路由。
func (r *RouteGroup) Get(path string, handlers ...interface{}) *Route {
	return r.newRoute("GET", path, handlers...)
}

// Post 會依照 POST 方法建立相對應的路由。
func (r *RouteGroup) Post(path string, handlers ...interface{}) *Route {
	return r.newRoute("POST", path, handlers...)
}

// Put 會依照 PUT 方法建立相對應的路由。
func (r *RouteGroup) Put(path string, handlers ...interface{}) *Route {
	return r.newRoute("PUT", path, handlers...)
}

// Patch 會依照 PATCH 方法建立相對應的路由。
func (r *RouteGroup) Patch(path string, handlers ...interface{}) *Route {
	return r.newRoute("PATCH", path, handlers...)
}

// Delete 會依照 DELETE 方法建立相對應的路由。
func (r *RouteGroup) Delete(path string, handlers ...interface{}) *Route {
	return r.newRoute("DELETE", path, handlers...)
}

// Options 會依照 OPTIONS 方法建立相對應的路由。
func (r *RouteGroup) Options(path string, handlers ...interface{}) *Route {
	return r.newRoute("OPTIONS", path, handlers...)
}
