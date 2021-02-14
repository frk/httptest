package httptest

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
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
	SetupAndTeardown func(ept string, t *Test) (teardown func() error, err error)

	mu sync.RWMutex
	// The number of passed tests.
	passed int
	// The number of failed tests.
	failed int
	// The number of skipped tests.
	skipped int
}

func (c *Config) Run(t testing.TB, epts []*Endpoint) {
	var passed, failed, skipped int
	for _, e := range epts {
		if e.Skip {
			skipped += len(e.Tests)
			continue
		}

		for i, tt := range e.Tests {
			if tt.Skip {
				skipped += 1
				continue
			}

			sat := tt.SetupAndTeardown
			if sat == nil {
				sat = c.SetupAndTeardown
			}

			ts := &tstate{
				host: c.HostURL,
				ept:  e.Ept,
				sat:  sat,
				i:    i,
				tt:   tt,
			}
			if err := runtest(ts, c.client()); err != nil {
				t.Error(err)
				failed += 1
				continue
			}

			passed += 1
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
		Passed, Failed, Skipped int
	}{
		Passed:  c.passed,
		Failed:  c.failed,
		Skipped: c.skipped,
	}
	if err := output_templates.ExecuteTemplate(os.Stderr, "test_report", report); err != nil {
		panic(err)
	}
}

// The tstate type holds the state of a test.
type tstate struct {
	host string
	ept  string
	sat  func(ept string, t *Test) (teardown func() error, err error) `cmp:"-"`
	i    int
	tt   *Test          `cmp:"+"`
	req  *http.Request  `cmp:"+"`
	res  *http.Response `cmp:"+"`
}

func runtest(s *tstate, c *http.Client) (e error) {
	if s.sat != nil {
		teardown, err := s.sat(s.ept, s.tt)
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
	if err != nil && err != redirect {
		return &testError{code: errRequestSend, s: s, err: err}
	} else if res.Body != nil {
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
	// build the url
	method, path, err := splitept(s)
	if err != nil {
		return err
	}
	if s.tt.Request.Params != nil {
		path = s.tt.Request.Params.SetParams(path)
	}
	if s.tt.Request.Query != nil {
		path += "?" + s.tt.Request.Query.QueryEncode()
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

	// set the necessary headers
	if s.tt.Request.Body != nil {
		req.Header.Set("Content-Type", s.tt.Request.Body.ContentType())
	}
	for k, vv := range s.tt.Request.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
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
		for key, _ := range s.tt.Response.Header {
			wantVals, gotVals := s.tt.Response.Header[key], s.res.Header[key]

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
		if err := s.tt.Response.Body.CompareContent(s.res.Body); err != nil {
			if e, ok := err.(*testError); ok {
				e.s = s
			}
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// splitept splits the given Endpoint's Ept string into its parts (method and pattern).
func splitept(s *tstate) (method, pattern string, err error) {
	slice := strings.Split(s.ept, " ")
	if len(slice) == 2 {
		method, pattern = slice[0], slice[1]
	}
	if len(method) == 0 || len(pattern) == 0 {
		return "", "", &testError{code: errEndpointEpt, s: s}
	}
	return method, pattern, nil
}
