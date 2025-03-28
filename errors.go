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
	// The failed test.
	test *test
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

func (e *testError) RequestDump() string {
	if len(e.test.reqdump) > 0 {
		return string(e.test.reqdump)
	}
	return ""
}

func (e *testError) ResponseDump() string {
	if len(e.test.resdump) > 0 {
		return string(e.test.resdump)
	}
	return ""
}

func (e *testError) TestName() string {
	return e.test.name
}

func (e *testError) TestIndex() string {
	return strconv.Itoa(e.test.index)
}

func (e *testError) EndpointString() string {
	return `"` + e.test.endpoint.String() + `"`
}

func (e *testError) RequestHeader() string {
	if e.test.tt.Request.Header != nil {
		return fmt.Sprintf("%v", e.test.tt.Request.Header)
	}
	return ""
}

func (e *testError) RequestMethod() string {
	return e.test.req.Method
}

func (e *testError) RequestPath() string {
	return e.test.req.URL.Path
}

func (e *testError) RequestURL() string {
	return e.test.req.URL.String()
}

func (e *testError) RequestBodyType() string {
	return fmt.Sprintf("%T", e.test.tt.Request.Body)
}

func (e *testError) GotStatus() string {
	return strconv.Itoa(e.test.res.StatusCode)
}

func (e *testError) WantStatus() string {
	return strconv.Itoa(e.test.tt.Response.StatusCode)
}

func (e *testError) HeaderKey() string {
	return e.hkey
}

func (e *testError) GotHeader() string {
	return fmt.Sprintf("%+v", e.test.res.Header[e.hkey])
}

func (e *testError) WantHeader() string {
	header := e.test.tt.Response.Header.GetHeader()
	return fmt.Sprintf("%+v", header[e.hkey])
}

func (e *testError) Err() (out string) {
	return e.err.Error()
}

type errorCode uint8

func (e errorCode) name() string { return fmt.Sprintf("error_template_%d", e) }

const (
	_ errorCode = iota
	errTestStateInit
	errTestStateCheck
	errTestStateCleanup
	errRequestBodyReader
	errRequestNew
	errRequestSend
	errResponseStatus
	errResponseHeader
	errResponseBody
)

var output_template_string = `
{{ define "` + errTestStateInit.name() + `" -}}
{{Wb "frk/httptest"}}: Test state initialization returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errTestStateCheck.name() + `" -}}
{{Wb "frk/httptest"}}: Test state check returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errTestStateCleanup.name() + `" -}}
{{Wb "frk/httptest"}}: Test state cleanup returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestBodyReader.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
({{.RequestBodyType}}).Reader() call returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestNew.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
http.NewRequest call returned an error.
 - {{R .Err}}
{{ end }}

{{ define "` + errRequestSend.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
(*http.Client).Do call returned an error.
 - {{R .Err}}

{{ with .RequestDump -}}
REQUEST: {{Y .}}
{{- end }}
{{ end }}

{{ define "` + errResponseStatus.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
http.Response.StatusCode got={{R .GotStatus}}, want={{G .WantStatus}}

{{ with .RequestDump -}}
REQUEST: {{Y .}}
{{ end }}
{{ with .ResponseDump -}}
RESPONSE: {{Y .}}
{{ end }}
{{ end }}

{{ define "` + errResponseHeader.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
http.Response.Header["{{.HeaderKey}}"] got={{R .GotHeader}}, want={{G .WantHeader}}

{{ with .RequestDump -}}
REQUEST: {{Y .}}
{{ end }}
{{ with .ResponseDump -}}
RESPONSE: {{Y .}}
{{ end }}
{{ end }}

{{ define "` + errResponseBody.name() + `" -}}
{{Y .EndpointString}}
{{R .TestName}} test failed.
http.Response.Body mismatch:
{{.Err}}

{{ with .RequestDump -}}
REQUEST: {{Y .}}
{{ end }}
{{ with .ResponseDump -}}
RESPONSE: {{Y .}}
{{ end }}
{{ end }}

{{ define "test_report" }}
{{ .Label }}:
{{- with .Failed }}
> {{R "FAILED"}}: {{W .}} test(s).
{{- end }}
{{- with .Skipped }}
> {{Y "SKIPPED"}}: {{W .}} test(s).
{{- end -}}
{{- with .Passed }}
> {{G "PASSED"}}: {{W .}} test(s).
{{- end }}
{{/* empty line */}}
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
	"Y":  func(v ...string) string { return getcolor("\033[0;93m", v) },
	"Yb": func(v ...string) string { return getcolor("\033[1;93m", v) },
	// yellow color, underlined (terminal)
	"y": func(v ...string) string { return getcolor("\033[0;33m", v) },
	// blue color HI (terminal)
	"B": func(v ...string) string { return getcolor("\033[0;94m", v) },
	// cyan color HI (terminal)
	"C": func(v ...string) string { return getcolor("\033[0;96m", v) },
	"c": func(v ...string) string { return getcolor("\033[0;36m", v) },
	// white color HI (terminal)
	"W":  func(v ...string) string { return getcolor("\033[0;97m", v) },
	"Wb": func(v ...string) string { return getcolor("\033[1;97m", v) },
	// no color (terminal)
	"off": func() string { return "\033[0m" },
	// quote the given string
	"q": func(v string) string { return strconv.Quote(v) },
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
