package davai

import "net/http"

// middlewareFunc 是一個中介軟體最基本的雛形型態。
type middlewareFunc func(http.Handler) http.Handler

// middleware 是 Davai 裡的標準中介軟體介面。
type middleware interface {
	Middleware(next http.Handler) http.Handler
}

// Middleware 會回傳一個能接收下一個中介函式並透過 `ServeHTTP` 來繼續整個路由函式鏈的函式。
func (mw middlewareFunc) Middleware(next http.Handler) http.Handler {
	return mw(next)
}
