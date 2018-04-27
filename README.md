# Davai [![GoDoc](https://godoc.org/github.com/teacat/go-davai?status.svg)](https://godoc.org/github.com/teacat/go-davai)

Davai（давай）是個基於 `net/http` 的 Golang 基本 HTTP 路由，這個套件試圖提供最核心且具動態路由功能的路由器。

# 索引

* [效能比較](#效能比較)
* [安裝方式](#安裝方式)
* [使用方式](#使用方式)
    * [變數路由](#變數路由)
    * [選擇性路由](#選擇性路由)
    * [正規表達式路由](#正規表達式路由)
        * [自訂規則](#自訂規則)
    * [路由群組](#路由群組)
    * [反向與命名路由](#反向與命名路由)
    * [中介軟體](#中介軟體)
    * [靜態檔案與目錄](#靜態檔案與目錄)
    * [無路由](#無路由)
* [比對優先度](#比對優先度)

# 效能比較

xx

#

|           範例            | 支援 |          說明         |       會相符的範例       |
| -----------------------  | ---- | -------------------- | ---------------------- |
| `/{*:title}`             |   ○  | 基於根目錄的隨意路由。   | `/hello`、`/foo/bar`    |
| `/`                      |   ○  | 根目錄。               | `/`                    |
| `/products`              |   ○  | 靜態路由。              | `/products`            |
| `/{page}.html`           |   ○  | 擷取路由和固定後輟。     | `/foo.html`            |
| `/user/{i:id?}`          |   ○  | 選擇性擷取路由。         | `/user`、`/user/58`    |
| `/album/{i:id}/detail`   |   ○  | 靜態路由和正規表達式路由。 | `/album/162/detail`    |
| `/api/user-{id}.json`    |   ○  | 固定前、後輟的擷取路由。   | `/api/user-admin.json` |
| `/{type}-{id}.html`      |   ✕  | 雙重擷取路由。           | `/tshirt-3.html`       |

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/teacat/go-davai
```

# 使用方式

透過 `davai.New` 建立一個新的路由器，並且以 `Get`、`Post`⋯等方法來建立基於不同路由的處理函式，接著將路由器傳入 `http.Handle` 來開始監聽服務。

```go
package main

import (
	"net/http"

	davai "github.com/teacat/go-davai"
)

func main() {
	d := davai.New()
	d.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})
	d.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})
	d.Post("/album", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})
	http.Handle("/", d)
}
```

## 變數路由

透過 `{}`（花括號）符號可以擷取路由中指定片段的內容並作為指定變數在路由器中使用。

```go
func main() {
	d := davai.New()
	// 這個路由與 `/1234`、`/hello` 相符。
	d.Get("/{name}", IndexHandler)
	// 這個路由會和 `/posts/1234`、`/posts/hello` 相符。
	d.Get("/posts/{title}", PostsHandler)
	http.Handle("/", r)
}
```

在路由中以 `davai.Vars` 並傳入 `*http.Request` 來取得在路由中所擷取的變數。

```go
func main() {
	d := davai.New()
	d.Get("/post/{title}", func(w http.ResponseWriter, r *http.Request) {
		// 透過 `davai.Vars` 並傳入 HTTP 請求的建構體就能夠取得已擷取的變數。
		vars := davai.Vars(r)
		fmt.Println(vars["title"])
	})
	http.Handle("/", d)
}
```

## 選擇性路由

如果擷取的變數並不一定是必要的，那麼可以在變數名稱後加上 `?` 來作為「選擇性變數」。

```go
func main() {
	d := davai.New()
	// 這個路由與 `/user`、`/user/1234`、`/user/admin` 相符。
	d.Get("/user/{name?}", UserHandler)
	// 這個路由與 `/post`、`/post/1234`、`/post/my-life` 相符。
	d.Get("/post/{title?}", PostHandler)
	http.Handle("/", d)
}
```

## 正規表達式路由

透過正規表達式路由可以更精準地表明路由應該要符合哪種格式，Davai 預設有數種正規表達式規則。

|格式 | 表達式 | 說明 | 範例 |
| -- | -- | -- | -- |
| * | .+? | 任何東西。 | `/post/{*:title}` |
| i | [0-9]++ | 僅數字。 | `/user/{i:userId}` |
| s | [0-9A-Za-z]++ | 數字和大小寫英文字母。 | `/profile/{s:username}` |

在變數路由名稱的前面加上 `:` 來表明欲使用的正規表達式規則，其格式為 `{規則:變數名稱}`。用上正規表達式後亦能在變數名稱後加上 `?`（問號）來作為選擇性路由。

```go
func main() {
	d := davai.New()
	// 使用 Davai 的預設正規表達式規則。
	d.Get("/user/{i:id}", UserHandler)
	d.Get("/user/{s:name?}", UserHandler)
	http.Handle("/", d)
}
```

### 自訂規則

如果 Davai 預設的正規表達式規則不合乎你的需求，可以考慮透過 `Rule` 來追加新的正規表達式規則。

```go
func main() {
	d := davai.New()
	// 透過 `AddRule` 可以追加新的正規表達式規則。
	d.Rule("s", "[0-9a-z]++")
	// 接著就能夠直接在路由中使用。
	d.Get("/post/{s:name}", PostHandler)
	http.Handle("/", d)
}
```

## 路由群組

如果有些路由的前輟、中介軟體是一樣的話那麼就可以建立一個路由群組來省去重複的手續。

```go
func main() {
	d := davai.New()
	// 透過 `Group` 可以替路由建立一個通用的前輟。
	v1 := d.Group("/v1")
	{
		// 這個路由與 `/v1/user/1234`、`/v1/user/admin` 相符。
		v1.Post("/user/{id}", UserHandler)
		// 這個路由與 `/v1/post`、`/v1/post/1234`、`/v1/post/my-life` 相符。
		v1.Post("/post/{title?}", PostHandler)
		// 這個路由與 `/v1/login` 相符。
		v1.Post("/login", LoginHandler)
	}
	// 路由群組也能夠有多個，這很適合用在設計 API 並區分版本上。
	v2 := d.Group("/v2")
	{
		v2.Post("/user/{id}", UserHandler)
		v2.Post("/post/{title?}", PostHandler)
		v2.Post("/login", LoginHandler)
	}
	http.Handle("/", d)
}
```

## 反向與命名路由

替定義好的路由命名，就能夠在稍後透過此名稱並傳入變數來反向產生該路由。

```go
func main() {
	d := davai.New()
	// 在路由後面透過 `Name` 函式來替定義好的路由命名。
	d.Get("/product/{title}/{type}/{id}", ProductHandler).Name("product")
	// 透過 `Generate` 可以傳入變數並產生已命名的路由。
	path := d.Generate("product", map[string]interface{}{
		"title": "t-shirt",
		"type":  "large",
		"id":    152,
	})
	// 結果：/product/t-shirt/large/152
	fmt.Println(path)
}
```

## 中介軟體

中介軟體也稱作中介層，這能夠在單個路由中執行多個處理函式並串在一起。

```go
func MyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
		// 呼叫 `net.ServeHTTP` 來呼叫下一個中介軟體或者是處理函式。
		// 如果不這麼做的話則會中斷繼續。
		next.ServeHTTP(w, r)
	})
}

func main() {
	d := davai.New()
	// 將 `MyMiddleware` 中介軟體安插於路由中。
	d.Get("/post", MyMiddleware, UserHandler)
	d.Get("/album", MyMiddleware, AlbumHandler)
	http.Handle("/", d)
}
```

如果一個中介軟體會用於很多個路由，那麼可以考慮替他們建立一個路由群組。

```go
func MyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
	})
}

func main() {
	d := davai.New()
	// 中介軟體可以安插於路由群組，這樣就不需要替每個路由安插一個中介軟體。
	v1 := d.Group("/v1", MyMiddleware)
	{
		v1.Post("/user", UserHandler)
		v1.Post("/login", LoginHandler)
	}
	http.Handle("/", d)
}
```

## 靜態檔案與目錄

```go
```

## 無路由

以 `NoRoute` 來傳入一個處理函式，這個函式會在客戶端呼叫了不存在的路由時所執行。

```go
func main() {
	d := davai.New()
	// 透過 `NoRoute` 指定當客戶端呼叫了不存在的路由時應該對應的處理函式。
	d.NoRoute(NoRouteHandler)
	http.Handle("/", d)
}
```

# 比對優先度

路由的比對有一定的優先順序，這個順序會從「最長」的路由往下到「最短」的來依序比對。

```go
func main() {
	d := davai.New()
	// 順序：1
	d.Get("/user/list/{i:id}", UserListHandler)
	// 順序：2
	d.Get("/user/list/", UserListHandler)
	// 順序：3
	d.Get("/user/{s:name}", UserHandler)
	// 順序：4
    d.Get("/{*:name}", AnyHandler)
    // 順序：5
	d.Get("/{*:name}", RootHandler)
	http.Handle("/", d)
}
```
