package httptest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"testing"
)

// redirect is used to signal a redirect.
var redirect = errors.New("frk/httptest:redirect")

// defaultClient is the default http.Client used by a TestRunner.
// The client is setup to not follow redirects.
var defaultClient = &http.Client{
	CheckRedirect: func(r *http.Request, via []*http.Request) error {
		return redirect
	},
}

// A Config specifies the configuration for test running. The zero value
// of Config is a ready-to-use default configuration.
type Config struct {
	// The URL of the host from which the target API is being served.
	//
	// If Host is left empty then Run will automatically start a test
	// server using its mux argument as the test server's handler.
	Host string
	// Label is the descriptive name of the set of tests ran by this config.
	// The label is primarily used in the logged report of the test results.
	Label string
	// The HTTP client to be used for sending test requests to the API.
	// If nil, a default client will be used.
	Client *http.Client
	// StateHandler, if set, will be used for managing the state of each test.
	StateHandler StateHandler

	// The base URL of the target API.
	url string
	// mu is used to synchronize access to the test results.
	mu sync.RWMutex
	// The number of passed tests.
	passed int
	// The number of failed tests.
	failed int
	// The number of skipped tests.
	skipped int
}

// Run executes the set of provided test groups. If the Config's Host is left
// empty then Run will automatically start a new test server using the mux argument
// as the test server's handler.
func (c *Config) Run(t *testing.T, tgs []*TestGroup, mux http.Handler) {
	if c.url = c.Host; c.url == "" {
		s := httptest.NewServer(mux)
		defer s.Close()

		c.url = s.URL
	}

	c.run(testing_t{t}, tgs)
}

func (c *Config) run(t T, tgs []*TestGroup) {
	var client = c.getClient()
	var passed, failed, skipped int
	for _, tg := range tgs {
		if tg.Skip {
			skipped += len(tg.Tests)
			continue
		}

		method, pattern := tg.E.Split()
		for i, tt := range tg.Tests {
			if tt.Skip {
				skipped += 1
				continue
			}

			name := c.testName(tt, tg, i)
			t.Run(name, func(t T) {
				x := &test{
					url:      c.url,
					client:   client,
					method:   method,
					pattern:  pattern,
					name:     name,
					index:    i,
					endpoint: tg.E,
					sh:       c.StateHandler,
					tt:       tt,
				}
				if err := x.exec(); err != nil {
					t.Error(err)
					failed += 1
				} else {
					x.print_dumps()
					passed += 1
				}
			})
		}
	}

	c.mu.Lock()
	c.passed += passed
	c.failed += failed
	c.skipped += skipped
	c.mu.Unlock()
}

// LogReport logs a summary of the test results to stderr.
func (c *Config) LogReport() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var report = struct {
		Title                   string
		Passed, Failed, Skipped string
	}{Title: c.Label}

	if c.passed > 0 {
		report.Passed = strconv.Itoa(c.passed)
	}
	if c.failed > 0 {
		report.Failed = strconv.Itoa(c.failed)
	}
	if c.skipped > 0 {
		report.Skipped = strconv.Itoa(c.skipped)
	}

	if err := output_templates.ExecuteTemplate(os.Stderr, "test_report", report); err != nil {
		panic(err)
	}
}

// getClient returns the http client that will be used for executing test requests.
func (c *Config) getClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return defaultClient
}

// testName constructs a name for the test t.
func (c *Config) testName(t *Test, g *TestGroup, i int) string {
	name := t.N
	if len(t.Name) > 0 {
		name = t.Name
	}
	if len(name) == 0 {
		// default to test index
		name = fmt.Sprintf("%02d", i)
	}

	// prefix test's name with test group's name
	group := g.N
	if len(g.Name) > 0 {
		group = g.Name
	}
	if len(group) > 0 {
		name = group + "/" + name
	}
	return name
}

// The test type manages the execution of an individual Test.
type test struct {
	client  *http.Client `cmp:"-"`
	url     string
	method  string
	pattern string
	sh      StateHandler   `cmp:"-"`
	tt      *Test          `cmp:"+"`
	req     *http.Request  `cmp:"+"`
	res     *http.Response `cmp:"+"`

	// the following are used for test result reporting
	name     string
	index    int
	endpoint E
	reqdump  []byte
	resdump  []byte
}

func (t *test) exec() (err error) {
	// initialize state & defer its cleanup
	if t.sh != nil && t.tt.State != nil {
		if err := t.sh.Init(t.tt.State); err != nil {
			return &testError{code: errTestStateInit, test: t, err: err}
		}
		defer func() {
			if e := t.sh.Cleanup(t.tt.State); e != nil && err == nil {
				err = &testError{code: errTestStateCleanup, test: t, err: e}
			}
		}()
	}

	if err := t.prepare_request(); err != nil {
		return err
	}
	if err := t.send_request(); err != nil {
		return err
	}
	defer t.close_response()

	if err := t.check_response(); err != nil {
		return err
	}

	// check state
	if t.sh != nil && t.tt.State != nil {
		if err := t.sh.Check(t.tt.State); err != nil {
			return &testError{code: errTestStateCheck, test: t, err: err}
		}
	}
	return nil
}

