package httptest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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

type Config struct {
	// The host URL from which the target API is being served.
	HostURL string
	// The HTTP client to be used for sending test requests to the API.
	// If nil, a default client will be used.
	Client *http.Client
	// SetupAndTeardown is a two-func chain that can be used to setup and
	// teardown the API's internal state needed for an endpoint's test.
	//
	// The setup function is the first one in the chain and it is invoked
	// before a test is executed, the teardown, returned by the setup, is
	// the second one in the chain and it is invoked after the test is executed.
	//
	// If a Test instance has its own SetupAndTeardown, then that will be
	// used instead of this one.
	SetupAndTeardown func(e E, t *Test) (teardown func() error, err error)

	mu sync.RWMutex
	// The number of passed tests.
	passed int
	// The number of failed tests.
	failed int
	// The number of skipped tests.
	skipped int
}

func (c *Config) Run(t *testing.T, tgs []*TestGroup) {
	c.run(testing_t{t}, tgs)
}

func (c *Config) run(t testing_T, tgs []*TestGroup) {
	var passed, failed, skipped int
	for _, tg := range tgs {
		if tg.Skip {
			skipped += len(tg.Tests)
			continue
		}

		method, pattern := tg.E.Split()

		gName := tg.GetName()
		for i, tt := range tg.Tests {
			if tt.Skip {
				skipped += 1
				continue
			}

			name := tt.GetName()
			if len(name) == 0 {
				name = fmt.Sprintf("%02d", i)
			}
			if len(tg.N) > 0 {
				name = gName + "/" + name
			}

			t.Run(name, func(t testing_T) {
				sat := tt.SetupAndTeardown
				if sat == nil {
					sat = c.SetupAndTeardown
				}

				ts := &tstate{
					host:    c.HostURL,
					method:  method,
					pattern: pattern,
					name:    name,
					e:       tg.E,
					sat:     sat,
					i:       i,
					tt:      tt,
				}
				if err := runtest(ts, c.client()); err != nil {
					t.Error(err)
					failed += 1
				} else {
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

func (c *Config) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return defaultClient
}

// LogReport logs a summary of the test to stderr. It is intended to be called at "teardown".
func (c *Config) LogReport() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var report = struct {
		Passed, Failed, Skipped string
	}{}

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

// The tstate type holds the state of a test.
type tstate struct {
	host    string
	method  string
	pattern string
	name    string
	e       E
	sat     func(e E, t *Test) (teardown func() error, err error) `cmp:"-"`
	i       int
	tt      *Test          `cmp:"+"`
	req     *http.Request  `cmp:"+"`
	res     *http.Response `cmp:"+"`

	reqdump []byte
	resdump []byte
}

func runtest(s *tstate, c *http.Client) (e error) {
	if s.sat != nil {
		teardown, err := s.sat(s.e, s.tt)
		if err != nil {
			return &testError{code: errTestSetup, s: s, err: err}
		}
		if teardown != nil {
			defer func() {
				if err := teardown(); err != nil && e == nil {
					e = &testError{code: errTestTeardown, s: s, err: err}
				}
			}()
		}
	}

	if err := initrequest(s); err != nil {
		return err
	}

	// send request
	res, err := c.Do(s.req)
	if err != nil && !errors.Is(err, redirect) {
		return &testError{code: errRequestSend, s: s, err: err}
	}
	if s.tt.Response.DumpOnError {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			return err
		}
		s.resdump = dump
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	s.res = res

	if err := checkresponse(s); err != nil {
		return err
	}
	return nil
}

// initrequest initializes an http request from the test's Request value.
func initrequest(s *tstate) error {
	method, path := s.method, s.pattern

	if s.tt.Request.Params != nil {
		path = s.tt.Request.Params.SetParams(path)
	}
	if s.tt.Request.Query != nil {
		path += "?" + s.tt.Request.Query.GetQuery()
	}
	url := s.host + path

	// prepare the body
	body, err := io.Reader(nil), error(nil)
	if s.tt.Request.Body != nil {
		body, err = s.tt.Request.Body.Reader()
		if err != nil {
			return &testError{code: errRequestBodyReader, s: s, err: err}
		}
	}

	// initialize the request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &testError{code: errRequestNew, s: s, err: err}
	}
	if s.tt.Request.DumpOnError {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return err
		}
		s.reqdump = dump
	}

	// set the necessary headers
	if s.tt.Request.Body != nil {
		req.Header.Set("Content-Type", s.tt.Request.Body.Type())
	}
	if s.tt.Request.Header != nil {
		h := s.tt.Request.Header.GetHeader()
		for k, vv := range h {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}
	if s.tt.Request.Auth != nil {
		s.tt.Request.Auth.SetAuth(req, s.tt.Request)
	}

	s.req = req
	return nil
}

// checkresponse checks the http response against the test's Response value.
func checkresponse(s *tstate) error {
	var errs errorList

	if s.tt.Response.StatusCode != s.res.StatusCode {
		return &testError{code: errResponseStatus, s: s}
	}
	if s.tt.Response.Header != nil {
		wantHeader := s.tt.Response.Header.GetHeader()
		for key, _ := range wantHeader {
			wantVals, gotVals := wantHeader[key], s.res.Header[key]

		wantloop:
			for _, want := range wantVals {
				for _, got := range gotVals {
					if got == want {
						continue wantloop
					}
				}

				errs = append(errs, &testError{code: errResponseHeader, s: s, hkey: key})
				break wantloop
			}
		}
	}
	if s.tt.Response.Body != nil {
		if err := s.tt.Response.Body.Compare(s.res.Body); err != nil {
			err = &testError{code: errResponseBody, s: s, err: err}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// The testing_T interface represents a tiny portion of the *testing.T functionality which
// is being used by the Config.run method. It's raison d'etre is to make Config.run testable.
type testing_T interface {
	Error(args ...interface{})
	Run(name string, f func(testing_T)) bool
}

// testing_t is a wrapper around *testing.T that satisfies the testing_T interface.
type testing_t struct {
	t *testing.T
}

func (tt testing_t) Error(args ...interface{}) {
	tt.t.Error(args...)
}

func (tt testing_t) Run(name string, f func(testing_T)) bool {
	return tt.t.Run(name, func(t *testing.T) { f(testing_t{t}) })
}

func aorb(a, b string) string {
	if len(a) > 0 {
		return a
	}
	return b
}
