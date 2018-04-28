package main

import (
	"net/http"

	davai "github.com/teacat/go-davai"
)

func main() {
	r := davai.New()
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to `/hello` path!"))
	})
	r.Get("/hello/world", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("And this is the `/hello/world` path."))
	})
	r.Run()
}
