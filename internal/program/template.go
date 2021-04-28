package program

import (
	"html/template"
	"strings"
)

var T = template.Must(template.New("t").Parse(strings.Join([]string{
	prog_file,
	func_new,
	type_page,
	type_handler,
	valid_paths,
	must_get_files_dir,
}, "")))

var prog_file = `package {{ .PkgName }}

import (
	{{- range .Imports }}
	"{{ . }}"
	{{- end }}
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

{{ template "type_page" . }}

{{ template "type_handler" . }}

{{ template "valid_paths" . }}

{{ template "must_get_files_dir" . }}
` //`

var func_new = `{{ define "func_new" -}}
func New() http.Handler {
	mux := http.NewServeMux()

	// initialize handlers
	docsHandler := newHandler("docs.html")

	// register handlers
	mux.Handle("/", docsHandler)

	// register file servers
	mux.Handle("/assets/css/", http.StripPrefix("/assets/css/", cssfs))
	mux.Handle("/assets/js/", http.StripPrefix("/assets/js/", jsfs))

	return mux
}
{{ end -}}
` //`

var type_page = `{{ define "type_page" -}}
type page struct {
	Id   string
	Path string
}

func (p page) IsHidden(s string) bool {
	return !strings.HasPrefix(p.Path, s)
}

func (p page) IsActive(s string) bool {
	return p.Path == s
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
	p := page{}
	p.Id = r.URL.Fragment
	p.Path = r.URL.Path

	if id, ok := validPaths[p.Path]; ok {
		if len(p.Id) == 0 {
			p.Id = id
		}
	} else {
		if p.Path != "{{ .RootPath }}" && p.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if p.Path != "{{ .RootPath }}" && p.Path == "/" {
			http.Redirect(w, r, "{{ .RootPath }}", http.StatusFound)
			return
		}
	}

	if err := h.t.Execute(w, p); err != nil {
		log.Println(err)
	}
}
{{ end -}}
` //`

var valid_paths = `{{ define "valid_paths" -}}
var validPaths = map[string]string{
	{{- range $k, $v := .ValidPaths }}
	"{{ $k }}": "{{ $v }}",
	{{- end }}
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
