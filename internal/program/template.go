package program

import (
	"strings"
	"text/template"
)

var T = template.Must(template.New("t").Parse(strings.Join([]string{
	prog_file,
	func_new,

	type_page,
	type_user,

	signout_handler,
	signin_handler,
	docs_handler,

	user_middleware,
	user_database,
	session_database,

	valid_paths,
	snippet_langs,
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

{{- if .Users }}
{{ template "type_user" . }}

{{ template "signout_handler" . }}

{{ template "signin_handler" . }}
{{- end }}

{{ template "docs_handler" . }}

{{- if .Users }}
{{ template "user_middleware" . }}

{{ template "user_database" . }}

{{ template "session_database" . }}
{{- end }}

{{ template "valid_paths" . }}

{{ template "snippet_langs" . }}

{{ template "must_get_files_dir" . }}
` //`

var func_new = `{{ define "func_new" -}}
func New() http.Handler {
	mux := http.NewServeMux()

	// initialize handlers
	docsHandler := newDocsHandler("docs.html")
	{{- if .Users }}
	signinHandler := newSigninHandler("signin.html")

	// add middleware
	docsHandler = userMiddleware(docsHandler)
	{{- end }}

	// register handlers
	mux.Handle("/", docsHandler)
	{{- if .Users }}
	mux.Handle("/signin", signinHandler)
	mux.Handle("/signout", http.HandlerFunc(signoutHandler))
	{{- end }}

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
	Lang string
	{{- if .Users }}
	User *user
	{{- end }}
}

func (p page) IsHidden(s string) bool {
	return !strings.HasPrefix(p.Path, s)
}

func (p page) IsActive(s string) bool {
	return p.Path == s
}

func (p page) IsLang(s string) bool {
	return p.Lang == s
}
{{ end -}}
` //`

var type_user = `{{ define "type_user" -}}
type user struct {
	Name string
	hash string // encrypted password
}
{{ end -}}
` //`

var signout_handler = `{{ define "signout_handler" -}}
func signoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if ck, err := r.Cookie("{{ .SessionName }}"); err == nil {
		session.del(ck.Value)

		http.SetCookie(w, &http.Cookie{
			Name:     "{{ .SessionName }}",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(1, 0),
			MaxAge:   -1,
		})
	}

	http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
}
{{ end -}}
` //`

var signin_handler = `{{ define "signin_handler" -}}
func newSigninHandler(filename string) http.Handler {
	t, err := template.ParseFiles(filepath.Join(htmldir, filename))
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		case "GET":
			// check if already signed in, if so then send the user to root page
			if ck, err := r.Cookie("{{ .SessionName }}"); err == nil {
				if u := session.get(ck.Value); u != nil {
					http.Redirect(w, r, "{{ .RootPath }}", http.StatusTemporaryRedirect)
					return
				}
			}

			if err := t.Execute(w, nil); err != nil {
				log.Println(err)
			}
		case "POST":
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}

			name := r.PostForm.Get("name")
			pass := r.PostForm.Get("password")
			u := users.get(name)
			if u == nil {
				http.NotFound(w, r)
				return
			}

			err := bcrypt.CompareHashAndPassword([]byte(u.hash), []byte(pass))
			if err != nil {
				if err != bcrypt.ErrMismatchedHashAndPassword && err != bcrypt.ErrHashTooShort {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Error(w, "invalid password", http.StatusUnprocessableEntity)
				return
			}

			b := make([]byte, 64/4*3)
			if _, err := rand.Read(b); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sid := base64.URLEncoding.EncodeToString(b)
			session.set(sid, u)

			http.SetCookie(w, &http.Cookie{
				Name:     "{{ .SessionName }}",
				Value:    sid,
				Path:     "/",
				Expires:  time.Now().Add(time.Hour * 24 * 365),
				Secure:   true,
				HttpOnly: true,
			})
			http.Redirect(w, r, "{{ .RootPath }}", http.StatusFound)
		}
	})
}
{{ end -}}
` //`

var docs_handler = `{{ define "docs_handler" -}}
func newDocsHandler(filename string) http.Handler {
	t, err := template.ParseFiles(filepath.Join(htmldir, filename))
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		p := page{}
		p.Id = r.URL.Fragment
		p.Path = r.URL.Path
		p.Lang = getSnippetLang(r)
		{{- if .Users }}
		p.User = getUser(r)
		{{- end }}

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

		if err := t.Execute(w, p); err != nil {
			log.Println(err)
		}
	})
}
{{ end -}}
` //`

var user_middleware = `{{ define "user_middleware" -}}
func userMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ck, err := r.Cookie("{{ .SessionName }}")
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		}

		u := session.get(ck.Value)
		if u == nil {
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		}
		
		ctx := context.WithValue(r.Context(), "user", u)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

func getUser(r *http.Request) *user {
	u, _ := r.Context().Value("user").(*user)
	return u
}
{{ end -}}
` //`

var user_database = `{{ define "user_database" -}}
var users = &userDatabase{
	users: map[string]*user{
	{{- range $name, $hash := .Users }}
		"{{ $name }}": &user{Name: "{{ $name }}", hash: "{{ $hash }}"},
	{{- end }}
	},
}

// readonly
type userDatabase struct {
	users map[string]*user
}

func (db *userDatabase) get(name string) *user {
	return db.users[name]
}
{{ end -}}
` //`

var session_database = `{{ define "session_database" -}}
var session = &sessionDatabase{
	users: make(map[string]*user),
}

type sessionDatabase struct {
	mu    sync.Mutex
	users map[string]*user
}

func (db *sessionDatabase) get(sid string) *user {
	db.mu.Lock()
	u := db.users[sid]
	db.mu.Unlock()
	return u
}

func (db *sessionDatabase) set(sid string, u *user) {
	db.mu.Lock()
	db.users[sid] = u
	db.mu.Unlock()
}

func (db *sessionDatabase) del(sid string) {
	db.mu.Lock()
	delete(db.users, sid)
	db.mu.Unlock()
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

var snippet_langs = `{{ define "snippet_langs" -}}
var snippetLangs = []string{
	{{- range .SnippetTypes }}
	"{{ . }}",
	{{- end }}
}

func getSnippetLang(r *http.Request) string {
	lang := r.URL.Query().Get("lang")
	for i := 0; i < len(snippetLangs); i++ {
		if snippetLangs[i] == lang {
			return lang
		}
	}
	return snippetLangs[0] // default
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
	_ = x
	return filepath.Join(filepath.Dir(x), "files")
	//return "/app/files"
}
{{ else -}}
func mustGetFilesDir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "files")
}
{{ end -}}
{{ end -}}
` //`
