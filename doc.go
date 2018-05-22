// Package davai 是一個十分快速的 HTTP 路由器，這能夠讓你有效地作為其它套件的基礎核心。
//
// 一個最簡單的使用範例如下：
//
//   package main
//
//   import (
//   	"net/http"
//
//   	davai "github.com/teacat/go-davai"
//   )
//
//   func main() {
//   	d := davai.New()
//   	d.Get("/", func(w http.ResponseWriter, r *http.Request) {
//   		// ...
//   	})
//   	d.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
//   		// ...
//   	})
//   	d.Post("/album", func(w http.ResponseWriter, r *http.Request) {
//   		// ...
//   	})
//   	d.Run()
//   }
//
// 透過 `{}`（花括號）符號可以擷取路由中指定片段的內容並作為指定變數在路由器中使用。
//
//  路由：/user/{name}
//
//  /user/admin                ○
//  /user/admin/profile        ✕
//  /user/                     ✕
//
// 如果擷取的變數並不一定是必要的，那麼可以在變數名稱後加上 `?` 來作為「選擇性變數」。
//
//  路由：/user/{name?}
//
//  /user/                     ○
//  /user/admin                ○
//  /user/admin/profile        ✕
//
// 擷取路由的前、後可以參雜靜態文字，這讓你很好設計一個基於 `.json` 副檔名的 RESTful API 系統。
//
//  路由：/api/resource-{id}.json
//
//  /api/resource-123.json     ○
//  /api/resource-.json        ✕
//  /api/                      ✕
//
// 透過 `*` 規則可以讓正規表達式符合任何型態的路徑。當這個規則被擺放在路由的最後面時即會成為「任意路由」，在這種情況下任何路徑都會符合。
//
//  路由：/src/{*:filename}
//
//  /src/                      ○
//  /src/example.png           ○
//  /src/subdir/example.png    ○
//  /api/                      ✕
//
// 透過正規表達式路由可以更精準地表明路由應該要符合哪種格式，Davai 預設有數種正規表達式規則：`i`（數字）、`s`（數字與英文字母）。
//
// 在變數路由名稱的前面加上 `:` 來表明欲使用的正規表達式規則，其格式為 `{規則:變數名稱}`。用上正規表達式後亦能在變數名稱後加上 `?`（問號）來作為選擇性路由。
//
//  路由：/user/{i:id}
//
//  /user/1234                 ○
//  /user/                     ✕
//  /user/profile              ✕
//  /user/1234/profile         ✕
//
// 以 `davai.Vars` 並傳入 `*http.Request` 來取得在路由中所擷取的變數。
//
//  d.Get("/post/{title}", func(w http.ResponseWriter, r *http.Request) {
//  	// 透過 `davai.Vars` 並傳入 HTTP 請求的建構體就能夠取得已擷取的變數。
//  	vars := davai.Vars(r)
//  	// 存取 `vars` 來取得網址中的變數。
//  	// 如果該變數是選擇性的，在沒有該變數的情況下會是一個空字串值。
//  	fmt.Println(vars["title"])
//  })
//
package davai
