package davai

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"context"

	rcontext "github.com/gorilla/context"
	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
)

const (
	methodPost    = "POST"
	methodGet     = "GET"
	methodDelete  = "DELETE"
	methodOptions = "OPTIONS"
	methodPut     = "PUT"
	methodPatch   = "PATCH"
)

type testRequest struct {
	Path       string
	Method     string
	StatusCode int
	Body       string
}

func sendTestRequests(a *assert.Assertions, reqs []testRequest) {
	request := gorequest.New()
	for _, r := range reqs {
		if r.StatusCode == 0 {
			r.StatusCode = http.StatusOK
		}
		if r.Method == "" {
			r.Method = methodGet
		}
		var resp gorequest.Response
		var body string
		var errs []error
		switch r.Method {
		case methodGet:
			resp, body, errs = request.Get(r.Path).End()
		case methodPost:
			resp, body, errs = request.Post(r.Path).End()
		case methodDelete:
			resp, body, errs = request.Delete(r.Path).End()
		case methodOptions:
			resp, body, errs = request.Options(r.Path).End()
		case methodPut:
			resp, body, errs = request.Put(r.Path).End()
		case methodPatch:
			resp, body, errs = request.Patch(r.Path).End()
		}
		a.Len(errs, 0)
		a.Equal(r.Body, body)
		a.NotNil(resp)
		a.Equal(r.StatusCode, resp.StatusCode, r.Path)
	}
}

