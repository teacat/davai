package davai

import (
	"context"
	"net/http"
)

// contextGet 能夠從一個請求中取得上下文資料。
func contextGet(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

// contextSet 可以將上下文資料保存至指定請求中。
func contextSet(r *http.Request, key, val interface{}) *http.Request {
	if val == nil {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), key, val))
}
