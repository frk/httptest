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

func pathFromTestGroup(tg *httptest.TestGroup, stripPrefix func(string) string) string {
	method, pattern := tg.E.Split()
	pattern = stripPrefix(pattern)

	// normalize
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

	// try to reduce stutter by dropping the first node of p2 iff the last
	// node of p1 is the same word (singular and plural included).

	last := p1
	if i := strings.LastIndexByte(last, '/'); i > -1 {
		last = last[i+1:]
	}

	first, sliceAt := p2, len(p2)
	if i := strings.IndexByte(first, '/'); i > -1 {
		first = first[:i]
		sliceAt = i + 1
	}

	if first == last || (first+"s") == last || first == (last+"s") {
		p2 = p2[sliceAt:]
	}

	return p1 + "/" + p2
}

func anchorJoin(a1, a2 string) string {
	a1, a2 = strings.TrimRight(a1, "."), strings.TrimLeft(a2, ".")

	// try to reduce stutter by dropping the first node of a2 iff the last
	// node of a1 is the same word (singular and plural included).

	last := a1
	if i := strings.LastIndexByte(last, '.'); i > -1 {
		last = last[i+1:]
	}
	if i := strings.LastIndexByte(last, '-'); i > -1 {
		last = last[i+1:]
	}

	first, sliceAt := a2, len(a2)
	if i := strings.IndexByte(first, '.'); i > -1 {
		first = first[:i]
		sliceAt = i + 1
	}
	if i := strings.IndexByte(first, '-'); i > -1 {
		first = first[:i]
		sliceAt = i + 1
	}

	if first == last || (first+"s") == last || first == (last+"s") {
		a2 = a2[sliceAt:]
	}

	return a1 + "." + a2
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
