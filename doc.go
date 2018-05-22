// Davai（давай）是一個十分快速的 HTTP 路由器，這能夠讓你有效地作為其它套件的基礎核心。
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
// 變數路由能透過 `{}`（花括號）符號可以擷取路由中指定片段的內容並作為指定變數在路由器中使用。
//
//  路由：/user/{name}
//
//  /user/admin                ○
//  /user/admin/profile        ✕
//  /user/                     ✕
package davai
