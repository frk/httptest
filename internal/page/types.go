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
// ArticleElement
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
	Sections []*ArticleSection
	//
	Example ExampleElement
	// A list of sub articles.
	SubArticles []*ArticleElement
}

type ArticleSection struct {
	// The title of the section.
	Title string
	// The following fields represent the content of the section.
	// NOTE: Only one of these fields should be set.
	Text       template.HTML
	AuthInfo   template.HTML
	FieldLists []*FieldList
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
// ExampleElement
////////////////////////////////////////////////////////////////////////////////

type ExampleElement struct {
	Sections []*ExampleSection
}

type ExampleSection struct {
	// The title of the section.
	Title string

	// The following fields represent the content of the section.
	// NOTE: Only one of these fields should be set.
	Text             template.HTML
	ExampleObject    *ExampleObject
	ExampleRequest   *ExampleRequest
	ExampleResponse  *ExampleResponse
	EndpointOverview *EndpointOverview
}

type ExampleObject struct {
	Title string
	Type  string
	Code  template.HTML
}

type ExampleRequest struct {
	Method   string
	Pattern  string
	Snippets []*ExampleSnippet
}

type ExampleSnippet struct {
	Lang string
	Code template.HTML
}

type ExampleResponse struct {
	// The title of the example.
	Title string
	// The HTTP response status
	Status int
	// The response's body
	Header []template.HTML
	// The response's body
	Code template.HTML
}

////////////////////////////////////////////////////////////////////////////////
// Endpoint Overview
////////////////////////////////////////////////////////////////////////////////

type EndpointOverview struct {
	Title string
	Items []*EndpointItem
}

type EndpointItem struct {
	Href    string
	Method  string
	Pattern string
	Tooltip string
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
// Etc.
////////////////////////////////////////////////////////////////////////////////

type SourceLink struct {
	// The link to the source code.
	Href string
	// The text inside the anchor.
	Text string
}
