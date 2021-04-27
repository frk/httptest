package httpdoc

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
)

var rxDeterminer = regexp.MustCompile(`^the-+|-+(?:an?(?:-+an?)*)?-`)

func slugFromString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return '-'
	}, s)
	s = rxDeterminer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// expects a "clean" path, no consecutive slashes, no placeholders, ...
func slugFromPath(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' {
			return '-'
		}
		return r
	}, strings.Trim(s, "/"))
}

var rxPlaceholder = regexp.MustCompile(`{[^{}]*}`) // to remove placeholders
var rxSlashes = regexp.MustCompile(`/{2,}`)        // to remove consecutive slashes
var rxVerbs = regexp.MustCompile(`^(?:list|read|retrieve|get|search|browse|select|` +
	`find|fetch|filter|create|new|insert|save|post|update|change|modify|replace|` +
	`patch|delete|remove|destroy|erase|purge)$`)

func pathFromTestGroup(tg *httptest.TestGroup) string {
	method, pattern := tg.E.Split()
	method, pattern = strings.Trim(method, "/-"), strings.Trim(pattern, "/-")
	method, pattern = strings.ToLower(method), strings.ToLower(pattern)

	pattern = rxPlaceholder.ReplaceAllString(pattern, "")
	pattern = rxSlashes.ReplaceAllString(pattern, "/")
	pattern = strings.Trim(pattern, "/")

	verb := strings.ToLower(strings.TrimSpace(tg.Desc))
	if i := strings.IndexByte(verb, ' '); i > -1 {
		verb = verb[:i]
	}
	if !rxVerbs.MatchString(verb) {
		// fallback to using the endpoint's method (HTTP verb)
		verb = method
	}

	return "/" + pattern + "/" + verb
}

func pathJoin(p1, p2 string) string {
	p1, p2 = strings.TrimRight(p1, "/"), strings.TrimLeft(p2, "/")

	// A mediocre attempt to "merge" last path segment of p1 with the first
	// path segment of p2 *if* they represent the same word (singular and
	// plural can be mixed).
	//
	// p1="/users/events" p2="/events/read" => "/users/events/read"
	// p1="/users/events" p2="/event/read" => "/users/events/read"
	// p1="/users/event" p2="/events/read" => "/users/event/read"
	//
	// NOTE(mkopriva): Undecided about this approach for generating article
	// paths... it might be ditched later on, which is why at the moment the
	// implementation is as it is. If it's decided that the approach stays
	// a more robust implementation would be good.
	p1s, p2s := strings.Split(p1, "/"), strings.Split(p2, "/")
	if len(p1s) > 0 && len(p2s) > 0 {
		p1e, p2e := p1s[len(p1s)-1], p2s[0]
		if p1e == p2e || (p1e+"s") == p2e || p1e == (p2e+"s") {
			p2 = strings.Join(p2s[1:], "/")
		}
	}

	return p1 + "/" + p2
}

// removes scheme from url, e.g. "https://example.com" becomes "example.com".
func trimURLScheme(url string) string {
	if i := strings.Index(url, "://"); i >= 0 {
		return url[i+3:]
	}
	return url
}

// adapted from net/http.Header
type headerSorter struct {
	items []page.HeaderItem
}

func (s *headerSorter) Len() int           { return len(s.items) }
func (s *headerSorter) Swap(i, j int)      { s.items[i], s.items[j] = s.items[j], s.items[i] }
func (s *headerSorter) Less(i, j int) bool { return s.items[i].Key < s.items[j].Key }
