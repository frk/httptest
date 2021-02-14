package main

import (
	"net/http"
	"path/filepath"
	"runtime"
)

func main() {
	_, f, _, _ := runtime.Caller(0)
	d := filepath.Dir(filepath.Dir(f))

	files := http.FileServer(http.Dir(d + "/testdata/out"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		files.ServeHTTP(w, r)
	})
	http.ListenAndServe(":8181", nil)
}
