package page

import (
	"html/template"
)

////////////////////////////////////////////////////////////////////////////////
// Page
////////////////////////////////////////////////////////////////////////////////

type Page struct {
	Title   string
	Sidebar Sidebar
	Content Content
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

type Sidebar struct {
	Header SidebarHeader
	Lists  []*SidebarList
	Footer SidebarFooter
}

type SidebarHeader struct {
	Logo   string
	Title  string
	Search interface{} // TODO
}

type SidebarFooter struct {
	// ...
}

type SidebarList struct {
	Title string
	Items []*SidebarItem
}

type SidebarItem struct {
	Text     string
	Href     string
	SubItems []*SidebarItem
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

type Content struct {
	Header   Header
	Articles []*ArticleElement
	Footer   Footer
}

type Header struct {
	// ...
}

type Footer struct {
	// ...
}

////////////////////////////////////////////////////////////////////////////////
// Article Element
////////////////////////////////////////////////////////////////////////////////

type ArticleElement struct {
	Id string
	// The article's anchor.
	Href string
	// The article's title.
	Title string
	// A link to the source code associated with the article.
	SourceLink *SourceLink
	// The article's HTML text.
	Text template.HTML
	// A list of additional sections of the article.
	Sections []ArticleSection
	// An example related to the article.
	Example ExampleElement
	// A list of sub articles.
	SubArticles []*ArticleElement
}

type ArticleSection interface {
	isArticleSection()
}

type TextArticleSection struct {
	Title string
	Text  template.HTML
}

type AuthInfoArticleSection struct {
	Title    string
	AuthInfo template.HTML
}

type FieldListArticleSection struct {
	Title string
	Lists []*FieldList
}

////////////////////////////////////////////////////////////////////////////////
// Example Element
////////////////////////////////////////////////////////////////////////////////

type ExampleElement struct {
	Sections []ExampleSection
}

type ExampleSection interface {
	isExampleSection()
}

type TextExampleSection struct {
	Title string
	Text  template.HTML
}

type EndpointsExampleSection struct {
	Title     string
	Endpoints []*EndpointItem
}

type ObjectExampleSection struct {
	Title  string
	Object template.HTML
}

type RequestExampleSection struct {
	Title    string
	Method   string
	Pattern  string
	Snippets []CodeSnippet
}

type ResponseExampleSection struct {
	Title  string
	Status int           // The HTTP response status
	Header []HeaderItem  // The response's header
	Body   template.HTML // The response's body
}

////////////////////////////////////////////////////////////////////////////////
// Fields
////////////////////////////////////////////////////////////////////////////////

type FieldList struct {
	Title string
	Items []*FieldItem
}

type FieldItem struct {
	// The unique identifier of the item.
	Id string
	// The href linking to the item.
	Href string
	// The name of the field.
	Name string
	// The path to the field's name if nested, otherwise empty.
	Path string
	// The name of the field's type.
	Type string
	// The field's documentation.
	Doc template.HTML
	// A link to the source of the field.
	SourceLink *SourceLink
	// SettingLabel and SettingText are used to indicates whether the field
	// is required, optional, or something else. The SettingLabel is used as
	// part of the associated CSS class name. The SettingText is used as the
	// text to be rendered in the documentation.
	SettingLabel, SettingText string
	// The field's validation documentation.
	Validation template.HTML
	// A list of values associated with the field.
	ValueList *ValueList
	// If the field's type is a struct then SubFields will hold the fields
	// of that struct. If the field's type is not a struct then SubFields
	// will be nil.
	SubFields []*FieldItem
}

////////////////////////////////////////////////////////////////////////////////
// Value List
////////////////////////////////////////////////////////////////////////////////

type ValueList struct {
	// The title to be use for the value list.
	Title string
	// Class is used as a CSS class prefix for the list's elements.
	Class string
	Items []*ValueItem
}

type ValueItem struct {
	// The text representation of the value.
	Text string
	// The value's documentation.
	Doc template.HTML
	// A link to the source of the value's declaration.
	SourceLink *SourceLink
}

////////////////////////////////////////////////////////////////////////////////
// Misc.
////////////////////////////////////////////////////////////////////////////////

type EndpointItem struct {
	Href    string
	Method  string
	Pattern string
	Tooltip string
}

type HeaderItem struct {
	Key   string
	Value string
}

type SourceLink struct {
	// The link to the source code.
	Href string
	// The text inside the anchor.
	Text string
}

////////////////////////////////////////////////////////////////////////////////
// Code Snippet
////////////////////////////////////////////////////////////////////////////////

type CodeSnippet interface {
	isCodeSnippet()
}

type CodeSnippetRequest struct {
	Method string
	Path   string
	Host   string
	URL    string
	Header []HeaderItem
	Body   template.HTML
}

type HTTPCodeSnippet struct {
	CodeSnippetRequest
}

type CURLCodeSnippet struct {
	CodeSnippetRequest
}

////////////////////////////////////////////////////////////////////////////////

func (*TextArticleSection) isArticleSection()      {}
func (*AuthInfoArticleSection) isArticleSection()  {}
func (*FieldListArticleSection) isArticleSection() {}

func (*TextExampleSection) isExampleSection()      {}
func (*EndpointsExampleSection) isExampleSection() {}
func (*ObjectExampleSection) isExampleSection()    {}
func (*RequestExampleSection) isExampleSection()   {}
func (*ResponseExampleSection) isExampleSection()  {}

func (*HTTPCodeSnippet) isCodeSnippet() {}
func (*CURLCodeSnippet) isCodeSnippet() {}
