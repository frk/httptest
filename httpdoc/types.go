package httpdoc

import (
	"html/template"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
)

type TopicGroup struct {
	// The name of the group, optional.
	Name string
	// The list of topics that belong to the group.
	Topics []*Topic
}

// - the resulting documentation will have paths for each direct Topic in each TopicGroup
// []TopicGroup{{Topics: []Topic{{}, ... }}, ... }
//
// - SubTopics will be linked to using fragments <- need to know the parent Topic's href
// - TestGroups will be linked to using fragments <- need to know the parent Topic's href
//
// - each *httpdoc.Topic & *httptest.TestGroup represent an article in the documentation
// that needs to be referenced accurately
//
// - root topics, i.e. those that are the direct children of an *httpdoc.TopicGroup,
// must will have an id but that will not be referenced

type Topic struct {
	// The name of the topic. Used to generate the nav-item link-text in the
	// sidebar and the heading of the corresponding section in the main column.
	Name string
	// A list of sub-topics that belong to this topic.
	SubTopics []*Topic
	// A list of test groups that will be used to generate endpoint
	// specific documentation related to the topic.
	TestGroups []*httptest.TestGroup
	// Can be set to a value that will be used as the source for the Topic's
	// main text. It is up to the user to ensure that the content of the source
	// is valid and safe HTML. The following types are supported:
	//
	//	- template.HTML: valid html text.
	//	- string: raw string expected to contain valid html text.
	//	- *os.File: a file expected to contain valid html text.
	//	- httpdoc.HTML: a value of type T that satisfies the httpdoc.HTML
	//		interface and returns valid html text.
	//
	// Any other type will result in an error.
	Text       interface{}
	Attributes interface{}
	Parameters interface{}
	Returns    interface{}

	// A reference to the associated SidebarNavItem.
	navItem *page.SidebarNavItem
	// A reference to the associated ContentSection.
	contentSection *page.ContentSection
}

type HTML interface {
	HTML() (template.HTML, error)
}

type Example struct {
	Request  *Request
	Response *Response
}

type Request struct {
	// ...
}

type Response struct {
	// ...
}

////////////////////////////////////////////////////////////////////////////////
// Table
////////////////////////////////////////////////////////////////////////////////

type HeaderStyle uint

const (
	HeaderHorizontal HeaderStyle = iota
	HeaderVertical
	HeaderNone
)

type Table struct {
	Title  string
	Rows   [][]TableCell
	Header HeaderStyle
}

type TableCell struct {
	Text interface{}
}