func varsToString(vars map[string]string) string {
	var slice []string

	for _, v := range vars {
		slice = append(slice, v)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
	a := strings.Join(slice, ",")
	return a
}

func TestBasicRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/one", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("One"))
	})
	r.Get("/one/two", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Two"))
	})
	r.Get("/one/two/three", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Three"))
	})
	r.Get("/four-four", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Four"))
	})
	r.Get("/five-five/five-five", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Five"))
	})
	r.Get("/six_six", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Six"))
	})
	r.Get("/seven_", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Seven"))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "One",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "Two",
		},
		{
			Path: "http://localhost:8080/one/two/three",
			Body: "Three",
		},
		{
			Path: "http://localhost:8080/four-four",
			Body: "Four",
		},
		{
			Path: "http://localhost:8080/five-five/five-five",
			Body: "Five",
		},
		{
			Path: "http://localhost:8080/six_six",
			Body: "Six",
		},
		{
			Path: "http://localhost:8080/seven_",
			Body: "Seven",
		},
		{
			Path:       "http://localhost:8080/one/two/three/four",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestBasicMethodRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Get"))
	})
	r.Post("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Post"))
	})
	r.Put("/put", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Put"))
	})
	r.Patch("/patch", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Patch"))
	})
	r.Delete("/delete", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Delete"))
	})
	r.Options("/options", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Options"))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/get",
			Body: "Get",
		},
		{
			Path:   "http://localhost:8080/post",
			Method: methodPost,
			Body:   "Post",
		},
		{
			Path:   "http://localhost:8080/put",
			Method: methodPut,
			Body:   "Put",
		},
		{
			Path:   "http://localhost:8080/patch",
			Method: methodPatch,
			Body:   "Patch",
		},
		{
			Path:   "http://localhost:8080/delete",
			Method: methodDelete,
			Body:   "Delete",
		},
		{
			Path:   "http://localhost:8080/options",
			Method: methodOptions,
			Body:   "Options",
		},
		{
			Path:       "http://localhost:8080/post",
			Method:     methodGet,
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/{one}/{two}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/{one}/{two}/{three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/1",
			Body: "1",
		},
		{
			Path: "http://localhost:8080/1/2",
			Body: "1,2",
		},
		{
			Path: "http://localhost:8080/1/2/3",
			Body: "1,2,3",
		},
		{
			Path:       "http://localhost:8080/1/2/3/4",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestOptionalParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/{one?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("one:" + varsToString(Vars(r))))
	})
	r.Get("/{one}/{two?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two:" + varsToString(Vars(r))))
	})
	r.Get("/{one}/{two}/{three?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("three:" + varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/1",
			Body: "two:,1",
		},
		{
			Path: "http://localhost:8080/1/2",
			Body: "three:,1,2",
		},
		{
			Path: "http://localhost:8080/1/2/3",
			Body: "three:1,2,3",
		},
		{
			Path:       "http://localhost:8080/1/2/3/4",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestOptionalParamRoute2(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/fixed/{one?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("one:" + varsToString(Vars(r))))
	})
	r.Get("/fixed/{one}/fixed", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("one/fixed:" + varsToString(Vars(r))))
	})
	r.Get("/fixed/{one}/fixed/{two?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two:" + varsToString(Vars(r))))
	})
	r.Get("/fixed/{one}/fixed/{two}/fixed", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two/fixed:" + varsToString(Vars(r))))
	})
	r.Get("/fixed/{one}/fixed/{two}/fixed/{three?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("three:" + varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path:       "http://localhost:8080/",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path: "http://localhost:8080/fixed",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/fixed/1",
			Body: "one:1",
		},
		{
			Path: "http://localhost:8080/fixed/1/fixed",
			Body: "two:,1",
		},
		{
			Path: "http://localhost:8080/fixed/1/fixed/2",
			Body: "two:1,2",
		},
		{
			Path: "http://localhost:8080/fixed/1/fixed/2/fixed",
			Body: "three:,1,2",
		},
		{
			Path: "http://localhost:8080/fixed/1/fixed/2/fixed/3",
			Body: "three:1,2,3",
		},
		{
			Path:       "http://localhost:8080/fixed/1/fixed/2/fixed/3/fixed",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Rule("a", "[A-Za-z]+")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{i:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|one:" + varsToString(Vars(r))))
	})
	r.Get("/{a:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("a|one:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{i:two}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|two:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{a:two}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("a|two:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{a:two}/{i:three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|three:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{a:two}/{a:three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("a|three:" + varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/1",
			Body: "i|one:1",
		},
		{
			Path: "http://localhost:8080/a",
			Body: "a|one:a",
		},
		{
			Path: "http://localhost:8080/1/2",
			Body: "i|two:1,2",
		},
		{
			Path: "http://localhost:8080/1/b",
			Body: "a|two:1,b",
		},
		{
			Path: "http://localhost:8080/1/b/3",
			Body: "i|three:1,3,b",
		},
		{
			Path: "http://localhost:8080/1/b/c",
			Body: "a|three:1,b,c",
		},
		{
			Path:       "http://localhost:8080/測試",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/a/2",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestOptionalRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/{i:one?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("one:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{i:two?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{i:two}/{i:three?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("three:" + varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/1",
			Body: "two:,1",
		},
		{
			Path: "http://localhost:8080/1/2",
			Body: "three:,1,2",
		},
		{
			Path: "http://localhost:8080/1/2/3",
			Body: "three:1,2,3",
		},
		{
			Path:       "http://localhost:8080/a",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/a/2",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/a/b/3",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/a/b/c/d",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestAnyRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{*:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("*|one:" + varsToString(Vars(r))))
	})
	r.Get("/two", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two:" + varsToString(Vars(r))))
	})
	r.Get("/{one}/two/{*:three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("*|three:" + varsToString(Vars(r))))
	})
	r.Get("/{one}/two/three/four/{*:five?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("*|five:" + varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "*|one:one",
		},
		{
			Path: "http://localhost:8080/two",
			Body: "two:",
		},
		{
			Path: "http://localhost:8080/one/two/three",
			Body: "*|three:one,three",
		},
		{
			Path: "http://localhost:8080/one/two/three/three/three",
			Body: "*|three:one,three/three/three",
		},
		{
			Path: "http://localhost:8080/one/two/three/four",
			Body: "*|five:,one",
		},
		{
			Path: "http://localhost:8080/1/two/three/four/5",
			Body: "*|five:1,5",
		},
		{
			Path: "http://localhost:8080/1/two/three/four/5/6",
			Body: "*|five:1,5/6",
		},
	})
	r.Shutdown(context.Background())
}

func TestPrefixSuffixRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{one}.sub", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{one}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{one}.suf/pre.{two}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/1.sub",
			Body: "1",
		},
		{
			Path: "http://localhost:8080/pre.1.suf",
			Body: "1",
		},
		{
			Path: "http://localhost:8080/pre.1.suf/pre.2.suf",
			Body: "1,2",
		},
		{
			Path:       "http://localhost:8080/1",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/.sub",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre..suf",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre..suf/pre.1.suf",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre..suf/pre..suf",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestOptionalPrefixSuffixRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{one?}.sub", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{one?}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{one?}.suf/pre.{two?}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{one}.suf/pre.{two}.suf/pre.{three}.suf/pre.{four?}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/1.sub",
			Body: "1",
		},
		{
			Path: "http://localhost:8080/.sub",
			Body: "",
		},
		{
			Path: "http://localhost:8080/pre.1.suf",
			Body: ",1",
		},
		{
			Path: "http://localhost:8080/pre..suf",
			Body: "",
		},
		{
			Path: "http://localhost:8080/pre.1.suf/pre.2.suf",
			Body: "1,2",
		},
		{
			Path: "http://localhost:8080/pre..suf/pre..suf",
			Body: ",",
		},
		{
			Path: "http://localhost:8080/pre.1.suf/pre.2.suf/pre.3.suf/pre..suf",
			Body: ",1,2,3",
		},
		{
			Path: "http://localhost:8080/pre.1.suf/pre.2.suf/pre.3.suf/pre.4.suf",
			Body: "1,2,3,4",
		},
		{
			Path:       "http://localhost:8080/1",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/1/2",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre.1",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre.1/pre.2.suf",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestPrefixSuffixRegExRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Rule("a", "[A-Za-z]+")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{a:one?}.sub", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/pre.{a:one?}.suf", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/foo.{a:one}.bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/a.sub",
			Body: "a",
		},
		{
			Path: "http://localhost:8080/.sub",
			Body: "",
		},
		{
			Path: "http://localhost:8080/pre.a.suf",
			Body: "a",
		},
		{
			Path: "http://localhost:8080/pre..suf",
			Body: "",
		},
		{
			Path: "http://localhost:8080/foo.a.bar",
			Body: "a",
		},
		{
			Path:       "http://localhost:8080/1",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/1.sub",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/pre.1.suf",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/foo..bar",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/foo.1.bar",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestLookbehindRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Rule("e", "(?:.html)")
	r.Get("/detail{e:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Rule("u", "(?:.html)?")
	r.Get("/user{u:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/detail.html",
			Body: ".html",
		},
		{
			Path: "http://localhost:8080/user",
			Body: "",
		},
		{
			Path: "http://localhost:8080/user.html",
			Body: ".html",
		},
		{
			Path:       "http://localhost:8080/detail",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/user.htmli",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/user.htm",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}
func TestMiddleware(t *testing.T) {
	assert := assert.New(t)
	r := New()
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo1", "bar1")
			next.ServeHTTP(w, r)
		})
	}
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo2", "bar2")
			next.ServeHTTP(w, r)
		})
	}
	r.Get("/", middleware, middleware2, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(rcontext.Get(r, "foo1").(string) + rcontext.Get(r, "foo2").(string)))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "bar1bar2",
		},
	})
	r.Shutdown(context.Background())
}

