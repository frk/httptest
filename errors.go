package httptest

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type errorList []error

func (list errorList) Error() (s string) {
	for _, e := range list {
		s += e.Error()
	}
	return s
}

type testError struct {
	code errorCode
	// The state of the failed test.
	s *tstate
	// The original error, or nil.
	err error `cmp:"+"`
	// The header key in case of errResponseHeader, or empty.
	hkey string
}

func (e *testError) Error() string {
	sb := new(strings.Builder)
	if err := output_templates.ExecuteTemplate(sb, e.code.name(), e); err != nil {
		panic(err)
	}
	return sb.String()
}

func (e *testError) RequestBody() string {
	if e.s.tt.Request.Body != nil {
		return fmt.Sprintf("%v", e.s.tt.Request.Body)
	}
	return ""
}

func (e *testError) TestIndex() string {
	return strconv.Itoa(e.s.i)
}

func (e *testError) EndpointString() string {
	return e.s.ep.String()
}

func (e *testError) RequestHeader() string {
	if len(e.s.tt.Request.Header) > 0 {
		return fmt.Sprintf("%v", e.s.tt.Request.Header)
	}
	return ""
}

func (e *testError) RequestMethod() string {
	return e.s.req.Method
}

func (e *testError) RequestPath() string {
	return e.s.req.URL.Path
}

func (e *testError) RequestURL() string {
	return e.s.req.URL.String()
}

func (e *testError) RequestBodyType() string {
	return fmt.Sprintf("%T", e.s.tt.Request.Body)
}

func (e *testError) GotStatus() string {
	return strconv.Itoa(e.s.res.StatusCode)
}

func (e *testError) WantStatus() string {
	return strconv.Itoa(e.s.tt.Response.StatusCode)
}

func (e *testError) HeaderKey() string {
	return e.hkey
}

func (e *testError) GotHeader() string {
	return fmt.Sprintf("%+v", e.s.res.Header[e.hkey])
}

func (e *testError) WantHeader() string {
	return fmt.Sprintf("%+v", e.s.tt.Response.Header[e.hkey])
}

func (e *testError) Err() (out string) {
	return e.err.Error()
}

type errorCode uint8

func (e errorCode) name() string { return fmt.Sprintf("error_template_%d", e) }

const (
	_ errorCode = iota
	errTestSetup
	errTestTeardown
	errRequestBodyReader
	errRequestNew
	errRequestSend
	errResponseStatus
	errResponseHeader
	errResponseBody
)

var output_template_string = `
{{ define "` + errTestSetup.name() + `" -}}
{{Wb "frk/httptest"}}: Test setup returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errTestTeardown.name() + `" -}}
{{Wb "frk/httptest"}}: Test teardown returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestBodyReader.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} ({{.RequestBodyType}}).Reader() call returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestNew.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} http.NewRequest call returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestSend.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} (*http.Client).Do call returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errResponseStatus.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} failed.
http.Response.StatusCode got={{R .GotStatus}}, want={{C .WantStatus}}
{{- with .RequestHeader }}
 - Request.Header: {{Y .}}
{{ end }}
{{- with .RequestBody }}
 - Request.Body: {{Y .}}
{{ end }}
{{ end }}

{{ define "` + errResponseHeader.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} failed.
http.Response.Header["{{.HeaderKey}}"] got={{R .GotHeader}}, want={{C .WantHeader}}
{{- with .RequestHeader }}
 - Request.Header: {{Y .}}
{{ end }}
{{- with .RequestBody }}
 - Request.Body: {{Y .}}
{{ end }}
{{ end }}

{{ define "` + errResponseBody.name() + `" -}}
{{Wb "frk/httptest"}}: "{{.EndpointString}}" test #{{.TestIndex}} failed.
http.Response.Body mismatch:
{{.Err}}
{{- with .RequestHeader }}
 - Request.Header: {{Y .}}
{{ end }}
{{- with .RequestBody }}
 - Request.Body: {{Y .}}
{{ end }}
{{ end }}

{{ define "test_report" -}}
{{G "Note"}}: Passed {{.Passed}} test(s).
{{- with .Skipped}}
{{Y "Warning"}}: Skipped {{.}} test(s).
{{ end }}
{{- with .Failed}}
{{R "Error"}}: Failed {{.}} test(s).
{{ end }}
{{ end }}
` // `

var output_templates = template.Must(template.New("t").Funcs(template.FuncMap{
	// white color (terminal)
	"wi": func(v ...string) string { return getcolor("\033[3;37m", v) },

	// red color HI (terminal)
	"R": func(v ...string) string { return getcolor("\033[0;91m", v) },
	// green color HI (terminal)
	"G": func(v ...string) string { return getcolor("\033[0;92m", v) },
	// yellow color HI (terminal)
	"Y": func(v ...string) string { return getcolor("\033[0;93m", v) },
	// blue color HI (terminal)
	"B": func(v ...string) string { return getcolor("\033[0;94m", v) },
	// cyan color HI (terminal)
	"C": func(v ...string) string { return getcolor("\033[0;96m", v) },
	// white color HI (terminal)
	"W":  func(v ...string) string { return getcolor("\033[0;97m", v) },
	"Wb": func(v ...string) string { return getcolor("\033[1;97m", v) },
	// no color (terminal)
	"off": func() string { return "\033[0m" },
}).Parse(output_template_string))

func getcolor(c string, v []string) string {
	if len(v) > 0 {
		return fmt.Sprintf("%s%v\033[0m", c, stringsStringer(v))
	}
	return c
}

type stringsStringer []string

func (s stringsStringer) String() string {
	return strings.Join([]string(s), "")
}
