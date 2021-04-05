package httpdoc

import (
	"github.com/frk/httptest"
)

type TopicGroup struct {
	// The name of the group, optional.
	Name string
	// The list of topics that belong to the group.
	Topics []*Topic
}

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
	//	- string: raw string expected to contain valid html text.
	//	- *os.File: a file expected to contain valid html text.
	//
	// Any other type will result in an error.
	Doc interface{}

	/////////////

	Attributes interface{}
	Parameters interface{}
	Returns    interface{}
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
