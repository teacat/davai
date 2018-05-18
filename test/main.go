package main

import (
	"net/http"

	davai "github.com/teacat/go-davai"
)

func main() {
	r := davai.New()
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
	r.Run()
}
