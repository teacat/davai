package davai

import "net/http"

type middlewareFunc func(http.Handler) http.Handler

type middleware interface {
	Middleware(handler http.Handler) http.Handler
}

func (mw middlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}
