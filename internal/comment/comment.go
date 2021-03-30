package comment

import (
	"strings"
	"unicode"
)

func ToText(comment []string) string {
	return strings.Join(comment, "\n")

	if len(comment) == 0 {
		return ""
	}

	var lines []string
	for _, c := range comment {
		switch c[1] {
		case '/':
			// remove "//" comment marker
			c = c[2:]
		case '*':
			// remove "/*" and "*/" markers
			c = reindent(c[2 : len(c)-2])
		}

		if len(c) > 0 && c[0] == ' ' {
			c = c[1:]
		}

		cl := strings.Split(c, "\n")
		for _, l := range cl {
			lines = append(lines, trimTrailingSpace(l))
		}
	}

	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	// remove last line if it's empty
	if n > 0 && lines[n-1] == "" {
		n--
	}
	lines = lines[0:n]

	return strings.Join(lines, "\n")
}

func reindent(text string) string {
	if len(text) > 0 && text[0] == ' ' {
		text = text[1:]
	}

	cip := 0 // common indent prefix length
	lines := strings.Split(text, "\n")
	for _, l := range lines {
		if il := indentLen(l); !isBlankString(l) && il != 0 && (il < cip || cip == 0) {
			cip = il
		}
	}

	if cip > 0 {
		var ll []string
		for _, l := range lines {
			if len(l) > cip && isBlankString(l[:cip]) {
				l = l[cip:]
			}
			ll = append(ll, l)
		}
		return strings.Join(ll, "\n")
	}

	return strings.Join(lines, "\n")
}

// indentLen returns the number of whitespace characters by which the given line is indented.
func indentLen(line string) (n int) {
	for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
		n += 1
	}
	return n
}

// isBlankString reports whether the given string is composed of whitespace characters only.
func isBlankString(s string) bool {
	n := len(s) - 1
	for n >= 0 && isWhitespace(s[n]) {
		n--
	}
	return n < 0
}

// is Whitespace reports whether the given byte is a whitespace character or not.
func isWhitespace(c byte) bool {
	return c == '\r' || c == '\n' || c == ' ' || c == '\t'
}

func trimTrailingSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}