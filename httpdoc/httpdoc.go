package httpdoc

import (
	"html/template"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/tagutil"
)

const (
	DefaultOutputName    = "apidoc"
	DefaultPageTitle     = "Example API v1"
	DefaultRootPath      = "/api"
	DefaultSigninPath    = "/signin"
	DefaultTCPListenPort = 8080
	DefaultExampleHost   = "https://api.example.com"
	DefaultFieldNameTag  = "json"
)

var (
	DefaultSnippetTypes = []SnippetType{SNIPP_HTTP}
)

type Config struct {
	////////////////////////////////////////////////////////////////////////
	// Output specific configuration
	////////////////////////////////////////////////////////////////////////

	// The name of the resulting package or executable.
	//
	// If empty, it will default to DefaultOutputName.
	OutputName string
	// The directory into which the resulting package or executable should
	// be written.
	//
	// If empty, the result will be written to the directory in which
	// the client program, the one calling httpdoc.Compile, lives.
	OutputDir string
	// If set to true, httpdoc will generate the source code of the
	// program as an importable package.
	//
	// When false, httpdoc will generate and compile the program to
	// produce an executable.
	OutputPackage bool

	////////////////////////////////////////////////////////////////////////
	// Page specific configuration
	////////////////////////////////////////////////////////////////////////

	// The title for the generated web page.
	//
	// If empty, it will default to DefaultPageTitle.
	PageTitle string

	// The HTML content for the page's sidebar banner. It is the user's
	// responsibility to ensure that the enclosed content is valid and
	// safe HTML as it will be included verbatim in the template output.
	//
	// If empty, the value of the PageTitle field will be used in the
	// sidebar's banner as a heading.
	SidebarBannerHTML template.HTML

	////////////////////////////////////////////////////////////////////////
	// Server specific configuration
	////////////////////////////////////////////////////////////////////////

	// The root of the path for the documentation and the links within it.
	// If left empty it will default to DefaultRootPath.
	RootPath string
	// The path to the documentation server's signin endpoint.
	// If left empty it will default to DefaultSigninPath.
	SigninPath string
	// The TCP port which the resulting program should listen on.
	// If left unset, it will default to DefaultTCPListenPort.
	TCPListenPort int

	// An optional map of users and their passwords.
	Users map[string]string

	////////////////////////////////////////////////////////////////////////
	//
	////////////////////////////////////////////////////////////////////////

	// The host which will be used in example snippets. If no host
	// is provided it will default to the value of DefaultExampleHost.
	ExampleHost string
	// TODO
	StripPrefix func(pattern string) string
	// The tag to be used to resolve a field's name for the documentation,
	// defaults to DefaultFieldNameTag. If no name is present in the tag's
	// value the field's name will be used as fallback.
	FieldNameTag string
	// FieldType returns the name for a specific field's type based on
	// given reflect.StructField value.
	//
	// If FieldType is nil or it returns false as the second return
	// value (ok) it will fall back to the default behaviour.
	FieldType func(field reflect.StructField) (typeName string, ok bool)
	// FieldSetting returns values that are used to document whether the given
	// field is required, optional, or something else. The structType argument
	// represents the type of struct to which the field belongs.
	//
	// The returned label is used in the corresponding element's class.
	// The returned text is used as the corresponding element's content.
	// The returned ok value indicates whether or not the field's setting
	// documentation should not be generated.
	//
	// If FieldSetting is nil then the documentation will be generated based
	// on the field's "doc" tag and if the field doesn't have a "doc" tag then
	// the field will be documented as optional.
	FieldSetting func(field reflect.StructField, structType reflect.Type) (label, text string, ok bool)
	// FieldValidation returns the documentation on the given field's validity
	// requirements. The structType argument represents the type of struct
	// to which the field belongs. If the returned text is empty then no
	// validation-specific documentation will be rendered.
	//
	// If FieldValidation is nil then no documentation on the field validity
	// requirements will be generated.
	FieldValidation func(field reflect.StructField, structType reflect.Type) (text template.HTML)
	// SourceURL returns the URL of the source code location corresponding
	// to the given filename and line. This can be used to add source links
	// to the documentation for certain objects like handlers, structs,
	// fields, and enums.
	//
	// If SourceURL is nil then no source links will be generated.
	SourceURL func(filename string, line int) (url string)
	// A list of basic SnippetTypes for which request-specific code examples
	// should be generated. If left empty it will default to DefaultSnippetTypes.
	SnippetTypes []SnippetType

	// The directory of the httpdoc package.
	pkgdir string
	// The directory of the client program (the one calling Compile).
	srcdir string

	// the following are hardcoded, may later be made configurable

	// this setting needs some more thought on how to best ensure that the logo
	// image is rendered well, like should a specific size limit be enforece?
	// specific image type? something else? ...
	logoURL string
	// XXX: just a placeholder, not yet implemented and may never be.
	useEmbed bool
}

