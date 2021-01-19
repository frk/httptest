package httptest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/frk/compare"
)

type fakeT struct {
	testing.TB
	errs []interface{}
}

func (ft *fakeT) Error(args ...interface{}) {
	ft.errs = append(ft.errs, args...)
}

type errorTransport struct{}

func (errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("bad round trip")
}

func Test_Config_Run(t *testing.T) {
	server := &http.Server{Addr: "localhost:3456"}
	defer server.Close()

	var handler http.Handler
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	conf := Config{HostURL: "http://" + server.Addr}

	host := "http://localhost:3456"
	tests := []struct {
		name     string
		tabs     []*TestTable
		handler  http.Handler
		want     []interface{}
		printerr bool
		rt       http.RoundTripper
	}{{
		// make sure that the request is sent to the correct endpoint #1
		name: "ept_test_1", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Request:  Request{},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/v1/foo" {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the request is sent to the correct endpoint #2
		name: "ept_test_2", tabs: []*TestTable{{Ept: "GET /v1/bar/baz", Tests: []*Test{{
			Request:  Request{},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" || r.URL.Path != "/v1/bar/baz" {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the Request.Params are correctly inserted into the http.Request's path
		name: "params_test", tabs: []*TestTable{{Ept: "GET /v1/foo/{id}/bar/{some_string}/baz/{boolean}", Tests: []*Test{{
			Request:  Request{Params: Params{"id": 87654, "some_string": "xyz", "boolean": true}},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/foo/87654/bar/xyz/baz/true" {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the Request.Query is correctly used to set the http.Request's query parameters
		name: "query_test", tabs: []*TestTable{{Ept: "GET /v1/foo", Tests: []*Test{{
			Request:  Request{Query: Query{"q": {"term"}, "page": {"4"}, "max": {"25"}}},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := Query(r.URL.Query())
			want := Query{"q": {"term"}, "page": {"4"}, "max": {"25"}}
			if e := compare.Compare(got, want); e != nil {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the Request.Header is correctly added to the http.Request's Header
		name: "header_test", tabs: []*TestTable{{Ept: "GET /v1/foo", Tests: []*Test{{
			Request:  Request{Header: Header{"A": {"Foo", "Bar"}, "B": {"BAZ"}}},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a, b := r.Header["A"], r.Header["B"]
			if e := compare.Compare(a, []string{"Foo", "Bar"}); e != nil {
				w.WriteHeader(500)
			}
			if e := compare.Compare(b, []string{"BAZ"}); e != nil {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the Request.Body is correctly sent as the http.Request's Body #1
		name: "body_test_1", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Request:  Request{Body: Text(`foo bar baz`)},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if ct != textContentType {
				w.WriteHeader(500)
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil || string(body) != `foo bar baz` {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure that the Request.Body is correctly sent as the http.Request's Body #2
		name: "body_test_2", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Request:  Request{Body: JSON([]interface{}{123, "foo", "bar"})},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if ct != jsonContentType {
				w.WriteHeader(500)
			}
			body, err := ioutil.ReadAll(r.Body)
			if err != nil || string(body) != `[123,"foo","bar"]` {
				w.WriteHeader(500)
			}
		}),
	}, {
		// make sure the test fails if TestTable.Ept does not match the prescribed format #1
		name: "bad_ept_1", tabs: []*TestTable{{Ept: "", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want:    []interface{}{&testError{code: errTableEpt, s: &tstate{host: host, ept: "", tt: &Test{}}}},
	}, {
		// make sure the test fails if TestTable.Ept does not match the prescribed format #2
		name: "bad_ept_2", tabs: []*TestTable{{Ept: "/v1/foo", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want:    []interface{}{&testError{code: errTableEpt, s: &tstate{host: host, ept: "/v1/foo", tt: &Test{}}}},
	}, {
		// make sure the test fails if TestTable.Ept does not match the prescribed format #3
		name: "bad_ept_3", tabs: []*TestTable{{Ept: "POST", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want:    []interface{}{&testError{code: errTableEpt, s: &tstate{host: host, ept: "POST", tt: &Test{}}}},
	}, {
		// make sure the test fails if TestTable.Ept does not match the prescribed format #4
		name: "bad_ept_4", tabs: []*TestTable{{Ept: "POST /v1/foo bar", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want:    []interface{}{&testError{code: errTableEpt, s: &tstate{host: host, ept: "POST /v1/foo bar", tt: &Test{}}}},
	}, {
		// make sure the test fails if TestTable.Ept does not match the prescribed format #5
		name: "bad_ept_5", tabs: []*TestTable{{Ept: "POST  /v1/foo", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want:    []interface{}{&testError{code: errTableEpt, s: &tstate{host: host, ept: "POST  /v1/foo", tt: &Test{}}}},
	}, {
		// make sure the error from Request.Body.Reader is returned
		name: "request_body_reader", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Request: Request{Body: JSON([]func(){nil})},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{&testError{code: errRequestBodyReader, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}},
			err: errors.New("dummy")}},
	}, {
		// make sure the error from http.NewRequest is returned
		name: "http_new_request", tabs: []*TestTable{{Ept: "世界 /v1/foo", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{&testError{code: errRequestNew, s: &tstate{host: host, ept: "世界 /v1/foo", tt: &Test{}},
			err: errors.New("dummy")}},
	}, {
		// make sure the error from http.Client.Do is returned
		name: "http_client_do", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{&testError{code: errRequestSend, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}, req: &http.Request{}},
			err: errors.New("dummy")}},
		rt: errorTransport{},
	}, {
		// make sure the test fails if response status code is not as expected
		name: "response_status_mismatch", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}),
		want: []interface{}{&testError{code: errResponseStatus,
			s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}, req: &http.Request{}, res: &http.Response{}}}},
	}, {
		// make sure the test fails if response header is not as expected
		name: "response_header_mismatch", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Response: Response{Status: 200, Header: Header{"A": {"foo"}, "B": {"bar"}, "C": {"baz"}}},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("A", "abc")
			w.Header().Set("B", "bar")
		}),
		want: []interface{}{errorList{
			&testError{code: errResponseHeader, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}, req: &http.Request{}, res: &http.Response{}}, hkey: "A"},
			&testError{code: errResponseHeader, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}, req: &http.Request{}, res: &http.Response{}}, hkey: "C"},
		}},
	}, {
		// make sure the test fails if response body is not as expected
		name: "response_body_mismatch", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			Response: Response{Status: 200, Body: JSON(jsonstruct{A: "foo", B: 123})},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(jsonstruct{A: "bar", B: 321})
		}),
		want: []interface{}{errorList{
			&testError{code: errResponseBody, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}, req: &http.Request{}, res: &http.Response{}}, err: errors.New("dummy")},
		}},
	}, {
		// make sure the error from the setup function is returned
		name: "setup_error", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			SetupAndTeardown: func(ept string, t *Test) (teardown func() error, err error) {
				return nil, errors.New("setup fail")
			},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{
			&testError{code: errTestSetup, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{}}, err: errors.New("setup fail")},
		},
	}, {
		// make sure the error from the teardown function is returned
		name: "teardown_error", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			SetupAndTeardown: func(ept string, t *Test) (teardown func() error, err error) {
				return func() error { return errors.New("teardown fail") }, nil
			},
			Response: Response{Status: 200},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{
			&testError{code: errTestTeardown, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{},
				req: &http.Request{}, res: &http.Response{}}, err: errors.New("teardown fail")},
		},
	}, {
		// make sure the error from the test takes precedence over the error from teardown
		name: "test_error_over_teardown_error", tabs: []*TestTable{{Ept: "POST /v1/foo", Tests: []*Test{{
			SetupAndTeardown: func(ept string, t *Test) (teardown func() error, err error) {
				return func() error { return errors.New("teardown fail") }, nil
			},
			Response: Response{Status: 234},
		}}}},
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		want: []interface{}{
			&testError{code: errResponseStatus, s: &tstate{host: host, ept: "POST /v1/foo", tt: &Test{},
				req: &http.Request{}, res: &http.Response{}}},
		},
	}}

	cmp := compare.Config{ObserveFieldTag: "cmp", IgnoreArrayOrder: true}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rt != nil {
				conf.Client = &http.Client{Transport: tt.rt}
				defer func() { conf.Client = nil }()
			}

			handler = tt.handler

			ft := &fakeT{}
			conf.Run(ft, tt.tabs)
			if e := cmp.Compare(ft.errs, tt.want); e != nil {
				t.Error(e)
			}
			if tt.printerr {
				for _, v := range ft.errs {
					fmt.Println(v)
				}
			}
		})
	}
}
