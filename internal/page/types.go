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
	Href       string
	Heading    string
	Text       template.HTML
	FieldLists []*FieldList
}

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
	Text template.HTML
	// A link to the source of the field.
	SourceLink string
	// SettingLabel and SettingText are used to indicates whether the field
	// is required, optional, or something else. The SettingLabel is used as
	// part of the associated CSS class name. The SettingText is used as the
	// text to be rendered in the documentation.
	SettingLabel, SettingText string
	// The field's validation documentation.
	Validation template.HTML
	// If the field's type is named and constants were declared with it
	// the EnumList will hold info about those constants. If the field's
	// type is unnamed or there are no associated constants then EnumList
	// will be nil.
	EnumList *EnumList
	// If the field's type is a struct then SubFields will hold the fields
	// of that struct. If the field's type is not a struct then SubFields
	// will be nil.
	SubFields []*FieldItem
}

type EnumList struct {
	// The title to be use for the enum list.
	Title string
	Items []*EnumItem
}

type EnumItem struct {
	// The enum value.
	Value string
	// The enum value's documentation.
	Text template.HTML
	// A link to the source of the enum value.
	SourceLink string
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
