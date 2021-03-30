package page

import (
	"html/template"
)

type Page struct {
	Title   string
	Sidebar Sidebar
	Content Content
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

type Sidebar struct {
	Header    SidebarHeader
	NavGroups []*SidebarNavGroup
	Footer    SidebarFooter
}

type SidebarHeader struct {
	Logo   string
	Title  string
	Search interface{} // TODO
}

type SidebarFooter struct {
	// ...
}

type SidebarNavGroup struct {
	Heading string
	Items   []*SidebarNavItem
}

type SidebarNavItem struct {
	Text     string
	Href     string
	SubItems []*SidebarNavItem
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

type Content struct {
	Header   ContentHeader
	Sections []*ContentSection
	Footer   ContentFooter
}

type ContentHeader struct {
	// ...
}

type ContentFooter struct {
	// ...
}

type ContentSection struct {
	Id          string
	Article     Article
	Example     Example
	SubSections []*ContentSection
}

////////////////////////////////////////////////////////////////////////////////
// Article
////////////////////////////////////////////////////////////////////////////////

type Article struct {
	Heading    string
	Href       string
	Text       template.HTML
	FieldLists []*FieldList
}

type FieldList struct {
	Title string
	Items []*FieldListItem
}

type FieldListItem struct {
	// The name of the field.
	Name string
	// The path to the field's name if nested, otherwise empty.
	Path string
	// The name of the field's type.
	Type string
	Href string
	// The field's documentation.
	Text      template.HTML
	SubFields []*FieldListItem
}

////////////////////////////////////////////////////////////////////////////////
// Example
////////////////////////////////////////////////////////////////////////////////

type Example struct {
	EndpointOverview []EndpointOverviewItem
	Code             []ExampleCode
}

type EndpointOverviewItem struct {
	Href    string
	Method  string
	Pattern string
	Tooltip string
}

type ExampleCode struct {
	Request  *ExampleRequest
	Response *ExampleResponse
}

type ExampleRequest struct {
	Method string
	Header map[string][]string
}

type ExampleResponse struct {
	StatusCode int
	Header     map[string][]string
	Body       string
}
