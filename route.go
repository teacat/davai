package davai

import (
	"net/http"
	"strings"
)

// Rule 呈現了單個正規表達式規則。
type Rule struct {
	// Name 是這個規則的代稱。
	name string
	// RegExp 是這個規則的正規表達式。
	regExp string
}

// Part 呈現了路由上的其中一個片段。
type Part struct {
	// Rule 是正規表達式的規則。
	rule *Rule
	// Name 是此路由的擷取名稱。
	name string
	// Path 是這個片段的標準路徑。
	path string
	// IsCaptureGroup 表明此片段是否為擷取群組。
	isCaptureGroup bool
	// IsRegExp 表明此片段是否有用上正規表達式。
	isRegExp bool
	// IsOptional 表明此片段是否為可選。
	isOptional bool
}

// Route 呈現了單個路由資訊。
type Route struct {
	// routeGroup 是這個路由所屬的路由群組。
	routeGroup *RouteGroup
	// name 是路由的名稱，供反向路由使用。
	name string
	// path 是路由的完整路徑。
	path string
	// method 是路由的方法。
	method string
	// parts 是路徑上的片段。
	parts []*Part
	// len 是這個路由所包含的片段數，越高則越有優先度。
	len uint8
	// hasRegExp 表示此路由中是否帶有正規表達式規則。
	hasRegExp bool
	// hasCaptureGroup 表示此路由中是否帶有擷取群組。
	hasCaptureGroup bool
	// handlers 是尚未分類的路由處理函式、中介軟體。
	rawHandlers []interface{}
	// middlewares 是這個路由的中介軟體。
	middlewares []func(http.Handler) http.Handler
	// handler 是這個路由最主要、最終的進入點處理函式。
	handler func(w http.ResponseWriter, r *http.Request)
}

// Name 能夠替此路由命名供稍後以反向路由的方式產生路徑。
func (r *Route) Name(name string) *Route {
	r.name = name
	r.routeGroup.router.routeNames[name] = r
	return r
}

// init 能夠初始化這個路由並且解析路徑成片段供服務開始後比對。
func (r *Route) init() *Route {
	// 歸類處理函式、中介軟體。
	r.sortHandlers()
	// 拆解路由片段。
	r.tearApart()
	return r
}

// sortHandlers 會歸類路由中的處理函式、中介軟體。
func (r *Route) sortHandlers() {
	for _, v := range r.rawHandlers {
		switch t := v.(type) {
		// 中介軟體。
		case func(http.Handler) http.Handler:
			r.middlewares = append(r.middlewares, t)
		// 處理函式。
		case func(w http.ResponseWriter, r *http.Request):
			r.handler = t
		}
	}
}

// tearApart 會將路由的路徑逐一拆解成片段供稍後方便使用。
func (r *Route) tearApart() {
	// 將路徑以 `/` 作為分水嶺來拆開。
	parts := strings.Split(r.path, "/")
	// 記載這個路由的片段數量，之後會以此作為優先順序考量。
	r.len = uint8(len(parts))
	// 遞迴每個片段，並且分析資料。
	for _, v := range parts {
		// 是否為 `{}` 擷取群組。
		var isCaptureGroup bool
		if v[0:1] == "{" {
			isCaptureGroup = true
			// 移除路徑上的擷取群組符號。
			v = strings.TrimRight(strings.TrimLeft(v, "{"), "}")
		}
		// 是否有 `?` 作為可選路由。
		var isOptional bool
		if v[len(v)-1:] == "?" {
			isOptional = true
			//移除路徑上的可選符號。
			v = strings.TrimRight(v, "?")
		}
		// 是否有正規表達式規則。
		var isRegExp bool
		if strings.Contains(v, ":") {
			isRegExp = true
		}
		// 取得擷取群組和規則名稱。
		var varName string
		var ruleName string
		if isCaptureGroup {
			if isRegExp {
				details := strings.Split(v, ":")
				varName = details[1]
				ruleName = details[0]
			} else {
				varName = v
			}
		}
		// 取得相對應的規則建構體。
		var rule *Rule
		if ruleName != "" {
			rule = r.routeGroup.router.rules[ruleName]
		}
		//
		r.parts = append(r.parts, &Part{
			rule:           rule,
			name:           varName,
			path:           v,
			isCaptureGroup: isCaptureGroup,
			isRegExp:       isRegExp,
			isOptional:     isOptional,
		})
		//
		if isCaptureGroup {
			r.hasCaptureGroup = true
		}
	}
}
