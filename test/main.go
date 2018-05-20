package main

import (
	"fmt"
	"net/http"
	"time"

	davai "github.com/teacat/go-davai"
)

func main() {
	r := davai.New()
	MyMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 從接收到請求的時候就開始記錄時間。
			start := time.Now()
			// ...
			// 呼叫 `net.ServeHTTP` 來呼叫下一個中介軟體或者是處理函式。
			// 如果不這麼做的話則會中斷繼續。
			next.ServeHTTP(w, r)
			// 取得本次請求的總費時。
			latency := time.Since(start)
			fmt.Println(latency)
		})
	}

	r.Get("/", MyMiddleware, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root!"))
	})
	r.Get("/{*:path}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root!"))
	})
	r.Get("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root!"))
	})
	r.Get("/user/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root!"))
	})
	r.Get("/user/{s:name}/profile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v", davai.Vars(r))
		w.Write([]byte("Root!"))
	})
	/*
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Root!"))
		})
		r.Get("/{*:any}/bar", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("The world of the `/{*:any}/bar` path."))
		})
		r.Get("/{*:any}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Wow, very lazy, such `/{*:any}`."))
		})
		r.Get("/foo/{*:any}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Foo-any! The `/foo/{*:any}` path."))
		})
		r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome to the `/hello` path!"))
		})
		r.Get("/{name}.json", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("The fake `/{name}.json` path, useful for the API design though."))
		})
		r.Get("/user-{name}.json", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Say hello to the prefixed, suffixed `/user-{name}.json` path."))
		})
		r.Get("/foo/bar", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("And this is the `/foo/bar` path."))
		})
		r.Get("/first/second/third", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Here comes the `/first/second/third` path."))
		})
		r.Get("/first/second/{s:third?}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("So here it is. The `/first/second/{s:third?}` path."))
		})
		r.Get("/{i:number}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("TADA, it's the `/{i:number}` path."))
		})
		r.Get("/{s:string}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, so you just visited the `/{s:string}` path."))
		})
		r.Get("/{i:number}/bar", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Oh, here's the `/{i:number}/bar` path."))
		})
	*/
	r.Run()
}
