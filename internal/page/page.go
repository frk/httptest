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
	// NOTE(mkopriva): The code that's writing the Page into files relies on
	// the Content field to be a non-pointer. If it later needs to be changed
	// to a pointer, then the page-writing code needs to be updated to create
	// a shallow copy of the pointed-to Content.
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
	Banner SidebarBanner
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

type SidebarBanner interface {
	isSidbarBanner()
}

type SidebarBannerTitle struct {
	Text string
	URL  template.URL
}

type SidebarBannerHTML struct {
	Text template.HTML
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
	Code  template.HTML
}

type ExampleRequest struct {
	Title    string
	Method   string
	Pattern  string
	Options  []*SelectOption
	Snippets []*CodeSnippetElement
}

type ExampleResponse struct {
	Title  string
	Status int           // The HTTP response status
	Header []HeaderItem  // The response's header
	Lang   string        // the body's language type
	Code   template.HTML // The response's body
}

////////////////////////////////////////////////////////////////////////////////
// Code Snippet
////////////////////////////////////////////////////////////////////////////////

type CodeSnippetElement struct {
	Id       string // the id of the snippet's primary element
	Show     bool   // indicates whether or not to show this snippet
	Lang     string // the language of the snippet's code
	NumLines int    // the number of lines needed to render the snippet
	Snippet  CodeSnippet
}

func (e *CodeSnippetElement) Lines() []struct{} {
	return make([]struct{}, e.NumLines)
}

type CodeSnippet interface {
	isCodeSnippet()
}

// CodeSnippetHTTP represents a raw HTTP request message.
type CodeSnippetHTTP struct {
	// The start line
	Method, RequestURI, HTTPVersion string
	// The header fields
	Headers []HeaderItem
	// The message body
	Body template.HTML
}

func (CodeSnippetHTTP) Name() string { return "HTTP" }
func (CodeSnippetHTTP) Lang() string { return "http" }

type CodeSnippetCURL struct {
	// The target URL
	URL string
	// The -X/--request option
	X string
	// The -H/--header options
	H []string
	// the -d/--data options
	Data []CURLDataType
}

func (CodeSnippetCURL) Name() string { return "cURL" }
func (CodeSnippetCURL) Lang() string { return "curl" }

func (cs *CodeSnippetCURL) NumOpts() int { return len(cs.H) + len(cs.Data) }

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

type SelectOption struct {
	Text     string
	Value    string
	DataId   string
	Selected bool
}

////////////////////////////////////////////////////////////////////////////////
// cURL specific types
////////////////////////////////////////////////////////////////////////////////
type CURLDataType interface {
	isCURLDataType()
}

type CURLDataText string

func (s CURLDataText) HTML() template.HTML { return template.HTML(s) }

type CURLDataKeyValue struct {
	Key   string
	Value string
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (*SidebarBannerTitle) isSidbarBanner() {}
func (*SidebarBannerHTML) isSidbarBanner()  {}

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

func (CURLDataText) isCURLDataType()     {}
func (CURLDataKeyValue) isCURLDataType() {}
