package httptest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// The E string type represents the endpoint to be tested, it is expected to be
// of the format "METHOD PATTERN", i.e. the endpoint's HTTP method, followed by
// a single space which, in turn, is followed by the endpoint's URL path pattern.
// For example: "GET /foo/{id}/bar".
//
// If the value provided by the user doesn't match the prescribed format the
// package will try to default to some sensible value however the result of
// that may not be what the user expects. It is therefore the user's
// responsibility to provide the value in the expected format.
type E string

func (e E) String() string {
	m, p := e.Split()
	return m + " " + p
}

// Split splits the endpoint into two strings, the method and the pattern, and returns them.
func (e E) Split() (method, pattern string) {
	s := strings.TrimSpace(string(e))
	if i := strings.IndexByte(s, ' '); i > -1 {
		method = strings.TrimSpace(s[:i])
		pattern = strings.TrimSpace(s[i+1:])
	} else {
		// If there is no space at which to split the string assume it
		// is either empty, or it contains only a pattern and no method.
		// Default to GET either way.
		method = "GET"
		pattern = s
	}

	// If the pattern is empty or doesn't start with a slash, prefix it with one.
	if len(pattern) == 0 || pattern[0] != '/' {
		pattern = "/" + pattern
	}

	return method, pattern
}

// A TestGroup is a set of tests to be executed against a specific endpoint.
type TestGroup struct {
	// E, short for "endpoint, holds the endpoint to be tested.
	E E
	// N or Name hold the TestGroup's name. Name takes precedence over N.
	//
	// The httptest package uses the name to identify the TestGroup's subtests.
	// The httpdoc package uses the name to generate the link text of the corresponding
	// sidebar item and the heading text for the associated documentation.
	N, Name string
	// The list of tests that will be executed against the endpoint.
	Tests []*Test
	// Indicates that the TestGroup should be skipped by the test runner.
	Skip bool
	// DocA and DocB are optional, they are ignored by the httptest package
	// and are used only by the httpdoc package. The httpdoc package uses
	// the first Test's Request and Response to generate input/output docs
	// for the resulting article.
	//
	// DocA, if set, is used by httpdoc to generate documentation [A]bove the input/output docs.
	// DocB, if set, is used by httpdoc to generate documentation [B]elow the input/output docs.
	//
	// The following types are supported:
	//	- string
	//	- *os.File
	//	- httpdoc.HTMLer
	//	- httpdoc.Valuer
	//	- interface{} (where the dynamic type MUST be declared)
	// Anything else will result in an error.
	//
	// If the type is string, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is *os.File, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is httpdoc.HTMLer, then its HTML() method will be used
	// to retrieve the content and it is up to the user to ensure that that
	// content is valid and safe HTML.
	//
	// If the type is httpdoc.Valuer, then its Value() method will be used
	// to get the underlying Value, the source of that Value's dynamic type
	// is then analyzed and any relevant documentation that's found will be
	// used to generate the HTML text. If the dynamic type is unnamed an
	// error will be returned instead.
	//
	// If the type is none of the above, then the type's source is analyzed
	// and any relevant documentation that's found will be used to generate
	// the HTML text. If the type is unnamed an error will be returned.
	DocA, DocB interface{}
	// If set, httpdoc will not generate documentation from this TestGroup.
	SkipDoc bool
}

// The Test type describes the HTTP request to be sent to an endpoint and the
// corresponding HTTP response that is expected to be received for that request.
type Test struct {
	// The request to be sent to the endpoint under test.
	Request Request
	// The expected response to the request.
	Response Response
	// N or Name hold the Test's name. Name takes precedence over N.
	//
	// The httptest package uses the name to identify the Test's subtest.
	N, Name string
	// State can optionally be set to a value that represents the test's state.
	// If set, it will be passed to the Config.StateHandler to manage the test's state.
	State State
	// Indicates that the Test should be skipped by the test runner.
	Skip bool
	// DocA and DocB are optional, they are ignored by the httptest package
	// and are used only by the httpdoc package. The httpdoc package uses the
	// Test's Request and Response to generate example docs for the resulting
	// article.
	//
	// DocA, if set, is used by httpdoc to generate documentation [A]bove the example docs.
	// DocB, if set, is used by httpdoc to generate documentation [B]elow the example docs.
	//
	// The following types are supported:
	//	- string
	//	- *os.File
	//	- httpdoc.HTMLer
	//	- httpdoc.Valuer
	//	- interface{} (named types only)
	// Anything else will result in an error.
	//
	// If the type is string, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is *os.File, it is expected to contain raw HTML and it is
	// up to the user to ensure that that HTML is valid and safe.
	//
	// If the type is httpdoc.HTMLer, then its HTML() method will be used
	// to retrieve the content and it is up to the user to ensure that that
	// content is valid and safe HTML.
	//
	// If the type is httpdoc.Valuer, then its Value() method will be used
	// to get the underlying Value, the source of that Value's dynamic type
	// is then analyzed and any relevant documentation that's found will be
	// used to generate the HTML text. If the dynamic type is unnamed an
	// error will be returned instead.
	//
	// If the type is none of the above, then the type's source is analyzed
	// and any relevant documentation that's found will be used to generate
	// the HTML text. If the type is unnamed an error will be returned.
	DocA, DocB interface{}
}