func Compile(c Config, dir ArticleDirectory) error {
	// TODO validate provided values that need to be validated
	// e.g. OutputDir should be a dir that can be written to

	// get httpdoc's dir
	_, f, _, _ := runtime.Caller(0)
	c.pkgdir = filepath.Dir(f)

	// get the caller's dir & then normalize
	_, f, _, _ = runtime.Caller(1)
	c.srcdir = filepath.Dir(f)

	c.normalize()
	b := &build{Config: c, dir: dir}
	if err := b.run(); err != nil {
		return err
	}
	w := &write{Config: c, page: b.page, prog: b.prog}
	if err := w.run(); err != nil {
		return err
	}
	return nil
}

func (c *Config) normalize() {
	c.logoURL = "/logo.png"

	if len(c.OutputName) == 0 {
		c.OutputName = DefaultOutputName
	}
	if len(c.OutputDir) == 0 {
		c.OutputDir = c.srcdir
	}

	if len(c.PageTitle) == 0 {
		c.PageTitle = DefaultPageTitle
	}
	if len(c.RootPath) > 0 {
		// RootPath should not end with a slash if it's longer than one char
		c.RootPath = strings.TrimRight(c.RootPath, "/")
		if len(c.RootPath) == 0 || c.RootPath[0] != '/' {
			c.RootPath = "/" + c.RootPath
		}
	} else {
		c.RootPath = DefaultRootPath
	}
	if len(c.SigninPath) == 0 {
		c.SigninPath = DefaultSigninPath
	}
	if c.TCPListenPort < 1 {
		c.TCPListenPort = DefaultTCPListenPort
	}

	if len(c.ExampleHost) == 0 {
		c.ExampleHost = DefaultExampleHost
	}
	if c.StripPrefix == nil {
		c.StripPrefix = func(s string) string { return s }
	}
	if len(c.FieldNameTag) == 0 {
		c.FieldNameTag = DefaultFieldNameTag
	}
	if c.FieldSetting == nil {
		c.FieldSetting = DefaultFieldSetting
	}
	if len(c.SnippetTypes) == 0 {
		c.SnippetTypes = DefaultSnippetTypes
	}
}

// SourceURLFunc returns a function that can be used in the Config's SourceURL
// field. The arguments local and remote will be used together with the arguments
// of the returned function to generate the source URLs.
//
// The argument local should be the local (i.e. on the host machine) root directory
// of the project for which the documentation is being generated.
//
// The argument remote should be the remote, web-accessible, root location
// of the project for which the documentation is being generated.
// For example:
//	// for a github repo the remote should have the following format.
//	remote = "https://github.com/<user>/<project>/tree/<branch>/"
//	// for a bitbucket repo the remote should have the following format.
//	remote = "https://bitbucket.org/<user>/<project>/src/<branch>/"
func SourceURLFunc(local, remote string) (f func(filename string, line int) (url string)) {
	// the code that constructs the source links expects these to *not* end with "/"
	if l := len(remote); l > 0 && remote[l-1] == '/' {
		remote = remote[:l-1]
	}
	if l := len(local); l > 0 && local[l-1] == '/' {
		local = local[:l-1]
	}

	lp := "#" // line prefix
	if strings.HasPrefix(remote, "https://github.com") {
		lp = "#L"
	} else if strings.HasPrefix(remote, "https://bitbucket.org") {
		lp = "#lines-"
	} else {
		// TODO(mkopriva): handle other popular remote repositories
	}

	return func(filename string, line int) (url string) {
		file := strings.TrimPrefix(filename, local)
		href := remote + file + lp + strconv.Itoa(line)
		return href
	}
}

func DefaultFieldSetting(s reflect.StructField, t reflect.Type) (label, text string, ok bool) {
	const required, optional = "required", "optional"

	tag := tagutil.New(string(s.Tag))
	if tag.Contains("doc", required) {
		return required, required, true
	}
	return optional, optional, true
}

// Value is a value that will be used by httpdoc as the source of documentation.
// Depending on the context in which it is used, different aspects of the value
// will be sourced to generate different kinds of documentation.
type Value interface{}

// The Valuer type returns a Value value.
type Valuer interface {
	Value() (Value, error)
}

// HTML represents a known safe HTML document fragment that will be used
// by httpdoc as the source of documentation.
//
// To understand the security risks involved with the use of this type, please
// read the documentation on the html/template.HTML type to which this type is
// converted verbatim by the httpdoc package.
type HTML string

