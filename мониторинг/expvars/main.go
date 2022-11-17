package main

import (
	"expvar"
	"fmt"
	"net/http"
	"runtime"
)

var (
	hits = expvar.NewMap("hist")
)

func handler(w http.ResponseWriter, r *http.Request) {
	hits.Add(r.URL.Path, 1)
	w.Write([]byte("expvar increased"))
}

func init() {
	hits.Init()
	expvar.Publish("mystat", expvar.Func(func() interface{} {
		return map[string]int{
			"test":          100500,
			"value":         42,
			"goroutine_num": runtime.NumGoroutine(),
		}
	}))
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)

}