func TestRouteGroup(t *testing.T) {
	assert := assert.New(t)
	r := New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}
	v1 := r.Group("/v1")
	{
		v1.Get("/", handler)
		v1.Get("/user/{id}", handler)
		v1.Get("/post/{title?}", handler)
		v1.Get("/login", handler)
	}
	v2 := r.Group("/v2")
	{
		v2.Get("/", handler)
		v2.Get("/user/{id}", handler)
		v2.Get("/post/{title?}", handler)
		v2.Get("/login", handler)
	}
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/v1/user/123",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v1",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v1/post",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v1/post/1234",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v1/login",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v2/user/123",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v2",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v2/post",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v2/post/1234",
			Body: "OK",
		},
		{
			Path: "http://localhost:8080/v2/login",
			Body: "OK",
		},
		{
			Path:       "http://localhost:8080/v1/user",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/v1/post/123/456",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestRouteGroupMiddleware(t *testing.T) {
	assert := assert.New(t)
	r := New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(rcontext.Get(r, "foo1").(string) + rcontext.Get(r, "foo2").(string) + rcontext.Get(r, "foo3").(string)))
	}
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo1", "bar1")
			next.ServeHTTP(w, r)
		})
	}
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo2", "bar2")
			next.ServeHTTP(w, r)
		})
	}
	middleware3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo3", "bar3")
			next.ServeHTTP(w, r)
		})
	}
	v1 := r.Group("/v1", middleware, middleware2)
	{
		v1.Get("/user/{id}", middleware3, handler)
		v1.Get("/post/{title?}", middleware3, handler)
		v1.Get("/login", middleware3, handler)
	}
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/v1/user/123",
			Body: "bar1bar2bar3",
		},
		{
			Path: "http://localhost:8080/v1/post",
			Body: "bar1bar2bar3",
		},
		{
			Path: "http://localhost:8080/v1/post/1234",
			Body: "bar1bar2bar3",
		},
		{
			Path: "http://localhost:8080/v1/login",
			Body: "bar1bar2bar3",
		},
	})
	r.Shutdown(context.Background())
}

func TestGlobalMiddleware(t *testing.T) {
	assert := assert.New(t)
	r := New()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(rcontext.Get(r, "foo1").(string) + rcontext.Get(r, "foo2").(string) + rcontext.Get(r, "foo3").(string)))
	}
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo1", "bar1")
			next.ServeHTTP(w, r)
		})
	}
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo2", "bar2")
			next.ServeHTTP(w, r)
		})
	}
	middleware3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rcontext.Set(r, "foo3", "bar3")
			next.ServeHTTP(w, r)
		})
	}
	r.Use(middleware)
	r.Get("/", middleware2, middleware3, handler)
	v1 := r.Group("/foo", middleware2)
	{
		v1.Get("/bar", middleware3, handler)
	}
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "bar1bar2bar3",
		},
		{
			Path: "http://localhost:8080/foo/bar",
			Body: "bar1bar2bar3",
		},
	})
	r.Shutdown(context.Background())
}

func TestGenerateRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/one", func(w http.ResponseWriter, r *http.Request) {}).Name("One")
	r.Get("/one/two", func(w http.ResponseWriter, r *http.Request) {}).Name("Two")
	r.Get("/one/two/{three}", func(w http.ResponseWriter, r *http.Request) {}).Name("Three")
	r.Get("/one/two/{three}/{four}", func(w http.ResponseWriter, r *http.Request) {}).Name("Four")
	r.Get("/one/two/{three}/{four}/{five?}", func(w http.ResponseWriter, r *http.Request) {}).Name("Five")
	r.Get("/one/two/{three}/{four}/{five?}/{s:six}", func(w http.ResponseWriter, r *http.Request) {}).Name("Six")
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/one",
			Body: "",
		},
	})
	assert.Equal("/one", r.Generate("One"))
	assert.Equal("/one/two", r.Generate("Two"))
	assert.Equal("/one/two", r.Generate("Three"))
	assert.Equal("/one/two/1", r.Generate("Three", map[string]string{
		"three": "1",
	}))
	assert.Equal("/one/two/1/2", r.Generate("Four", map[string]string{
		"three": "1",
		"four":  "2",
	}))
	assert.Equal("/one/two", r.Generate("Five"))
	assert.Equal("/one/two/1", r.Generate("Five", map[string]string{
		"three": "1",
	}))
	assert.Equal("/one/two/1/2", r.Generate("Five", map[string]string{
		"three": "1",
		"four":  "2",
	}))
	assert.Equal("/one/two/1/2/3", r.Generate("Five", map[string]string{
		"three": "1",
		"four":  "2",
		"five":  "3",
	}))
	assert.Equal("/one/two/1/2/3/4", r.Generate("Six", map[string]string{
		"three": "1",
		"four":  "2",
		"five":  "3",
		"six":   "4",
	}))
	r.Shutdown(context.Background())
}

func TestStaticDirRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	f, err := filepath.Abs(".")
	assert.NoError(err)
	assert.NoError(os.Chdir(f))
	r.Get("/{*:filename}", http.FileServer(http.Dir("test")))
	r.Get("/static/{*:filename}", http.StripPrefix("/static", http.FileServer(http.Dir("test"))))
	r.Get("/prefix/{*:filename}", http.StripPrefix("/prefix", http.FileServer(http.Dir("."))))
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestServeFilesDirRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	f, err := filepath.Abs(".")
	assert.NoError(err)
	assert.NoError(os.Chdir(f))
	r.ServeFiles("/", http.Dir("test"))
	r.ServeFiles("/static", http.Dir("test"))
	r.ServeFiles("/prefix", http.Dir("."))
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestServeFilesRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	f, err := filepath.Abs(".")
	assert.NoError(err)
	assert.NoError(os.Chdir(f))
	r.ServeFiles("/", "test").DirectoryListing = true
	r.ServeFiles("/static", "test").DirectoryListing = true
	r.ServeFiles("/prefix", ".").DirectoryListing = true
	r.ServeFiles("/nolisting", "test")
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/nolisting/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/nolisting/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
		{
			Path:       "http://localhost:8080/nolisting",
			StatusCode: http.StatusForbidden,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/nolisting/",
			StatusCode: http.StatusForbidden,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/nolisting/directory",
			StatusCode: http.StatusForbidden,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/nolisting/directory/",
			StatusCode: http.StatusForbidden,
			Body:       "",
		},
	})
	r.Shutdown(context.Background())
}

func TestServeFileRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	f, err := filepath.Abs(".")
	assert.NoError(err)
	assert.NoError(os.Chdir(f))
	r.ServeFile("/one", "test/file.txt")
	r.ServeFile("/one/two", "test/directory/file.txt")
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/one",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/one/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/one//",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/one/two/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/one/two//",
			Body: "This is directory/file.txt.",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestServeFilesAndStaticRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	f, err := filepath.Abs(".")
	assert.NoError(err)
	assert.NoError(os.Chdir(f))
	r.Get("/one", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("One"))
	})
	r.Get("/{i:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{i:one}"))
	})
	r.ServeFiles("/", http.Dir("test"))
	r.ServeFiles("/resource/static", http.Dir("test"))
	r.ServeFiles("/static", http.Dir("test"))
	r.ServeFiles("/prefix", http.Dir("."))
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/one",
			Body: "One",
		},
		{
			Path: "http://localhost:8080/123",
			Body: "{i:one}",
		},
		{
			Path: "http://localhost:8080/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt/",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt/",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/resource/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/static/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path: "http://localhost:8080/prefix/test/",
			Body: "<pre>\n<a href=\"directory/\">directory/</a>\n<a href=\"file.txt\">file.txt</a>\n<a href=\"main.go\">main.go</a>\n</pre>\n",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: http.StatusNotFound,
			Body:       "404 page not found\n",
		},
	})
	r.Shutdown(context.Background())
}

func TestTrailingSlashesRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/one", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("One"))
	})
	r.Get("/one/two", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Two"))
	})
	go func() {
		err := r.Run()
		if err != nil && err != http.ErrServerClosed {
			assert.NoError(err)
		}
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "One",
		},
		{
			Path: "http://localhost:8080/one/",
			Body: "One",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "Two",
		},
		{
			Path: "http://localhost:8080/one/two/",
			Body: "Two",
		},
	})
	r.Shutdown(context.Background())
}