// prepare_request initializes an http request from the Test.Request value.
func (t *test) prepare_request() error {
	method, path := t.method, t.pattern
	if t.tt.Request.Params != nil {
		path = t.tt.Request.Params.SetParams(path)
	}
	if t.tt.Request.Query != nil {
		path += "?" + t.tt.Request.Query.GetQuery()
	}
	url := t.url + path

	// prepare the body
	body, err := io.Reader(nil), error(nil)
	if t.tt.Request.Body != nil {
		body, err = t.tt.Request.Body.Reader()
		if err != nil {
			return &testError{code: errRequestBodyReader, test: t, err: err}
		}
	}

	// initialize the request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &testError{code: errRequestNew, test: t, err: err}
	}
	t.req = req

	// set the necessary headers
	if t.tt.Request.Body != nil {
		t.req.Header.Set("Content-Type", t.tt.Request.Body.Type())
	}
	if t.tt.Request.Header != nil {
		h := t.tt.Request.Header.GetHeader()
		for k, vv := range h {
			for _, v := range vv {
				t.req.Header.Add(k, v)
			}
		}
	}
	if t.tt.Request.Auth != nil {
		t.tt.Request.Auth.SetAuth(t.req, t.tt.Request)
	}

	// retain a dump of the request for debugging
	if t.tt.Request.DumpOnFail || t.tt.Request.Dump {
		dump, err := httputil.DumpRequestOut(t.req, true)
		if err != nil {
			return err
		}
		t.reqdump = dump
	}
	return nil
}

// send_request sends the request and records the response.
func (t *test) send_request() (err error) {
	res, err := t.client.Do(t.req)
	if err != nil && !errors.Is(err, redirect) {
		return &testError{code: errRequestSend, test: t, err: err}
	}
	t.res = res

	defer func() {
		if err != nil {
			t.close_response()
		}
	}()

	if t.tt.Response.DumpOnFail || t.tt.Response.Dump {
		dump, err := httputil.DumpResponse(t.res, true)
		if err != nil {
			return err
		}
		t.resdump = dump
	}
	return nil
}

// checkresponse checks the http response against the Test.Response value.
func (t *test) check_response() error {
	var errs errorList

	// check the response status
	if t.tt.Response.StatusCode != t.res.StatusCode {
		return &testError{code: errResponseStatus, test: t}
	}

	// check the response headers
	if t.tt.Response.Header != nil {
		wantHeader := t.tt.Response.Header.GetHeader()
		for key, _ := range wantHeader {
			wantVals, gotVals := wantHeader[key], t.res.Header[key]

		wantloop:
			for _, want := range wantVals {
				for _, got := range gotVals {
					if got == want {
						continue wantloop
					}
				}

				errs = append(errs, &testError{code: errResponseHeader, test: t, hkey: key})
				break wantloop
			}
		}
	}

	// check the response body
	if t.tt.Response.Body != nil {
		if err := t.tt.Response.Body.Compare(t.res.Body); err != nil {
			err = &testError{code: errResponseBody, test: t, err: err}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (t *test) print_dumps() {
	if t.tt.Request.Dump && len(t.reqdump) > 0 {
		fmt.Printf("REQUEST: \033[0;93m%s\033[0m\n", string(t.reqdump))
	}
	if t.tt.Response.Dump && len(t.resdump) > 0 {
		fmt.Printf("RESPONSE: \033[0;93m%s\033[0m\n", string(t.resdump))
	}
}

// close_response closes the test response's body.
func (t *test) close_response() error {
	if t.res != nil && t.res.Body != nil {
		return t.res.Body.Close()
	}
	return nil
}

// The T interface represents a tiny portion of the *testing.T functionality
// which is being used by the Config.run method.
//
// It's primary raison d'etre is to make Config.run testable.
type T interface {
	Error(args ...interface{})
	Run(name string, f func(T)) bool
}

// testing_t is a wrapper around *testing.T that satisfies the T interface.
type testing_t struct {
	t *testing.T
}

func (tt testing_t) Error(args ...interface{}) {
	tt.t.Error(args...)
}

func (tt testing_t) Run(name string, f func(T)) bool {
	return tt.t.Run(name, func(t *testing.T) { f(testing_t{t}) })
}

func aorb(a, b string) string {
	if len(a) > 0 {
		return a
	}
	return b
}
