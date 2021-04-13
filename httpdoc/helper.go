package httpdoc

import (
	"strings"

	"github.com/frk/httptest/internal/page"
)

// adapted from net/http.Header
type headerSorter struct {
	items []page.HeaderItem
}

func (s *headerSorter) Len() int           { return len(s.items) }
func (s *headerSorter) Swap(i, j int)      { s.items[i], s.items[j] = s.items[j], s.items[i] }
func (s *headerSorter) Less(i, j int) bool { return s.items[i].Key < s.items[j].Key }

// removes scheme from url, e.g. "https://example.com" becomes "example.com".
func trimURLScheme(url string) string {
	if i := strings.Index(url, "://"); i >= 0 {
		return url[i+3:]
	}
	return url
}
