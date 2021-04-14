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
	Title   string
	RootURL template.URL
	LogoURL template.URL
}

type SidebarFooter struct {
	SigninURL template.URL
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

func (si *SidebarItem) AnchorClass() string {
	if len(si.SubItems) > 0 {
		return "sidebar-item expandable"
	}
	return "sidebar-item"
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
	Search interface{} // TODO
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

type ArticleText struct {
	Title string
	Text  template.HTML
}

type ArticleAuthInfo struct {
	Title string
	Text  template.HTML
}

type ArticleFieldList struct {
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

type ExampleEndpoints struct {
	Title     string
	Endpoints []*EndpointItem
}

type ExampleText struct {
	Title string
	Text  template.HTML
}

type ExampleObject struct {
	Title string
	Lang  string
	Text  template.HTML
}

type ExampleRequest struct {
	Title    string
	Method   string
	Pattern  string
	Snippets []CodeSnippet
}

type ExampleResponse struct {
	Title  string
	Status int           // The HTTP response status
	Header []HeaderItem  // The response's header
	Body   template.HTML // The response's body
	Lang   string        // the body's language type
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

type CodeSnippetHTTP struct {
	CodeSnippetRequest
}

type CodeSnippetCURL struct {
	CodeSnippetRequest
}

////////////////////////////////////////////////////////////////////////////////
// Fields
////////////////////////////////////////////////////////////////////////////////

type FieldListClass uint

const (
	_ FieldListClass = iota
	FIELD_LIST_HEADER
	FIELD_LIST_PATH
	FIELD_LIST_QUERY
	FIELD_LIST_BODY
	FIELD_LIST_OBJECT
)

var fieldListClasses = [...]string{
	FIELD_LIST_HEADER: "header-field-list",
	FIELD_LIST_PATH:   "path-field-list",
	FIELD_LIST_QUERY:  "query-field-list",
	FIELD_LIST_BODY:   "body-field-list",
	FIELD_LIST_OBJECT: "object-field-list",
}

func (c FieldListClass) String() string {
	return fieldListClasses[c]
}

type FieldList struct {
	Title string
	Class FieldListClass
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
	Text template.HTML
	// A link to the source of the field.
	SourceLink *SourceLink
	// SettingLabel and SettingText are used to indicates whether the field
	// is required, optional, or something else. The SettingLabel is used as
	// part of the associated CSS class name. The SettingText is used as the
	// text to be rendered in the documentation.
	SettingLabel, SettingText string
	// The field's validation documentation.
	Validation template.HTML
	// A list of enum values associated with the field's type.
	EnumList *EnumList
	// If the field's type is a struct then SubFields will hold the fields
	// of that struct. If the field's type is not a struct then SubFields
	// will be nil.
	SubFields []*FieldItem
}

////////////////////////////////////////////////////////////////////////////////
// Enum List
////////////////////////////////////////////////////////////////////////////////

type EnumList struct {
	// The title to be use for the enum list.
	Title string
	Items []*EnumItem
}

type EnumItem struct {
	// The enum's value.
	Value string
	// The enum's documentation text.
	Text template.HTML
	// A link to the source of the enum's declaration.
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

func (*ArticleText) isArticleSection()      {}
func (*ArticleAuthInfo) isArticleSection()  {}
func (*ArticleFieldList) isArticleSection() {}

func (*ExampleText) isExampleSection()      {}
func (*ExampleEndpoints) isExampleSection() {}
func (*ExampleObject) isExampleSection()    {}
func (*ExampleRequest) isExampleSection()   {}
func (*ExampleResponse) isExampleSection()  {}

func (*CodeSnippetHTTP) isCodeSnippet() {}
func (*CodeSnippetCURL) isCodeSnippet() {}