// The Request type describes the data to be sent in a single HTTP request.
type Request struct {
	// The auth information to be sent with the request.
	//
	// [httpdoc]: If the AuthSetter's type implements either the httpdoc.HTMLer
	// interface or the httpdoc.Valuer interface, then it will be used by httpdoc
	// to produce auth-specific documentation.
	Auth AuthSetter
	// The HTTP header to be sent with the request.
	//
	// [httpdoc]: If the HeaderGetter's type implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce input-specific documentation.
	Header HeaderGetter
	// The path parameters to be substituted in an endpoint pattern.
	//
	// [httpdoc]: If the ParamSetter's type implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce input-specific documentation.
	Params ParamSetter
	// The URL query parameters to be appended to an endpoint's path.
	//
	// [httpdoc]: If the QueryGetter's type implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce input-specific documentation.
	Query QueryGetter
	// The request body to be sent.
	//
	// [httpdoc]: If the Body's type implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce input-specific documentation.
	Body Body
	// If set to true and the test fails, a dump of the HTTP request
	// will be included in the test's output.
	DumpOnFail bool
	// If set to true, a dump of the HTTP request will be included in the test's output.
	Dump bool
}

// Response is used to describe the expected HTTP response to a request.
type Response struct {
	// The expected HTTP status code.
	StatusCode int
	// The expected HTTP response headers.
	//
	// [httpdoc]: If the HeaderGetter's type implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce output-specific documentation.
	Header HeaderGetter
	// The expected response body.
	//
	// [httpdoc]: If the Body's type also implements the httpdoc.Valuer interface,
	// then it will be used by httpdoc to produce output-specific documentation.
	Body Body
	// If set to true and the test fails, a dump of the HTTP response
	// will be included in the test's output.
	DumpOnFail bool
	// If set to true, a dump of the HTTP response will be included in the test's output.
	Dump bool
}

////////////////////////////////////////////////////////////////////////////////
// interfaces
////////////////////////////////////////////////////////////////////////////////

// The ParamSetter substitutes the placeholders of an endpoint pattern with actual parameter values.
//
// SetParams should return a copy of the given pattern with all of its placeholders
// replaced with actual parameter values. How the placeholders should be demarcated
// in the pattern depends on the implementation of the interface.
type ParamSetter interface {
	SetParams(pattern string) string
}

// The QueryGetter returns a string of query parameters in the "URL encoded" form.
type QueryGetter interface {
	GetQuery() string
}

// AuthSetter updates an HTTP request with auth information.
type AuthSetter interface {
	SetAuth(r *http.Request, t Request)
}

// HeaderGetter returns an HTTP header.
type HeaderGetter interface {
	GetHeader() http.Header
}

// The Body type represents the contents of an HTTP request or response body.
//
// The httptype package contains a number of useful implementations.
type Body interface {
	// Type returns the string used in an HTTP request's Content-Type header.
	Type() string
	// Reader returns an io.Reader that provides the content of an HTTP request's body.
	Reader() (body io.Reader, err error)
	// Compare returns the result of the comparison between the Body's contents
	// and the contents of the given io.Reader. The level of strictness of the
	// comparison depends on the implementation. If the contents are equivalent
	// the returned error will be nil, otherwise the error will describe the
	// negative result of the comparison.
	Compare(body io.Reader) error
}

////////////////////////////////////////////////////////////////////////////////
// implementations
////////////////////////////////////////////////////////////////////////////////

// compiler check
var _ HeaderGetter = Header(nil)

// A Header represents the key-value pairs in an HTTP header.
type Header map[string][]string

// GetHeader returns the receiver as http.Header.
func (h Header) GetHeader() http.Header {
	return http.Header(h)
}

// compiler check
var _ QueryGetter = Query(nil)

// Query is a QueryGetter that returns its' contents encoded into "URL encoded" form.
type Query url.Values

// GetQuery encodes the Query's underlying values into "URL encoded"
// form. GetQuery uses net/url's Values.Encode to encode the values,
// see the net/url documentation for more info.
func (q Query) GetQuery() string {
	return url.Values(q).Encode()
}

// compiler check
var _ ParamSetter = Params(nil)

// Params is a ParamSetter that substitues an endpoint pattern's placeholders with
// its mapped values. The Params' keys represent the placeholders while the values
// are the actual parameters to be used to substitue those placeholders.
type Params map[string]interface{}

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
