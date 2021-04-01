package comment

import (
	"bytes"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ToHTML(comment []string) (string, error) {
	text := removeSlashes(comment)
	if len(text) == 0 {
		return "", nil
	}

	root := parsedoc(text)
	buf := bytes.NewBuffer(nil)
	if err := root.write(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func removeSlashes(comment []string) string {
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

// itemType identifies the type of lex items.
type itemType int

const (
	itemEOF itemType = iota
	itemParaStart
	itemText
	itemRaw
	itemStar
	itemStar2
	itemURL
	itemNamedURL
	itemIdent

	eof = -1 // The eof rune signals the graceful end of input.
)

// item represents a token returned from the lexer.
type item struct {
	typ itemType
	val string
	pos int
}

type stateFn func(*lexer) stateFn

// NOTE(mkopriva): The lexer implementation is a modified copy of Go's lexer from:
// go/src/text/template/parse/lex.go (24a088d)
//
// lexer holds the state of the scanner.
type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan item
}

// lex initializes and runs a lexer with the given input in a goroutine.
func lex(input string) *lexer {
	lx := &lexer{
		input: input,
		items: make(chan item),
	}
	go lx.run()
	return lx
}

// run runs the state machine for the lexer.
func (lx *lexer) run() {
	for next := lexParaStart; next != nil; {
		next = next(lx)
	}
	close(lx.items)
}

// next returns the next rune in the input.
func (lx *lexer) next() (r rune) {
	if lx.pos >= len(lx.input) {
		lx.width = 0
		return eof
	}
	r, lx.width = utf8.DecodeRuneInString(lx.input[lx.pos:])
	lx.pos += lx.width
	return r
}

// prev unreads the last rune read by next and returns it.
func (lx *lexer) prev() (r rune) {
	if lx.pos <= 0 {
		lx.width = 0
		return eof
	}
	r, lx.width = utf8.DecodeLastRuneInString(lx.input[:lx.pos])
	lx.pos -= lx.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (lx *lexer) peek() (r rune) {
	r = lx.next()
	lx.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (lx *lexer) backup() {
	lx.pos -= lx.width
}

// ignore skips over the pending input before this point.
func (lx *lexer) ignore() {
	lx.start = lx.pos
}

// emit passes a item back to the client.
func (lx *lexer) emit(typ itemType) {
	lx.items <- item{typ: typ, val: lx.input[lx.start:lx.pos], pos: lx.start}
	lx.start = lx.pos
}

// lexParaStart scans the start of a paragraph. The position of the scanner is known
// to be at the beginning of the input or at the end of a previous paragraph or indented block.
func lexParaStart(lx *lexer) (fn stateFn) {
	if lx.peek() == eof {
		return nil
	}
	lx.emit(itemParaStart)
	return lexText
}

const markerChars = "`*:[\n\r"

// lexText scans regular text up to one of the marker characters that mark non-text items.
// FIXME(mkopriva): seems to blow up when a comment in the form "NOTE:TEMPORARY:" is encountered.
func lexText(lx *lexer) (fn stateFn) {
	r := lx.next()
	for !strings.ContainsRune(markerChars, r) && r != eof {
		r = lx.next()
	}

	if r != eof {
		lx.backup()
	}
	if lx.start < lx.pos && r != ':' {
		lx.emit(itemText)
	}
	return lexDelim
}

// lexDelim scans the next character and returns the appropriate stateFn. The
// position of the scanner is known to be at a character that marks a non-text item.
func lexDelim(lx *lexer) (fn stateFn) {
	switch lx.next() {
	case eof:
		lx.emit(itemEOF)
		return nil
	case '`':
		return lexRaw
	case '*':
		return lexStar
	case ':':
		return lexURL
	case '[':
		return lexNamedURL
	case '\n', '\r':
		switch lx.peek() {
		case ' ', '\t':
			lx.ignore()
			return lexIndent
		case '\n', '\r':
			lx.next()
			lx.ignore()
			return lexParaStart
		}
	}
	return lexText
}

// lexRaw scans a raw string. The initial backtick is already scanned.
func lexRaw(lx *lexer) stateFn {
	for r := lx.next(); r != eof; r = lx.next() {
		if isNewLine(r) && isBlank(lx.peek()) {
			break
		}
		if r == '`' {
			lx.emit(itemRaw)
			return lexText
		}
	}

	lx.pos = lx.start + 1
	return lexText
}

// lexStar scans a single or double asterisk delimiter. The initial asterisk is already scanned.
func lexStar(lx *lexer) stateFn {
	typ := itemStar
	if lx.peek() == '*' {
		lx.next()
		typ = itemStar2
	}
	lx.emit(typ)

	if lx.peek() == eof {
		lx.emit(itemEOF)
		return nil
	}
	return lexText
}

const (
	// Regexp for URLs (taken from Go's src/go/doc/comment.go)
	protocol = `https?|ftp|file|gopher|mailto|news|nntp|telnet|wais|prospero`
	hostPart = `[a-zA-Z0-9_@\-]+`
	filePart = `[a-zA-Z0-9_?%#~&/\-+=()]+`
	urlRx    = `(?:` + protocol + `)://` +
		hostPart + `(?:[.:]` + hostPart + `)*/?` +
		filePart + `(?:[:.,;]` + filePart + `)*`
)

var (
	urlLeftRE    = regexp.MustCompile(`^` + urlRx)
	protoRightRE = regexp.MustCompile(`(?:^|\b)(` + protocol + `):$`)
)

// lexURL scans a URL. The last read rune is known to be a colon which could
// possibly be following a "scheme" of a valid URL. If the colon's preceding and
// following inputs do not, together, match a valid URL the lexer will continue
// scanning the input from where it left off.
func lexURL(lx *lexer) stateFn {
	if plen := len(protoRightRE.FindString(lx.input[:lx.pos])); plen > 0 {
		if ulen := len(urlLeftRE.FindString(lx.input[lx.pos-plen:])); ulen > 0 {
			lx.pos -= plen
			if lx.start < lx.pos {
				lx.emit(itemText)
			}

			lx.pos += ulen
			lx.emit(itemURL)
			return lexText
		}
	}

	lx.pos = lx.pos + 1
	return lexText
}

var urlNamedRE = regexp.MustCompile(`\[[a-zA-Z0-9_@\- ]+\]\(` + urlRx + `\)`)

// lexNamedURL scans a named URL. The last read rune is known to be '['.
func lexNamedURL(lx *lexer) stateFn {
	pos := lx.pos - 1
	if ln := len(urlNamedRE.FindString(lx.input[pos:])); ln > 0 {
		lx.pos = pos + ln
		lx.emit(itemNamedURL)
	}
	return lexText
}

// lexIndent scans an indented block of text. The position of the scanner is known
// to be at the first indentation character, ' ' or '\t'.
func lexIndent(lx *lexer) stateFn {
	for r := lx.next(); r != eof; r = lx.next() {
		if isNewLine(r) {
			r = lx.peek()
			if !isNewLine(r) && !isSpace(r) {
				break
			}
		}
	}
	lx.emit(itemIdent)
	return lexParaStart
}

// nodeType identifies the type of parsed nodes.
type nodeType int

const (
	nodeRoot nodeType = iota
	nodeText
	nodePara
	nodePre
	nodeCode
	nodeEm
	nodeStrong
	nodeAnchor
)

// htmlTags returns the nodeType's HTML start and end tags.
func (typ nodeType) htmlTags() (start, end []byte) {
	switch typ {
	case nodePara:
		return []byte("<p>"), []byte("</p>\n")
	case nodePre:
		return []byte("<pre><code>"), []byte("</code></pre>\n")
	case nodeCode:
		return []byte("<code>"), []byte("</code>")
	case nodeEm:
		return []byte("<em>"), []byte("</em>")
	case nodeStrong:
		return []byte("<strong>"), []byte("</strong>")
	case nodeAnchor:
		return []byte(`<a href="`), []byte(`">`)
	}
	return nil, nil
}

// ..
func (typ nodeType) marker() string {
	switch typ {
	case nodeEm:
		return "*"
	case nodeStrong:
		return "**"
	}
	return ""
}

// node represents an HTML node.
type node struct {
	typ      nodeType
	pos      int // the position of the node's token in the input
	data     string
	href     string
	children []*node
}

// for debugging
func (n *node) String() string {
	s := "{typ:" + strconv.Itoa(int(n.typ))
	s += ",pos:" + strconv.Itoa(n.pos)
	s += ",data:'" + n.data + "'"
	s += ",href:'" + n.href + "'"
	s += ",children:["
	for _, c := range n.children {
		s += c.String() + ","
	}
	if i := len(s) - 1; s[i] == ',' {
		s = s[:i]
	}
	return s + "]}"
}

// add allocates and adds a node with the given parameters to the children slice.
func (n *node) add(typ nodeType, pos int, data string) *node {
	c := &node{typ: typ, pos: pos, data: data}
	n.children = append(n.children, c)
	return c
}

func (n *node) _add(c *node) *node {
	n.children = append(n.children, c)
	return c
}

// wirte writes the html representation of the node and its children to the buffer.
func (n *node) write(b *bytes.Buffer) (err error) {
	if n.typ == nodeText {
		_, err = b.WriteString(n.data)
		return err
	}

	start, end := n.typ.htmlTags()
	if len(start) > 0 {
		if _, err = b.Write(start); err != nil {
			return err
		}
	}

	// if the node is an anchor, first write the href attribute into the start tag.
	if n.typ == nodeAnchor {
		if _, err = b.WriteString(n.href); err != nil {
			return err
		}
		// end is `">`, close the start tag with it
		if _, err = b.Write(end); err != nil {
			return err
		}
		end = []byte("</a>") // reset the end tag and continue
	}

	if len(n.data) > 0 {
		if _, err = b.WriteString(n.data); err != nil {
			return err
		}
	}
	for _, c := range n.children {
		if err = c.write(b); err != nil {
			return err
		}
	}

	if len(end) > 0 {
		if _, err = b.Write(end); err != nil {
			return err
		}
	}
	return err
}

type docparser struct {
	lx    *lexer
	tree  *node
	stack []*node // tracks the nestedness of nodes
}

// parsedoc parses the given text into a node tree.
func parsedoc(text string) *node {
	n := &node{typ: nodeRoot}
	p := &docparser{
		lx:    lex(text),
		tree:  n,
		stack: []*node{n},
	}

	for {
		it := <-p.lx.items
		if it.typ == itemEOF {
			if p.within(nodePara) && !p.istop(nodePara) {
				p.flatten(nodePara)
			}
			break
		}
		p.parse(it)
	}
	return n
}

// push appends the given node to the stack.
func (p *docparser) push(n *node) {
	p.stack = append(p.stack, n)
}

// pop removes the topmost node from the stack.
func (p *docparser) pop() *node {
	if i := len(p.stack) - 1; i >= 0 {
		n := p.stack[i]
		p.stack = p.stack[:i]
		return n
	}
	return nil
}

// pop2 removes the topmost node of the given type from the stack. If the given
// typ is -1 the result will be the same as calling pop, if the given typ is not
// present in the stack the stack will remain unchanged and nil will be returned.
func (p *docparser) pop2(typ nodeType) *node {
	if typ == -1 {
		return p.pop()
	}
	if i := p.find(typ); i >= 0 {
		n := p.stack[i]
		p.stack = p.stack[:i]
		return n
	}
	return nil
}

// top returns the node at the top of the stack.
func (p *docparser) top() *node {
	if i := len(p.stack) - 1; i >= 0 {
		return p.stack[i]
	}
	return nil
}

// istop reports whether the node at the top of the stack is of the given type.
func (p *docparser) istop(typ nodeType) bool {
	return p.top().typ == typ
}

// find returns the stack index of the last node with the given type.
func (p *docparser) find(typ nodeType) int {
	for i := len(p.stack) - 1; i >= 0; i-- {
		if p.stack[i].typ == typ {
			return i
		}
	}
	return -1
}

// within reports whether a node of the given type is present in the stack.
func (p *docparser) within(typ nodeType) bool {
	return p.find(typ) >= 0
}

// flatten joins all leaf nodes that are the descendants of the nearest node with
// the given nodeType, transforming any non-leaf descendant nodes into leaf nodes.
func (p *docparser) flatten(typ nodeType) (n *node) {
	for p.top().typ != typ {
		if n = p.pop(); n != nil {
			if i := len(p.stack) - 1; i >= 0 {
				if j := len(p.stack[i].children) - 1; j >= 0 {
					p.stack[i].children = p.stack[i].children[:j]
				}
			}

			top := p.top()
			top.add(nodeText, n.pos, n.typ.marker())
			top.children = append(top.children, n.children...)
		}
	}
	return n
}

// parse, based on the given item, adds a new node to the tree.
func (p *docparser) parse(it item) int {
	switch it.typ {
	case itemParaStart:
		if !p.istop(nodeRoot) && !p.istop(nodePara) {
			p.flatten(nodePara)
		}
		if p.istop(nodePara) {
			p.pop()
		}
		p.push(p.top().add(nodePara, it.pos, ""))
	case itemIdent:
		p.pop2(nodePara)
		data := trimTrailingSpace(it.val)
		data = template.HTMLEscapeString(reindent(data))
		p.top().add(nodePre, it.pos, data)
	case itemText:
		p.top().add(nodeText, it.pos, it.val)
	case itemRaw:
		data := it.val[1 : len(it.val)-1]
		data = template.HTMLEscapeString(data)
		p.top().add(nodeCode, it.pos, data)
	case itemStar:
		if p.within(nodeEm) {
			if !p.istop(nodeEm) {
				p.flatten(nodeEm)
			}
			p.pop2(nodeEm)
		} else {
			p.push(p.top().add(nodeEm, it.pos, ""))
		}
	case itemStar2:
		if p.within(nodeStrong) {
			if !p.istop(nodeStrong) {
				p.flatten(nodeStrong)
			}
			p.pop2(nodeStrong)
		} else {
			p.push(p.top().add(nodeStrong, it.pos, ""))
		}
	case itemURL:
		p.top()._add(&node{typ: nodeAnchor, pos: it.pos, data: it.val, href: it.val})
	case itemNamedURL:
		name, url := parseNamedURL(it.val)
		p.top()._add(&node{typ: nodeAnchor, pos: it.pos, data: name, href: url})
	}
	return -1
}

// is Whitespace reports whether the given byte is a whitespace character or not.
func isWhitespace(c byte) bool {
	return c == '\r' || c == '\n' || c == ' ' || c == '\t'
}

// parseNamedURL parses the given string as named URL in the form of "[name](url)"
// returning it's name and url separately.
func parseNamedURL(s string) (name, url string) {
	s = s[1:] // drop [
	if i := strings.IndexByte(s, ']'); i > 0 {
		name = s[:i]
		url = s[i+2 : len(s)-1] // +2 to drop "]("
	}
	return name, url
}

// isNewLine reports whether the given rune is a new line character or not.
func isNewLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isSpace reports whether the given rune is a space or a tab.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isBlank reports whether the given rune is a whitespace character.
func isBlank(r rune) bool {
	return isSpace(r) || isNewLine(r)
}

func trimTrailingSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}