package httpdoc

import (
	"html/template"
	"reflect"
	"strconv"
	"strings"

	"github.com/frk/tagutil"
)

const (
	DefaultPageTitle    = "Example API v1"
	DefaultRootPath     = "/api"
	DefaultSigninPath   = "/signin"
	DefaultExampleHost  = "https://api.example.com"
	DefaultFieldNameTag = "json"
)

var (
	DefaultSnippetTypes = []SnippetType{SNIPP_HTTP}
)

type Config struct {
	// The title for the generated web page.
	// If left empty it will default to DefaultPageTitle.
	PageTitle string
	// The root of the path for the documentation and the links within it.
	// If left empty it will default to DefaultRootPath.
	RootPath string
	// The path to the documentation server's signin endpoint.
	// If left empty it will default to DefaultSigninPath.
	SigninPath string
	// The host which will be used in example snippets. If no host
	// is provided it will default to the value of DefaultExampleHost.
	ExampleHost string
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

	// the following are hardcoded, may later be made configurable

	// this setting needs some more thought on how to best ensure that the logo
	// image is rendered well, like should a specific size limit be enforece?
	// specific image type? something else? ...
	logoURL string
}

func (c *Config) Build(dir ArticleDirectory) error {
	c.normalize()

	b := &build{Config: *c, dir: dir}
	if err := b.loadCallerSource(1); err != nil {
		return err
	}
	return b.run()
}

func (c *Config) normalize() {
	c.logoURL = "/logo.png"

	// RootPath should not end with a slash if it's longer than one char
	if len(c.RootPath) > 0 {
		c.RootPath = strings.TrimRight(c.RootPath, "/")
		if len(c.RootPath) == 0 || c.RootPath[0] != '/' {
			c.RootPath = "/" + c.RootPath
		}
	} else {
		c.RootPath = DefaultRootPath
	}

	if len(c.PageTitle) == 0 {
		c.PageTitle = DefaultPageTitle
	}
	if len(c.SigninPath) == 0 {
		c.SigninPath = DefaultSigninPath
	}
	if len(c.ExampleHost) == 0 {
		c.ExampleHost = DefaultExampleHost
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
