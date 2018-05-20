# Davai [![GoDoc](https://godoc.org/github.com/teacat/go-davai?status.svg)](https://godoc.org/github.com/teacat/go-davai)

基於 `net/http` 的 Golang 基本 HTTP 路由，這個套件試圖提供最核心且具動態路由功能的路由器。

# 這是什麼？

Davai（давай）是一個十分快速的 HTTP 路由器，這能夠讓你有效地作為其它套件的基礎核心。

* 支援中介軟體（Middleware）。
* 極具動態的路由擷取、前後輟功能。
* 相容 `net/http` 的原生用法而無需重新學習。
* 反向路由能從變數產生網址。
* 路由群組以避免繁雜的重複手續。
* 正規表達式快取提升效能。

# 為什麼？

多數的 Golang 路由器以 `:name` 作為路徑擷取的用法，但這種用法導致你不能夠擁有固定前後輟（例如：`user-:id.json`），而 Davai 解決了這個問題，且讓路由擷取變得更加多元化、也更適合 RESTful API 設計。

在不少路由器比對路徑時，都會從上搜尋到下（即為：從多到少）。就算是個最基本的 `/` 路徑都必須先從最長且可能帶有正規表達式的動態路由開始探索，這浪費了不少的時間，而 Davai 將動態和靜態路由區分進而減少效能損耗。

額外一點在於 Davai 能夠快取部分網址來避免重複的正規表達式驗證、且相容原生的 `net/http` 函式讓使用設計更加地通用。

# 效能比較

這裡有份簡略化的效能測試報表。

```
測試規格：
1.7 GHz Intel Core i7 (4650U)
8 GB 1600 MHz DDR3
```

# 索引

* [安裝方式](#安裝方式)
* [使用方式](#使用方式)
    * [變數路由](#變數路由)
    * [選擇性路由](#選擇性路由)
	* [前後輟路由](#前後輟路由)
	* [任意路由](#任意路由)
    * [正規表達式路由](#正規表達式路由)
        * [自訂規則](#自訂規則)
    * [路由群組](#路由群組)
    * [反向與命名路由](#反向與命名路由)
    * [中介軟體](#中介軟體)
    * [靜態檔案與目錄](#靜態檔案與目錄)
    * [無路由](#無路由)
* [如何運作的？](#如何運作的？)

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

```
路由：/user/{name}

/user/admin                ○
/user/admin/profile        ✕
/user/                     ✕
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

```
路由：/user/{name?}

/user/                     ○
/user/admin                ○
/user/admin/profile        ✕
```

## 前後輟路由

擷取路由的前、後可以參雜靜態文字，這讓你很好設計一個基於 `.json` 副檔名的 RESTful API 系統。

```
路由：/api/resource-{id}.json

/api/resource-123.json     ○
/api/resource-.json        ✕
/api/                      ✕
```

## 任意路由

透過 `*` 規則可以讓正規表達式符合任何型態的路徑。當這個規則被擺放在路由的最後面時即會成為「任意路由」，在這種情況下任何路徑都會符合。

```
路由：/src/{*:filename}

/src/                      ○
/src/example.png           ○
/src/subdir/example.png    ○
/api/                      ✕
```

## 正規表達式路由

透過正規表達式路由可以更精準地表明路由應該要符合哪種格式，Davai 預設有數種正規表達式規則：`i`（數字）、`s`（數字與英文字母）。

在變數路由名稱的前面加上 `:` 來表明欲使用的正規表達式規則，其格式為 `{規則:變數名稱}`。用上正規表達式後亦能在變數名稱後加上 `?`（問號）來作為選擇性路由。

```
路由：/user/{i:id}

/user/1234                 ○
/user/                     ✕
/user/profile              ✕
/user/1234/profile         ✕
```

### 自訂規則

如果 Davai 預設的正規表達式規則不合乎你的需求，可以考慮透過 `Rule` 來追加新的正規表達式規則。

```go
func main() {
	d := davai.New()
	// 透過 `AddRule` 可以追加新的正規表達式規則。
	d.Rule("r", "[0-9a-z]++")
	// 接著就能夠直接在路由中使用。
	d.Get("/post/{r:name}", PostHandler)
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

或者乾脆一點用於所有路由。

```go
func main() {
	d := davai.New()
	// 將多個中介軟體作為全域用途並在所有路由中執行。
	d.Use(MyMiddleware, AnotherMiddlewate)
	// ...
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

# 如何運作的？

Davai 有兩個路由樹，分別是「靜態」與「動態」。當定義一個路由的時候，Davai 會事先解析該路由中的所有片段，並且轉換成靜態資料以提升後續的查詢效能。當接收到一個請求的時候，Davai 會優先找尋靜態路由中是否有符合此路徑的路由，若無則會逐一搜尋動態路由並且相互比對。

每個路由樹都有多個路由並且依照優先程度進行排列，簡單地說：「越長的路由就越會先被執行」。優先順序的定義以路由中的每個片段加總為主：

```
優先度    定義
20       根目錄
16       路徑片段
8        靜態路徑
4        擷取群組
2        帶前後輟
1        正規表達式
-1       可選片段
-2       任意片段
```

這有點類似 Unix 的權限（`1`、`2`、`4`）計算方式，也因為越長的路由會有更多分數，所以就會先被執行、比對；而越怠惰的路由分數就會越低。實際的路由優先順序程度如下：

```
優先度    路由
69       /user/{s:name}/profile
48       /user/admin
24       /user
20       /
19       /{*:path}
```

靜態路由的比對十分簡單，就是直接透過 `map` 的鍵值來比對是否有此路由。然而動態路由則不同，動態路由會將網址拆分成多個片段，並且將其從左至右、從長到短逐一比對，並且判斷下一個片段是否為可選片段，若是可選片段且網址已經到底的話就算符合（因為可選）。

然而 Davai 也支援路由片段的前後輟，這個判別方法十分簡單和迅速。定義路由時 Davai 會事先分析路由片段的前後輟為何，並且保存起來供之後比對請求網址。

```
   片段                  片段
+--------+------------+---------+--------+
|        |     前輟    | 擷取群組 |   後輟  |
+--------+============+---------+========+
/  user  /  resource- |   {id}  | .json  /
+--------+============+---------+========+
```

比對時 Davai 會透過 `strings.HasPrefix` 和 `strings.HasSuffix` 函式來確認該路由是否有指定的前後輟，並且將接收到的網址片段透過 `strings.TrimLeft` 與 `strings.TrimRight` 來移除前後輟，最終就能取得路由中的擷取群組內容。