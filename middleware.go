package davai

import "net/http"

type MiddlewareFunc func(http.Handler) http.Handler

type middleware interface {
	Middleware(handler http.Handler) http.Handler
}

func (mw MiddlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}
