package program

import (
	"html/template"
	"strings"
)

var T = template.Must(template.New("t").Parse(strings.Join([]string{
	prog_file,
	func_new,
	type_handler,
	index_handler,
	must_get_files_dir,
}, "")))

var prog_file = `package {{ .PkgName }}

import (
	"html/template"
	"log"
	"net/http"
	{{- if .IsExecutable }}
	"os"
	{{- end }}
	"path/filepath"
)

var (
	// directories
	filesdir = mustGetFilesDir()
	htmldir  = filepath.Join(filesdir, "html")
	cssdir   = filepath.Join(filesdir, "css")
	jsdir    = filepath.Join(filesdir, "js")

	// file servers
	cssfs = http.FileServer(http.Dir(cssdir))
	jsfs  = http.FileServer(http.Dir(jsdir))
)

{{- if .IsExecutable }}

func main() {
	http.ListenAndServe("{{ .ListenAddr }}", New())
}
{{- end }}

{{ template "func_new" . }}

{{ template "type_handler" . }}

{{ template "index_handler" . }}

{{ template "must_get_files_dir" . }}
` //`

var func_new = `{{ define "func_new" -}}
func New() http.Handler {
	mux := http.NewServeMux()

	// initialize handlers
	{{ .IndexHandler.Name }} := indexHandler("{{ .IndexHandler.File }}")
	{{- range .Handlers }}
	{{ .Name }} := newHandler("{{ .File }}")
	{{- end }}

	// register handlers
	mux.Handle("/", {{ .IndexHandler.Name }})
	mux.Handle("{{ .IndexHandler.Path }}", {{ .IndexHandler.Name }})
	{{- range .Handlers }}
	mux.Handle("{{ .Path }}", {{ .Name }})
	{{- end }}

	// register file servers
	mux.Handle("/assets/css/", http.StripPrefix("/assets/css/", cssfs))
	mux.Handle("/assets/js/", http.StripPrefix("/assets/js/", jsfs))

	return mux
}
{{ end -}}
` //`

var type_handler = `{{ define "type_handler" -}}
type handler struct {
	t *template.Template
}

func newHandler(filename string) *handler {
	t, err := template.ParseFiles(filepath.Join(htmldir, filename))
	if err != nil {
		panic(err)
	}
	return &handler{t: t}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.t.Execute(w, nil); err != nil {
		log.Println(err)
	}
}
{{ end -}}
` //`

var index_handler = `{{ define "index_handler" -}}
func indexHandler(filename string) http.Handler {
	t, err := template.ParseFiles(filepath.Join(htmldir, filename))
	if err != nil {
		panic(err)
	}
	h := &handler{t: t}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "{{ .RootPath }}" {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

			http.Redirect(w, r, "{{ .RootPath }}", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}
{{ end -}}
` //`

var must_get_files_dir = `{{ define "must_get_files_dir" -}}
{{ if .IsExecutable -}}
func mustGetFilesDir() string {
	x, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(x), "files")
}
{{ else -}}
func mustGetFilesDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "files")
}
{{ end -}}
{{ end -}}
` //`
