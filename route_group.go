package davai

import (
	"net/http"
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
	middlewares []func(http.Handler) http.Handler
}

// newRoute 會在目前的路由群組中依指定的方法、路徑、處理函式來插入新的路由。
func (r *RouteGroup) newRoute(method string, path string, handlers ...interface{}) *Route {
	//
	var rawHandlers []interface{}
	for _, v := range r.middlewares {
		rawHandlers = append(rawHandlers, v)
	}
	for _, v := range handlers {
		rawHandlers = append(rawHandlers, v)
	}
	route := &Route{
		routeGroup:  r,
		path:        r.prefix + path,
		rawHandlers: rawHandlers,
		method:      method,
	}
	// 初始化路由。
	route.init()
	// 保存路由至此群組。
	r.routes = append(r.routes, route)
	//
	if route.isStatic {
		//
		r.router.staticRoutes[route.path] = route
	} else {
		// 保存路由至路由器。
		r.router.routes = append(r.router.routes, route)
		// 依照優先度重新排序路由。
		r.router.sort()
	}
	return route
}

// Get 會依照 GET 方法建立相對應的路由。
func (r *RouteGroup) Get(path string, handlers ...interface{}) *Route {
	return r.newRoute("get", path, handlers...)
}

// Post 會依照 POST 方法建立相對應的路由。
func (r *RouteGroup) Post(path string, handlers ...interface{}) *Route {
	return r.newRoute("post", path, handlers...)
}

// Put 會依照 PUT 方法建立相對應的路由。
func (r *RouteGroup) Put(path string, handlers ...interface{}) *Route {
	return r.newRoute("put", path, handlers...)
}

// Patch 會依照 PATCH 方法建立相對應的路由。
func (r *RouteGroup) Patch(path string, handlers ...interface{}) *Route {
	return r.newRoute("patch", path, handlers...)
}

// Delete 會依照 DELETE 方法建立相對應的路由。
func (r *RouteGroup) Delete(path string, handlers ...interface{}) *Route {
	return r.newRoute("delete", path, handlers...)
}

// Options 會依照 OPTIONS 方法建立相對應的路由。
func (r *RouteGroup) Options(path string, handlers ...interface{}) *Route {
	return r.newRoute("options", path, handlers...)
}
