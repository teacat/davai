package davai

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	priorityRoot      = 20
	priorityPath      = 16
	priorityStatic    = 8
	priorityGroup     = 4
	priorityText      = 2
	priorityRegExp    = 1
	priorityOptional  = -1
	priorityAnyRegExp = -2
)

// Rule 呈現了單個正規表達式規則。
type rule struct {
	// Name 是這個規則的代稱。
	name string
	// expr 是這個規則的表達式內容。
	expr string
	// regexp 是編譯後的正規表達式。
	regexp *regexp.Regexp
}

// Part 呈現了路由上的其中一個片段。
type part struct {
	// rule 是正規表達式的規則。
	rule *rule
	// name 是此路由的擷取名稱。
	name string
	// path 是這個片段的標準路徑。
	path string
	// suffix 是片段的固定後輟。
	suffix string
	// prefix 是片段的固定前輟。
	prefix string
	// isStatic 表明是否為靜態片段無任何屬性。
	isStatic bool
	// isCaptureGroup 表明此片段是否為擷取群組。
	isCaptureGroup bool
	// isRegExp 表明此片段是否有用上正規表達式。
	isRegExp bool
	// isOptional 表明此片段是否為可選。
	isOptional bool
}

// Route 呈現了單個路由資訊。
type Route struct {
	// RegexCache 能啟用路由器的正規表達式快取，如果路由中有正規表達式規則且內容通常是固定的，那麼開啟此功能可以增進效能。
	RegexCache bool
	// DirectoryListing 可以決定此路由所提供的靜態目錄是否允許暴露底下的檔案。
	DirectoryListing bool

	// routeGroup 是這個路由所屬的路由群組。
	routeGroup *RouteGroup
	// name 是路由的名稱，供反向路由使用。
	name string
	// path 是路由的完整路徑。
	path string
	// method 是路由的方法。
	method string
	// parts 是路徑上的片段。
	parts []*part
	// len 是這個路由的片段數量。
	len int
	// priority 是這個路由的優先度。
	priority int16
	// isStatic 表明是否為靜態路由無任何屬性。
	isStatic bool
	// hasRegExp 表示此路由中是否帶有正規表達式規則。
	hasRegExp bool
	// hasCaptureGroup 表示此路由中是否帶有擷取群組。
	hasCaptureGroup bool
	// defaultCaptureVars 是預設的擷取變數結果，這是個唯獨資料必須複製來更改。
	// 這個的存在是用來讓 `Vars` 能夠在沒有接收到指定變數的情況下回傳一個完整的截取變數結果，
	// 而不需要每次都完整掃描請求路由來取得空結果。
	defaultCaptureVars map[string]string
	// rawHandlers 是尚未分類的路由處理函式、中介軟體。
	rawHandlers []interface{}
	// middlewares 是這個路由的中介軟體。
	middlewares []middleware
	// handler 是這個路由最主要、最終的進入點處理函式。
	handler http.Handler
}

// Name 能夠替此路由命名供稍後以反向路由的方式產生路徑。
func (r *Route) Name(name string) *Route {
	r.name = name
	r.routeGroup.router.routeNames[name] = r
	return r
}

// init 能夠初始化這個路由並且解析路徑成片段供服務開始後比對。
func (r *Route) init() *Route {
	// 拆解路由片段。
	r.tearApart()
	return r
}

// AddPriority 會替此路由增加指定的優先度。
func (r *Route) AddPriority(priority int) {
	r.priority += int16(priority)
}

// tearApart 會將路由的路徑逐一拆解成片段供稍後方便使用。
func (r *Route) tearApart() {
	r.isStatic = true
	// 將路徑以 `/` 作為分水嶺來拆開。
	parts := strings.Split(r.path, "/")

	if r.path == "/" {
		r.AddPriority(priorityRoot)
		return
	}

	// 遞迴每個片段，並且分析資料。
	for _, v := range parts {
		if v == "" {
			continue
		}
		// 路由片段數量遞增。
		r.len++
		// 是否為靜態路由。
		var isStatic bool
		// 是否為 `{}` 擷取群組。
		var isCaptureGroup bool
		if strings.Contains(v, "{") {
			isCaptureGroup = true
			r.hasCaptureGroup = true
			r.isStatic = false
		} else {
			isStatic = true
		}
		// 取得前後輟。
		var prefix string
		var suffix string
		if isCaptureGroup {
			left := strings.Split(v, "{")
			right := strings.Split(v, "}")
			// 如果擷取群組左側不是空的，那麼就取得左側內容當作前輟。
			if left[0] != "" {
				prefix = left[0]
			}
			// 如果擷取群組右側側不是空的，那麼就取得右側側內容當作後輟。
			if right[1] != "" {
				suffix = right[1]
			}
			// 移除路徑上的擷取群組符號與固定前後輟。
			v = strings.Split(strings.Split(v, "{")[1], "}")[0]
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
		var rule *rule
		if ruleName != "" {
			rule = r.routeGroup.router.rules[ruleName]
		}
		// 整理此片段。
		r.parts = append(r.parts, &part{
			rule:           rule,
			name:           varName,
			path:           strings.ToLower(v),
			prefix:         strings.ToLower(prefix),
			suffix:         strings.ToLower(suffix),
			isStatic:       isStatic,
			isCaptureGroup: isCaptureGroup,
			isRegExp:       isRegExp,
			isOptional:     isOptional,
		})
		// 如果這個片段有擷取群組的話就建立一個預設的空擷取群組。
		if isCaptureGroup {
			if r.defaultCaptureVars == nil {
				r.defaultCaptureVars = make(map[string]string)
			}
			r.defaultCaptureVars[varName] = ""
		}
		r.AddPriority(priorityPath)
		if prefix != "" {
			r.AddPriority(priorityText)
		}
		if suffix != "" {
			r.AddPriority(priorityText)
		}
		if isCaptureGroup {
			r.AddPriority(priorityGroup)
			if isRegExp {
				r.AddPriority(priorityRegExp)
				if ruleName == "*" {
					r.AddPriority(priorityAnyRegExp)
				}
			}
			if isOptional {
				r.AddPriority(priorityOptional)
			}
		} else {
			r.AddPriority(priorityStatic)
		}
	}
}
