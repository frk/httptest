package page

type Page struct {
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
	Heading     string
	Href        string
	Body        ContentBody
	Tables      []*InfoTable
	Examples    []*ExampleEndpoint
	SubSections []*ContentSection
}

type ContentBody struct {
	Text string
	Type []*ContentField
}

type ContentField struct {
	Name string
	Path string
	Href string
}

////////////////////////////////////////////////////////////////////////////////
// Info Table
////////////////////////////////////////////////////////////////////////////////

type InfoTable struct {
	// ...
}

////////////////////////////////////////////////////////////////////////////////
// Example
////////////////////////////////////////////////////////////////////////////////

type ExampleEndpoint struct {
	Request  ExampleRequest
	Response ExampleResponse
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
