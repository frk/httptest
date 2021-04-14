package httpdoc

import (
	"github.com/frk/httptest"
)

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
}

// ArticleGroup is a list of loosely related articles. It is used by httpdoc to generate
// a named or unnamed group of sidebar items that will point to the individual articles.
type ArticleGroup struct {
	// The name of the group, optional.
	Name string
	// The list of articles that belong to the group.
	Articles []*Article
}

// ArticleDirectory is the hierarchy of articles used to generate the documentation.
type ArticleDirectory []*ArticleGroup

type SnippetType uint8

const (
	SNIPP_HTTP SnippetType = iota
	SNIPP_CURL
	// TODO add support for more snippet types: js, go, etc..
)