// The HTMLer type returns an HTML value.
type HTMLer interface {
	HTML() (HTML, error)
}

// Article is the primary data structure of the ArticleDirectory type. It is used
// by httpdoc to generate <article> elements that contain documentation extracted
// from the Article's data.
type Article struct {
	// The title of the article.
	Title string
	// A list of test groups used to generate child <article> elements
	// that contain endpoint-specific documentation.
	TestGroups []*httptest.TestGroup
	// A list of articles used to generate child <article> elements
	// that contain further documentation.
	SubArticles []*Article
	// The Text field, if set, will be used as the source for the article's
	// primary-column content. The following types are accepted:
	//	- string
	//	- *os.File
	//	- httpdoc.HTMLer
	//	- httpdoc.Valuer
	//	- interface{} (named types only)
	// Anything else will result in an error.
	//
	// If the type is string, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is *os.File, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is httpdoc.HTMLer, then its HTML() method will be used
	// to retrieve the content and it is up to the user to ensure that that
	// content is valid and safe HTML.
	//
	// If the type is httpdoc.Valuer, then its Value() method will be used
	// to get the underlying Value, the source of that Value's dynamic type
	// is then analyzed and any relevant documentation that's found will be
	// used to generate the HTML text. If the dynamic type is unnamed an
	// error will be returned instead.
	//
	// If the type is none of the above, then the type's source is analyzed
	// and any relevant documentation that's found will be used to generate
	// the HTML text. If the type is unnamed an error will be returned.
	Text interface{}
	// The Code field, if set, will be used as the source for the article's
	// example-column content. The following types are accepted:
	//	- string
	//	- *os.File
	//	- httpdoc.HTMLer
	//	- httpdoc.Valuer
	//	- interface{} (named types only)
	// Anything else will result in an error.
	//
	// If the type is string, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is *os.File, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is httpdoc.HTMLer, then its HTML() method will be used
	// to retrieve the content and it is up to the user to ensure that that
	// content is valid and safe HTML.
	//
	// If the type is httpdoc.Valuer, then its Value() method will be used to
	// get the underlying Value, that Value will then be marshaled according to
	// its MIME type (which resolved based on the Article's Type field) and the
	// result of that will be used to generate a code snippet.
	//
	// If the type is none of the above, then the value will then be marshaled
	// according to its MIME type (which resolved based on the Article's Type field)
	// and the result of that will be used to generate a code snippet.
	//
	// Note that in the last two cases, if the value's dynamic type is a struct,
	// or its base element type is a struct, then that struct's source code will
	// be analyzed to generate the documentation of the individual fields for
	// the article's primary-column.
	Code interface{}
	// The Type field can optionally be set to the MIME type that should be used
	// to present the data in the Code field. If left unset, and the Code field's
	// dynamic type implements the httptest.Body interface then its Type() method
	// will be used to resolve the MIME type. Otherwise the MIME type will default
	// to "application/json".
	Type string
	// LoadExpanded controls how much of the article's content will be loaded
	// during the initial page load.
	//
	// If set, the entire article will be loaded.
	//
	// If unset, only the "preamble" of the article will be loaded and
	// subsequent user interaction will be required to trigger the loading
	// of the rest of the content.
	//
	// NOTE: "root" articles only, this field is ignored for SubArticles.
	LoadExpanded bool
}

// ArticleGroup is a list of loosely related articles. It is used by httpdoc to generate
// a named or unnamed group of sidebar items that will point to the individual articles.
type ArticleGroup struct {
	// The name of the group, optional.
	Name string
	// The list of articles that belong to the group.
	Articles []*Article
	// LoadExpanded controls how much of the group articles' content will be
	// loaded during the initial page load.
	//
	// If set, the group's entire articles will be loaded.
	//
	// If unset, only the "preambles" of the group's articles will be loaded
	// and subsequent user interaction will be required to trigger the loading
	// of the rest of the content.
	LoadExpanded bool
}

// ArticleDirectory is the hierarchy of articles used to generate the documentation.
type ArticleDirectory []*ArticleGroup

type SnippetType uint8

const (
	SNIPP_HTTP SnippetType = iota
	SNIPP_CURL
	// TODO add support for more snippet types: js, go, etc..
)

func (st SnippetType) Name() string { return snippetTypeName[st] }
func (st SnippetType) Lang() string { return snippetTypeLang[st] }

var snippetTypeName = [...]string{
	SNIPP_HTTP: "HTTP",
	SNIPP_CURL: "cURL",
}

var snippetTypeLang = [...]string{
	SNIPP_HTTP: "http",
	SNIPP_CURL: "curl",
}
