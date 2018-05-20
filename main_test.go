package davai

import (
	"net/http"
	"strings"
	"testing"

	"context"

	rcontext "github.com/gorilla/context"
	"github.com/parnurzeal/gorequest"
	"github.com/stretchr/testify/assert"
)

const (
	statusOK       = 200
	statusNotFound = 404
	methodPost     = "POST"
	methodGet      = "GET"
	methodDelete   = "DELETE"
	methodOptions  = "OPTIONS"
	methodPut      = "PUT"
	methodPatch    = "PATCH"
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
			r.StatusCode = statusOK
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
		a.Equal(body, r.Body)
		a.Equal(r.StatusCode, resp.StatusCode, r.Path)
	}
}

func varsToString(vars map[string]string) string {
	var slice []string
	for _, v := range vars {
		slice = append(slice, v)
	}
	return strings.Join(slice, ",")
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
		assert.NoError(r.Run())
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
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
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
			Method:     methodPost,
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "one",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "one,two",
		},
		{
			Path: "http://localhost:8080/one/two/three",
			Body: "one,two,three",
		},
		{
			Path:       "http://localhost:8080/one/two/three/four",
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "two:one",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "three:one,two",
		},
		{
			Path: "http://localhost:8080/one/two/three",
			Body: "three:one,two,three",
		},
		{
			Path:       "http://localhost:8080/one/two/three/four",
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path:       "http://localhost:8080/",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path: "http://localhost:8080/fixed",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/fixed/one",
			Body: "one:one",
		},
		{
			Path: "http://localhost:8080/fixed/one/fixed",
			Body: "two:one",
		},
		{
			Path: "http://localhost:8080/fixed/one/fixed/two",
			Body: "two:one,two",
		},
		{
			Path: "http://localhost:8080/fixed/one/fixed/two/fixed",
			Body: "three:one,two",
		},
		{
			Path: "http://localhost:8080/fixed/one/fixed/two/fixed/three",
			Body: "three:one,two,three",
		},
		{
			Path:       "http://localhost:8080/fixed/one/fixed/two/fixed/three/fixed",
			StatusCode: statusNotFound,
			Body:       "",
		},
	})
	r.Shutdown(context.Background())
}

func TestRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{i:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|one:" + varsToString(Vars(r))))
	})
	r.Get("/{s:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("s|one:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{i:two}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|two:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{s:two}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("s|two:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{s:two}/{i:three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("i|three:" + varsToString(Vars(r))))
	})
	r.Get("/{i:one}/{s:two}/{s:three}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("s|three:" + varsToString(Vars(r))))
	})
	go func() {
		assert.NoError(r.Run())
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
			Path: "http://localhost:8080/one",
			Body: "s|one:one",
		},
		{
			Path: "http://localhost:8080/1/2",
			Body: "i|two:1,2",
		},
		{
			Path: "http://localhost:8080/1/two",
			Body: "s|two:1,two",
		},
		{
			Path: "http://localhost:8080/1/two/3",
			Body: "i|three:1,two,3",
		},
		{
			Path: "http://localhost:8080/1/two/three",
			Body: "s|three:1,two,three",
		},
		{
			Path:       "http://localhost:8080/測試",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/one/2",
			StatusCode: statusNotFound,
			Body:       "",
		},
	})
	r.Shutdown(context.Background())
}

func TestCustomRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Rule("a", "[0-9A-Za-z]++")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{a:one}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("a|one:" + varsToString(Vars(r))))
	})
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "a|one:one",
		},
		{
			Path:       "http://localhost:8080/1",
			StatusCode: statusNotFound,
			Body:       "",
		},
	})
	r.Shutdown(context.Background())
}

func TestOptionalRegExParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/{s:one?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("one:" + varsToString(Vars(r))))
	})
	r.Get("/{s:one}/{s:two?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("two:" + varsToString(Vars(r))))
	})
	r.Get("/{s:one}/{s:two}/{s:three?}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("three:" + varsToString(Vars(r))))
	})
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "one:",
		},
		{
			Path: "http://localhost:8080/one",
			Body: "two:one",
		},
		{
			Path: "http://localhost:8080/one/two",
			Body: "three:one,two",
		},
		{
			Path: "http://localhost:8080/one/two/three",
			Body: "three:one,two,three",
		},
		{
			Path:       "http://localhost:8080/1",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/one/2",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/one/two/3",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/one/two/three/four",
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
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
			Body: "two",
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
			Body: "*|five:one",
		},
		{
			Path: "http://localhost:8080/one/two/three/four/five",
			Body: "*|five:one,five",
		},
	})
	r.Shutdown(context.Background())
}

func TestRemixParamRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root"))
	})
	r.Get("/{type}.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/{type}/{id}.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/{type}/{id}-{anotherID}.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Get("/{type}/{id}-{anotherID}-{andAnotherID?}.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	r.Rule("e", "(?:.html)?")
	r.Get("/{type}/{id}/detail{e:extension}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(varsToString(Vars(r))))
	})
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "Root",
		},
		{
			Path: "http://localhost:8080/type.json",
			Body: "type",
		},
		{
			Path: "http://localhost:8080/type/id.json",
			Body: "type,id",
		},
		{
			Path: "http://localhost:8080/type/id-anotherID.json",
			Body: "type,id,anotherID",
		},
		{
			Path: "http://localhost:8080/type/id-anotherID-andAnotherID.json",
			Body: "type,id,anotherID,andAnotherID",
		},
		{
			Path: "http://localhost:8080/type/id-anotherID-.json",
			Body: "type,id,anotherID",
		},
		{
			Path: "http://localhost:8080/type/id/detail",
			Body: "type,id",
		},
		{
			Path: "http://localhost:8080/type/id/detail.html",
			Body: "type,id,.html",
		},
		{
			Path:       "http://localhost:8080/type",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/type.html",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/type/id.html",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/type/id-.json",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/type/id-anotherID-",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/type/id/detail.htm",
			StatusCode: statusNotFound,
			Body:       "",
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
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/",
			Body: "foo1foo2",
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
		v1.Post("/user/{id}", handler)
		v1.Post("/post/{title?}", handler)
		v1.Post("/login", handler)
	}
	v2 := r.Group("/v2")
	{
		v2.Post("/user/{id}", handler)
		v2.Post("/post/{title?}", handler)
		v2.Post("/login", handler)
	}
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/v1/user/123",
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
			Path:       "http://localhost:8080/v1",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/v1/user",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/v1/post/123/456",
			StatusCode: statusNotFound,
			Body:       "",
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
		v1.Post("/user/{id}", middleware3, handler)
		v1.Post("/post/{title?}", middleware3, handler)
		v1.Post("/login", middleware3, handler)
	}
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/v1/user/123",
			Body: "foo1foo2foo3",
		},
		{
			Path: "http://localhost:8080/v1/post",
			Body: "foo1foo2foo3",
		},
		{
			Path: "http://localhost:8080/v1/post/1234",
			Body: "foo1foo2foo3",
		},
		{
			Path: "http://localhost:8080/v1/login",
			Body: "foo1foo2foo3",
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
		assert.NoError(r.Run())
	}()
	assert.Equal("/one", r.Generate("One"))
	assert.Equal("/one/two", r.Generate("Two"))
	assert.Equal("/one/two", r.Generate("Three"))
	assert.Equal("/one/two/three", r.Generate("Three", map[string]string{
		"three": "three",
	}))
	assert.Equal("/one/two/three/four", r.Generate("Four", map[string]string{
		"three": "three",
		"four":  "four",
	}))
	assert.Equal("/one/two/three/four/five", r.Generate("Five", map[string]string{
		"three": "three",
		"four":  "four",
		"five":  "five",
	}))
	assert.Equal("/one/two/three/four/five/six", r.Generate("Six", map[string]string{
		"three": "three",
		"four":  "four",
		"five":  "five",
		"six":   "six",
	}))
	r.Shutdown(context.Background())
}

func TestStaticDirRoute(t *testing.T) {
	assert := assert.New(t)
	r := New()
	r.Get("/{*:file}", http.StripPrefix("/test/", http.FileServer(http.Dir("test"))))
	r.Get("/static/{*:file}", http.StripPrefix("/test/", http.FileServer(http.Dir("test"))))
	r.Get("/prefix/{*:file}", http.FileServer(http.Dir("test")))
	go func() {
		assert.NoError(r.Run())
	}()
	sendTestRequests(assert, []testRequest{
		{
			Path: "http://localhost:8080/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/static/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/static/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/file.txt",
			Body: "This is file.txt.",
		},
		{
			Path: "http://localhost:8080/prefix/test/directory/file.txt",
			Body: "This is directory/file.txt.",
		},
		{
			Path:       "http://localhost:8080/",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/static",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/prefix",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/prefix/test",
			StatusCode: statusNotFound,
			Body:       "",
		},
		{
			Path:       "http://localhost:8080/wow.txt",
			StatusCode: statusNotFound,
			Body:       "",
		},
	})
	r.Shutdown(context.Background())
}
