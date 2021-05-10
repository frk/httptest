package httpdoc

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
)

// removeStutter attempts to remove the word, or part of the word, from the
// given text and returns the result.
func removeStutter(text, word string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	word = strings.ToLower(strings.TrimSpace(word))

	if words := strings.Split(word, " "); len(words) > 1 {
		// this supposes that the preceding word(s) are determinants
		// and/or qualifiers and the last word is the subject noun.
		word = words[len(words)-1]
	}

	// TODO(mkopriva): would be nice if this could handle singular/plural matches.
	// Even if only for the English language.
	// - https://en.wikipedia.org/wiki/English_plurals

	if re, err := regexp.Compile(`\b` + regexp.QuoteMeta(word) + `\b`); err == nil {
		if re.MatchString(text) {
			return re.ReplaceAllString(text, "")
		}
	}
	return text
}

var rxDeterminer = regexp.MustCompile(`^the-+|-+(?:an?(?:-+an?)*)?-`)

func slugFromString(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return '-'
	}, s)

	return strings.Trim(rxDeterminer.ReplaceAllString(s, "-"), "-")
}

// expects a "clean" path, no consecutive slashes, no placeholders, ...
func slugFromPath(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' {
			return '.'
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
