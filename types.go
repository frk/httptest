package httptest

import (
	"fmt"
	"net/url"
	"strings"
)

// The Endpoint type describes an API endpoint to be tested.
type Endpoint struct {
	// The endpoint's HTTP method (or verb).
	Method string
	// The endpoint's URL path pattern.
	Pattern string
}

// String returns the result of concatenating the Endpoint's fields.
func (e Endpoint) String() string {
	return e.Method + " " + e.Pattern
}

// A TestGroup represents a set of tests to be executed against a specific endpoint.
type TestGroup struct {
	// A short description of what the endpoint-under-test is for. The httpdoc
	// package uses the description to generate the link text of the corresponding
	// sidebar item and the heading text for the associated documentation.
	Desc string
	// The endpoint to be tested.
	Endpoint Endpoint
	// The list of tests that will be executed against the endpoint.
	Tests []*Test
	// Indicates that the TestGroup should be skipped by the test runner.
	Skip bool
}

// The Test type describes the HTTP request to be sent to an endpoint and the
// corresponding HTTP response that is expected to be received for that request.
type Test struct {
	// The request to be sent to the endpoint under test.
	Request Request
	// The expected response to the request.
	Response Response
	// SetupAndTeardown is a two-func chain that can be used to setup and
	// teardown the API's internal state needed for an endpoint's test.
	//
	// The setup function is the first one in the chain and it is invoked
	// before a test is executed. The teardown, returned by the setup, is
	// the second one in the chain and it is invoked after the test is executed.
	SetupAndTeardown func(ep Endpoint, t *Test) (teardown func() error, err error)
	// Indicates that the Test should be skipped by the test runner.
	Skip bool
	// A short description of the test.
	Desc string
}

// The Request type describes the data to be sent in a single HTTP request.
type Request struct {
	// TODO
	Auth interface{}

	// The path parameters to be substituted in an endpoint pattern.
	Params ParamSetter
	// The URL query parameters to be appended to an endpoint's path.
	Query QueryEncoder
	// The HTTP header to be sent with the request.
	Header Header
	// The request body to be sent.
	Body Body
}

// Response is used to describe the expected HTTP response to a request.
type Response struct {
	// The expected HTTP status code.
	StatusCode int
	// The expected HTTP response headers.
	Header Header
	// The expected response body.
	Body Body
}

// A Header represents the key-value pairs in an HTTP header.
type Header map[string][]string

// The QueryEncoder returns a string of query parameters in the "URL encoded" form.
type QueryEncoder interface {
	QueryEncode() string
}

// Query is a QueryEncoder that returns its' contents encoded into "URL encoded" form.
type Query url.Values

// compiler check
var _ QueryEncoder = Query(nil)

// QueryEncode encodes the Query's underlying values into "URL encoded"
// form. QueryEncode uses net/url's Values.Encode to encode the values,
// see the net/url documentation for more info.
func (q Query) QueryEncode() string {
	return url.Values(q).Encode()
}

// The ParamSetter substitutes the placeholders of an endpoint pattern with parameter values.
//
// SetParams should return a copy of the given pattern with all of its placeholders
// replaced with actual parameter values. How the placeholders should be demarcated
// in the pattern depends on the implementation of the interface.
type ParamSetter interface {
	SetParams(pattern string) string
}

// Params is a ParamSetter that substitues an endpoint pattern's placeholders with
// its mapped values. The Params' keys represent the placeholders while the values
// are the actual parameters to be used to substitue those placeholders.
type Params map[string]interface{}

// compiler check
var _ ParamSetter = Params(nil)

// SetParams returns a copy of the given pattern replacing all of its placeholders
// with the actual parameter values contained in Params. The placeholders, used as
// keys to get the corresponding parameter values, are expected to be demarcated
// with curly braces; e.g.
//
//	pattern := "/users/{user_id}"
//	params := Params{"user_id": 123}
//	path := params.SetParams(pattern)
//	fmt.Println(path)
//	// outputs "/users/123"
//
func (pp Params) SetParams(pattern string) (path string) {
	var i, j int

	for {
		if i = strings.IndexByte(pattern, '{'); i > -1 {
			if j = strings.IndexByte(pattern, '}'); j > -1 && j > i {
				if v, ok := pp[pattern[i+1:j]]; ok {
					path += pattern[:i] + fmt.Sprintf("%v", v)
				}
				pattern = pattern[j+1:]
				continue
			}
		}
		break
	}
	return path + pattern
}
