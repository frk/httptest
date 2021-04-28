package httpdoc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/godoc"
	"github.com/frk/httptest/internal/markup"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/program"
	"github.com/frk/httptest/internal/types"
	"github.com/frk/tagutil"
)

// holds the state of the build process
type build struct {
	Config
	// the input
	dir ArticleDirectory
	// source code info
	src *types.Source
	// the page being built
	page page.Page
	// the program being built
	prog program.Program
	// the sidebar list currently being built, or nil
	sbls *page.SidebarList

	ids   map[string]int              // set of already generated ids
	csIds map[page.CodeSnippet]string // cache of CodeSnippet specific ids

	// Keeps track of the parent-relation between an *Article and its parent
	// *Article or between an *httptest.TestGroups and its parent *Article.
	// The map is populated during the first pass through the input.
	parents map[interface{}]*Article
	// ...
	objkeys map[interface{}]objkeys
	// keep track of already generated values to ensure uniquenes
	slugs, paths, anchors map[string]int
}

// set of properties that uniquely identify an object across the documentation.
type objkeys struct {
	slug   string // a slug of an article
	path   string // the url path of an article
	anchor string // the id of an html element (not necessarily an article)
}

func (c *build) run() error {
	// load the source info
	src, err := types.Load(c.srcdir)
	if err != nil {
		return err
	}
	c.src = src

	// initialize build
	c.ids = make(map[string]int)
	c.csIds = make(map[page.CodeSnippet]string)

	////xxxxxxxxxxxxxx
	c.parents = make(map[interface{}]*Article)
	c.objkeys = make(map[interface{}]objkeys)
	c.slugs = make(map[string]int)
	c.paths = make(map[string]int)
	c.anchors = make(map[string]int)

	// ensure the configured paths and the paths generated later don't collide
	c.paths[c.RootPath] = 0
	c.paths[c.logoURL] = 0
	c.paths[c.SigninPath] = 0

	// build
	c.prepBuild()
	if err := c.buildSidebar(); err != nil {
		return err
	}
	if err := c.buildContent(); err != nil {
		return err
	}
	if err := c.buildProgram(); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Prep (first pass)
////////////////////////////////////////////////////////////////////////////////

func (c *build) prepBuild() {
	for _, g := range c.dir {
		c.prepArticles(g.Articles, nil)
	}
}

func (c *build) prepArticles(articles []*Article, parent *Article) {
	for _, a := range articles {
		c.parents[a] = parent
		c.objkeys[a] = c.newObjKeysForArticle(a, parent)
		c.prepTestGroups(a.TestGroups, a)
		c.prepArticles(a.SubArticles, a)
	}
}

func (c *build) prepTestGroups(tgs []*httptest.TestGroup, parent *Article) {
	for _, g := range tgs {
		c.parents[g] = parent
		c.objkeys[g] = c.newObjKeysForTestGroup(g, parent)
	}
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

func (c *build) buildSidebar() error {
	c.page.Title = c.PageTitle

	banner, err := c.newSidebarBanner()
	if err != nil {
		return err
	}
	c.page.Sidebar.Header.Banner = banner

	lists := []*page.SidebarList{}
	for _, g := range c.dir {
		if len(g.Articles) == 0 {
			continue
		}

		items, err := c.newSidebarItemsFromArticles(g.Articles, nil)
		if err != nil {
			return err
		}

		list := new(page.SidebarList)
		list.Title = g.Name
		list.Items = items
		lists = append(lists, list)
	}
	c.page.Sidebar.Lists = lists
	return nil
}

func (c *build) newSidebarBanner() (page.SidebarBanner, error) {
	if text := c.Config.SidebarBannerHTML; len(text) > 0 {
		banner := new(page.SidebarBannerHTML)
		banner.Text = text
		return banner, nil
	}

	banner := new(page.SidebarBannerTitle)
	banner.Text = c.PageTitle
	banner.URL = template.URL(c.RootPath)
	return banner, nil
}

var errNoArticleTitle = fmt.Errorf("httpdoc: Article.Title is required.")

func (c *build) newSidebarItemsFromArticles(articles []*Article, parent *Article) ([]*page.SidebarItem, error) {
	items := make([]*page.SidebarItem, len(articles))
	for i, a := range articles {
		if a.Title == "" {
			// title is required, fail if none was provided
			return nil, errNoArticleTitle
		}

		key := c.objkeys[a]
		item := new(page.SidebarItem)
		item.Text = a.Title
		item.Path = key.path
		item.Href = key.path + "#" + key.anchor

		if len(a.TestGroups) > 0 {
			items := c.newSidebarItemsFromTestGroups(a.TestGroups, a)
			item.SubItems = append(item.SubItems, items...)
		}
		if len(a.SubArticles) > 0 {
			items, err := c.newSidebarItemsFromArticles(a.SubArticles, a)
			if err != nil {
				return nil, err
			}
			item.SubItems = append(item.SubItems, items...)
		}

		items[i] = item
	}
	return items, nil
}

func (c *build) newSidebarItemsFromTestGroups(tgs []*httptest.TestGroup, parent *Article) (items []*page.SidebarItem) {
	for _, g := range tgs {
		// include the test group only if a decent description was extracted
		if desc := getTestGroupDesc(g); len(desc) > 0 {
			key := c.objkeys[g]
			item := new(page.SidebarItem)
			item.Text = desc
			item.Path = key.path
			item.Href = key.path + "#" + key.anchor
			items = append(items, item)
		}
	}
	return items
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

func (c *build) buildContent() error {
	for _, g := range c.dir {
		if len(g.Articles) == 0 {
			continue
		}

		if g.LoadExpanded {
			for _, a := range g.Articles {
				a.LoadExpanded = true
			}
		}

		list, err := c.newArticleElementListFromArticles(g.Articles, nil)
		if err != nil {
			return err
		}

		c.page.Content.Articles = append(c.page.Content.Articles, list...)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Program
////////////////////////////////////////////////////////////////////////////////

func (b *build) buildProgram() error {
	// package name
	b.prog.IsExecutable = !b.Config.OutputPackage
	if b.prog.IsExecutable {
		b.prog.PkgName = "main"
	} else {
		// TODO this needs to be converted to a valid name....
		b.prog.PkgName = b.Config.OutputName
	}

	// root path & listening port
	b.prog.RootPath = b.Config.RootPath
	b.prog.ListenAddr = ":" + strconv.Itoa(b.Config.TCPListenPort)

	b.prog.ValidPaths = make(map[string]string, 3+len(b.objkeys))
	b.prog.ValidPaths[b.RootPath] = ""
	b.prog.ValidPaths[b.logoURL] = ""
	b.prog.ValidPaths[b.SigninPath] = ""
	for _, k := range b.objkeys {
		b.prog.ValidPaths[k.path] = k.anchor
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Aritcle Element List
////////////////////////////////////////////////////////////////////////////////

func (c *build) newArticleElementListFromArticles(articles []*Article, parent *Article) ([]*page.ArticleElement, error) {
	list := make([]*page.ArticleElement, len(articles))
	for i, a := range articles {
		aElem, err := c.newArticleElementFromArticle(a, parent)
		if err != nil {
			return nil, err
		}

		if len(a.TestGroups) > 0 {
			section := c.newExampleEndpoints(a.TestGroups)
			aElem.Example.Sections = append(aElem.Example.Sections, section)

			list, err := c.newArticleElementListFromTestGroups(a.TestGroups, a)
			if err != nil {
				return nil, err
			}
			aElem.SubArticles = append(aElem.SubArticles, list...)
		}
		if len(a.SubArticles) > 0 {
			list, err := c.newArticleElementListFromArticles(a.SubArticles, a)
			if err != nil {
				return nil, err
			}
			aElem.SubArticles = append(aElem.SubArticles, list...)
		}

		list[i] = aElem
	}
	return list, nil
}

func (c *build) newArticleElementListFromTestGroups(tgs []*httptest.TestGroup, parent *Article) ([]*page.ArticleElement, error) {
	list := make([]*page.ArticleElement, len(tgs))
	for i, g := range tgs {
		aElem, err := c.newArticleElementFromTestGroup(g, parent)
		if err != nil {
			return nil, err
		}

		if len(g.Tests) > 0 {
			sections, err := c.newExampleSectionsFromTestGroup(g.Tests[0], g)
			if err != nil {
				return nil, err
			}
			aElem.Example.Sections = sections
		}

		list[i] = aElem
	}
	return list, nil
}

////////////////////////////////////////////////////////////////////////////////
// Aritcle Element
////////////////////////////////////////////////////////////////////////////////

func (c *build) newArticleElementFromArticle(a *Article, parent *Article) (*page.ArticleElement, error) {
	aElem := new(page.ArticleElement)
	aElem.Id = c.objkeys[a].anchor
	aElem.Href = c.objkeys[a].path
	aElem.Title = a.Title

	if a.Text != nil {
		text, err := c.newHTML(a.Text, nil)
		if err != nil {
			return nil, err
		}
		aElem.Text = text
	}

	if a.Code != nil {
		switch v := a.Code.(type) {
		case string, *os.File, HTMLer:
			text, err := c.newHTML(a.Code, nil)
			if err != nil {
				return nil, err
			}

			s := &page.ExampleText{Text: text}
			aElem.Example.Sections = append(aElem.Example.Sections, s)
		case Valuer, interface{}:
			s, err := c.newExampleObject(v, a.Type, a.Title)
			if err != nil {
				return nil, err
			}
			aElem.Example.Sections = append(aElem.Example.Sections, s)

			// Optionally generate a section of field docs to add
			// to the article's primary-column. If not struct, ignore.
			list, err := c.newFieldList(v, aElem, page.FIELD_LIST_OBJECT, false)
			if err != nil && err != errNotStructType {
				return nil, err
			} else if err == nil {
				s := new(page.ArticleFieldList)
				s.Title = "Fields"
				s.Lists = []*page.FieldList{list}
				aElem.Sections = append(aElem.Sections, s)
			}
		}
	}

	return aElem, nil
}

func (c *build) newArticleElementFromTestGroup(tg *httptest.TestGroup, parent *Article) (*page.ArticleElement, error) {
	aElem := new(page.ArticleElement)
	aElem.Id = c.objkeys[tg].anchor
	aElem.Href = c.objkeys[tg].path
	aElem.Title = getTestGroupDesc(tg)

	if tg.DocA != nil {
		var decl types.TypeDecl
		text, err := c.newHTML(tg.DocA, &decl)
		if err != nil {
			return nil, err
		}
		aElem.Text = text

		// With TestGroup.DocA, if the value was a named type, assume that
		// that type is directly related to the endpoint, e.g. it could be
		// the handler registered to handle the endpoint, therefore grab
		// its source code position and generate a source link for it.
		if len(decl.Name) > 0 && c.SourceURL != nil {
			href := c.SourceURL(decl.Pos.Filename, decl.Pos.Line)
			text := decl.PkgPath + "." + decl.Name

			aElem.SourceLink = new(page.SourceLink)
			aElem.SourceLink.Href = href
			aElem.SourceLink.Text = text
		}
	}

	// The input/output docs, only the 0th httptest.Test in the group is used,
	// it is up to the user to make sure that if tests are presents then it is
	// the 0th one that is representative.
	if len(tg.Tests) > 0 {
		t := tg.Tests[0]

		// auth info
		switch v := t.Request.Auth.(type) {
		case HTMLer, Valuer:
			text, err := c.newHTML(v, nil)
			if err != nil {
				return nil, err
			}
			section := &page.ArticleAuthInfo{Text: text}
			section.Title = "Auth"
			aElem.Sections = append(aElem.Sections, section)
		}

		// input field lists
		var inputFields []*page.FieldList
		if v, ok := t.Request.Header.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, page.FIELD_LIST_HEADER, true)
			if err != nil {
				return nil, err
			}
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Params.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, page.FIELD_LIST_PATH, true)
			if err != nil {
				return nil, err
			}
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Query.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, page.FIELD_LIST_QUERY, true)
			if err != nil {
				return nil, err
			}
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Body.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, page.FIELD_LIST_BODY, true)
			if err != nil {
				return nil, err
			}
			inputFields = append(inputFields, list)
		}
		if len(inputFields) > 0 {
			section := new(page.ArticleFieldList)
			section.Title = "INPUT"
			section.Lists = inputFields
			aElem.Sections = append(aElem.Sections, section)
		}

		// NOTE(mkopriva): for now create an output section only for
		// endpoints where the test request doesn't have a body.
		// Consider, however, to change this in the future, maybe by
		// using a tabbed view of input and output for every endpoint.
		if v, ok := t.Request.Body.(Valuer); !ok || v == nil {
			var outputFields []*page.FieldList
			if v, ok := t.Response.Header.(Valuer); ok && v != nil {
				list, err := c.newFieldList(v, aElem, page.FIELD_LIST_HEADER, false)
				if err != nil {
					return nil, err
				}
				outputFields = append(outputFields, list)
			}
			if v, ok := t.Response.Body.(Valuer); ok && v != nil {
				list, err := c.newFieldList(v, aElem, page.FIELD_LIST_BODY, false)
				if err != nil {
					return nil, err
				}
				outputFields = append(outputFields, list)
			}
			if len(outputFields) > 0 {
				section := new(page.ArticleFieldList)
				section.Title = "OUTPUT"
				section.Lists = outputFields
				aElem.Sections = append(aElem.Sections, section)
			}
		}
	}

	if tg.DocB != nil {
		text, err := c.newHTML(tg.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ArticleText{Text: text}
		aElem.Sections = append(aElem.Sections, section)
	}
	return aElem, nil
}

////////////////////////////////////////////////////////////////////////////////
// Example Sections
////////////////////////////////////////////////////////////////////////////////

func (c *build) newExampleSectionsFromTestGroup(t *httptest.Test, tg *httptest.TestGroup) (sections []page.ExampleSection, err error) {
	if t.DocA != nil {
		text, err := c.newHTML(t.DocA, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleText{Text: text}
		sections = append(sections, section)
	}

	reqSection, err := c.newExampleRequest(t.Request, tg)
	if err != nil {
		return nil, err
	}
	sections = append(sections, reqSection)

	respSection, err := c.newExampleResponse(t.Response, tg)
	if err != nil {
		return nil, err
	}
	sections = append(sections, respSection)

	if t.DocB != nil {
		text, err := c.newHTML(t.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleText{Text: text}
		sections = append(sections, section)
	}
	return sections, nil
}

func (c *build) newExampleEndpoints(tgs []*httptest.TestGroup) *page.ExampleEndpoints {
	section := new(page.ExampleEndpoints)
	section.Title = "ENDPOINTS"

	for _, tg := range tgs {
		method, pattern := tg.E.Split()

		item := new(page.EndpointItem)
		item.Href = c.objkeys[tg].path
		item.Method = method
		item.Pattern = pattern
		item.Tooltip = getTestGroupDesc(tg)
		section.Endpoints = append(section.Endpoints, item)
	}

	return section
}

func (c *build) newExampleObject(obj interface{}, mediatype, title string) (*page.ExampleObject, error) {
	// if the object is a Valuer then get the underlying value
	if v, ok := obj.(Valuer); ok {
		val, err := v.Value()
		if err != nil {
			return nil, err
		}
		obj = val
	}

	// default if no type was provided
	if mediatype == "" {
		mediatype = "application/json"
	}

	text, err := marshalValue(obj, mediatype, true)
	if err != nil {
		return nil, err
	}

	xo := new(page.ExampleObject)
	xo.Title = title
	xo.Lang = getLangFromMediaType(mediatype)
	xo.Code = template.HTML(text)
	return xo, nil
}

func (c *build) newExampleRequest(req httptest.Request, tg *httptest.TestGroup) (*page.ExampleRequest, error) {
	method, pattern := tg.E.Split()
	xr := new(page.ExampleRequest)
	xr.Title = "REQUEST"
	xr.Method = method
	xr.Pattern = pattern

	xr.Options = make([]*page.SelectOption, len(c.SnippetTypes))
	xr.Snippets = make([]*page.CodeSnippetElement, len(c.SnippetTypes))
	for i, st := range c.SnippetTypes {
		var snip page.CodeSnippet
		var name, lang string
		var numlines int

		switch st { // TODO snippets for vanilla-js, Go
		case SNIPP_HTTP:
			cs, nl, err := c.newCodeSnippetHTTP(req, tg)
			if err != nil {
				return nil, err
			}
			snip = cs
			name, lang = cs.Name(), cs.Lang()
			numlines = nl
		case SNIPP_CURL:
			cs, nl, err := c.newCodeSnippetCURL(req, tg)
			if err != nil {
				return nil, err
			}
			snip = cs
			name, lang = cs.Name(), cs.Lang()
			numlines = nl
		}

		elem := new(page.CodeSnippetElement)
		elem.Id = c.getIdForCodeSnippet(snip, tg, lang)
		elem.Show = (i == 0)
		elem.Lang = lang
		elem.Snippet = snip
		elem.NumLines = numlines
		xr.Snippets[i] = elem

		opt := new(page.SelectOption)
		opt.Text = name
		opt.Value = lang
		opt.DataId = elem.Id
		opt.Selected = elem.Show
		xr.Options[i] = opt
	}

	return xr, nil
}

func (c *build) newExampleResponse(resp httptest.Response, tg *httptest.TestGroup) (*page.ExampleResponse, error) {
	section := new(page.ExampleResponse)
	section.Title = "RESPONSE"
	section.Status = resp.StatusCode

	if resp.Header != nil {
		for key, values := range resp.Header.GetHeader() {
			for _, val := range values {
				item := page.HeaderItem{}
				item.Key = key
				item.Value = val
				section.Header = append(section.Header, item)
			}
		}
		sort.Sort(&headerSorter{section.Header})
	}

	if resp.Body != nil {
		text, mediatype, _, err := marshalBody(resp.Body, true)
		if err != nil {
			return nil, err
		}

		section.Lang = getLangFromMediaType(mediatype)
		section.Code = template.HTML(text)

	}

	return section, nil
}

////////////////////////////////////////////////////////////////////////////////
// Code Snippets
////////////////////////////////////////////////////////////////////////////////

func (c *build) newCodeSnippetHTTP(req httptest.Request, tg *httptest.TestGroup) (*page.CodeSnippetHTTP, int, error) {
	cs := new(page.CodeSnippetHTTP)
	numlines := 0
	method, pattern := tg.E.Split()

	// The message start-line
	cs.Method = method
	cs.RequestURI = getRequestPath(req, pattern)
	cs.HTTPVersion = "HTTP/1.1"
	numlines += 1

	// the message body
	if req.Body != nil {
		text, _, nl, err := marshalBody(req.Body, false)
		if err != nil {
			return nil, 0, err
		}
		cs.Body = template.HTML(text)
		numlines += nl + 1 // +1 for the new line that separates the headers from the body
	}

	// the message headers
	cs.Headers = []page.HeaderItem{{Key: "Host", Value: trimURLScheme(c.ExampleHost)}}
	if req.Body != nil {
		h1 := page.HeaderItem{Key: "Content-Type", Value: req.Body.Type()}
		h2 := page.HeaderItem{Key: "Content-Length", Value: strconv.Itoa(len(cs.Body))}
		cs.Headers = append(cs.Headers, h1, h2)
	}
	if req.Header != nil {
		for key, values := range req.Header.GetHeader() {
			// if the body is present the content type and length
			// headers were already set above, so skip them here.
			if req.Body != nil && (key == "Content-Type" || key == "Content-Length") {
				continue
			}

			for _, val := range values {
				cs.Headers = append(cs.Headers, page.HeaderItem{Key: key, Value: val})
			}
		}
	}
	numlines += len(cs.Headers)

	// finish off by sorting the message headers, if any are present
	if len(cs.Headers) > 0 {
		sort.Sort(&headerSorter{cs.Headers})
	}
	return cs, numlines, nil
}

func (c *build) newCodeSnippetCURL(req httptest.Request, tg *httptest.TestGroup) (*page.CodeSnippetCURL, int, error) {
	cs := new(page.CodeSnippetCURL)
	numlines := 0
	method, pattern := tg.E.Split()

	// the target URL
	cs.URL = c.ExampleHost + getRequestPath(req, pattern)
	// the -X option
	cs.X = method
	numlines += 1

	// the -H options
	if req.Body != nil {
		cs.H = append(cs.H, fmt.Sprintf("Content-Type: %s", req.Body.Type()))
		numlines += 1
	}
	if req.Header != nil {
		for key, values := range req.Header.GetHeader() {
			// if the body is present the content type header was
			// already set above, so skip it here
			if req.Body != nil && key == "Content-Type" {
				continue
			}
			for _, val := range values {
				cs.H = append(cs.H, fmt.Sprintf("%s: %s", key, val))
				numlines += 1
			}
		}
	}

	// the -d/--data options
	if req.Body != nil {
		text, _, nl, err := marshalBody(req.Body, false)
		if err != nil {
			return nil, 0, err
		}
		cs.Data = append(cs.Data, page.CURLDataText(text))
		numlines += nl
	}
	return cs, numlines, nil
}

////////////////////////////////////////////////////////////////////////////////
// Misc.
////////////////////////////////////////////////////////////////////////////////

var errNotNamedType = fmt.Errorf("httpdoc: type is not named")

func (c *build) newHTML(value interface{}, decl *types.TypeDecl) (_ template.HTML, err error) {
	var text string

	switch v := value.(type) {
	case string:
		text = v
	case *os.File:
		out, err := ioutil.ReadAll(v)
		if err != nil {
			return "", err
		}
		text = string(out)
	case HTMLer:
		out, err := v.HTML()
		if err != nil {
			return "", err
		}
		text = string(out)
	case Valuer, interface{}:
		if vv, ok := value.(Valuer); ok && vv != nil {
			if value, err = vv.Value(); err != nil {
				return "", err
			}
		}

		// Extract documentation from the value's type.
		d := c.src.TypeDeclOf(value)
		if d == nil {
			return "", errNotNamedType
		}
		if text, err = godoc.ToHTML(d.Doc); err != nil {
			return "", err
		}

		if decl != nil { // retain type declaration?
			*decl = *d
		}
	}

	return template.HTML(text), nil
}

var errNotStructType = fmt.Errorf("httpdoc: type is not a struct")

func (c *build) newFieldList(value interface{}, aElem *page.ArticleElement, class page.FieldListClass, isInput bool) (list *page.FieldList, err error) {
	// if this is a Valuer then get the underlying value
	if v, ok := value.(Valuer); ok {
		if value, err = v.Value(); err != nil {
			return nil, err
		}
	}

	typ := c.src.TypeOf(value)
	if typ = getNearestStructType(typ); typ == nil {
		return nil, errNotStructType
	}

	var title, idpfx string
	switch class {
	case page.FIELD_LIST_PATH:
		title = "Path"
		idpfx = "path."
	case page.FIELD_LIST_QUERY:
		title = "Query"
		idpfx = "query."
	case page.FIELD_LIST_HEADER:
		title = "Header"
		idpfx = "header."
	case page.FIELD_LIST_BODY:
		title = "Body"
		idpfx = "body."
	case page.FIELD_LIST_OBJECT:
		title = ""
		idpfx = "obj."
	}

	list, err = c._newFieldList(typ, aElem, class, isInput, idpfx, nil)
	if err != nil {
		return nil, err
	}

	list.Title = title
	return list, nil
}

func (c *build) _newFieldList(typ *types.Type, aElem *page.ArticleElement, class page.FieldListClass, isInput bool, idpfx string, path []string) (*page.FieldList, error) {
	list := new(page.FieldList)

	tagKey := c.FieldNameTag
	for _, f := range typ.Fields {
		tag := tagutil.New(f.Tag)
		if tag.Contains(tagKey, "-") || tag.Contains("doc", "-") { // skip field?
			continue
		} else if !f.IsExported && (!f.IsEmbedded || !f.Type.CanSelectFields()) {
			// Nothing to do here if: The field is unexported and not embedded,
			// or, unexported, embedded, but has not fields to promote.
			continue
		}

		// If this is an embedded field that promotes fields to the parent
		// then unpack those fields directly, rather than as sub-fields.
		if f.IsEmbedded && f.Type.CanSelectFields() {
			if stype := getNearestStructType(f.Type); stype != nil && len(stype.Fields) > 0 {
				subList, err := c._newFieldList(stype, aElem, class, isInput, idpfx, path)
				if err != nil {
					return nil, err
				}
				list.Items = append(list.Items, subList.Items...)
			}
			continue
		}

		sf, ok := typ.ReflectType.FieldByName(f.Name)
		if !ok {
			// shouldn't happen
			panic(fmt.Sprintf("httpdoc: reflect.Type.FieldByName(%q) failed.", f.Name))
		}

		item := new(page.FieldItem)

		// the field's name
		item.Name = f.Name
		if name := tag.First(tagKey); name != "" {
			item.Name = name
		}

		// the field's path
		if len(path) > 0 {
			item.Path = strings.Join(path, ".") + "."
		}

		// the field's id
		item.Id = aElem.Id + "." + idpfx + item.Path + item.Name
		// the field's anchor
		item.Href = "#" + item.Id

		// "trim" pointers
		ftype := f.Type
		for ftype.Kind == types.KindPtr {
			ftype = ftype.Elem
		}

		// the field's type
		item.Type = ftype.String()
		if c.FieldType != nil {
			if name, ok := c.FieldType(sf); len(name) > 0 || ok {
				item.Type = name
			}
		}

		// the field's documentation
		if len(f.Doc) > 0 {
			text, err := godoc.ToHTML(f.Doc)
			if err != nil {
				return nil, err
			}
			item.Text = template.HTML(text)
		}

		// the field type's enum values
		if len(ftype.Values) > 0 {
			enumList, err := c.newEnumList(ftype)
			if err != nil {
				return nil, err
			}
			item.EnumList = enumList
		}

		// the field's sub fields
		if stype := getNearestStructType(ftype); stype != nil && len(stype.Fields) > 0 {
			itemName := item.Name
			if ftype.Kind == types.KindSlice || ftype.Kind == types.KindArray {
				itemName += "[]"
			}

			subList, err := c._newFieldList(stype, aElem, class, isInput, idpfx, append(path, itemName))
			if err != nil {
				return nil, err
			}
			item.SubFields = subList.Items
		}

		// the field's source link
		if c.SourceURL != nil {
			item.SourceLink = new(page.SourceLink)
			item.SourceLink.Href = c.SourceURL(f.Pos.Filename, f.Pos.Line)
		}

		// for input fields additionally generate validation info
		if isInput {
			// the field's setting
			if label, text, ok := c.FieldSetting(sf, typ.ReflectType); ok {
				item.SettingLabel = label
				item.SettingText = text
			}

			// the field's validation
			if c.FieldValidation != nil {
				item.Validation = c.FieldValidation(sf, typ.ReflectType)
			}
		}

		list.Items = append(list.Items, item)
	}

	return list, nil
}

func (c *build) newEnumList(typ *types.Type) (*page.EnumList, error) {
	list := new(page.EnumList)
	list.Title = "Possible enum values"
	list.Items = make([]*page.EnumItem, len(typ.Values))

	for i, v := range typ.Values {
		item := new(page.EnumItem)
		item.Value = v.Value

		// the internal/types package returns const values of the
		// string kind quoted, unquote the item for display.
		if typ.Kind == types.KindString {
			value, err := strconv.Unquote(item.Value)
			if err != nil {
				return nil, err
			}
			item.Value = value
		}

		if len(v.Doc) > 0 {
			text, err := godoc.ToHTML(v.Doc)
			if err != nil {
				return nil, err
			}
			item.Text = template.HTML(text)
		}

		if c.SourceURL != nil {
			item.SourceLink = new(page.SourceLink)
			item.SourceLink.Href = c.SourceURL(v.Pos.Filename, v.Pos.Line)
		}

		list.Items[i] = item
	}

	return list, nil
}

////////////////////////////////////////////////////////////////////////////////
// objkeys
////////////////////////////////////////////////////////////////////////////////

func (c *build) newObjKeysForArticle(a *Article, parent *Article) objkeys {
	var k objkeys
	k.slug = slugFromString(a.Title)
	k.path = c.RootPath + "/" + k.slug
	k.anchor = k.slug

	if parent != nil {
		if pk, ok := c.objkeys[parent]; ok {
			k.path = pk.path + "/" + k.slug
			k.anchor = pk.anchor + "." + k.slug
		}
	}

	c.makeObjKeysUnique(&k)
	return k
}

func (c *build) newObjKeysForTestGroup(tg *httptest.TestGroup, parent *Article) objkeys {
	var k objkeys
	k.path = pathFromTestGroup(tg)
	k.slug = slugFromPath(k.path)
	k.anchor = k.slug

	if parent != nil {
		if pk, ok := c.objkeys[parent]; ok {
			k.path = pathJoin(pk.path, k.path)
			k.anchor = pk.anchor + "." + k.slug
		}
	}

	c.makeObjKeysUnique(&k)
	return k
}

// make sure the slugs, paths, and anchors are unique
func (c *build) makeObjKeysUnique(k *objkeys) {
	snum := c.slugs[k.slug]
	c.slugs[k.slug] = snum + 1
	if snum > 0 {
		k.slug += "-" + strconv.Itoa(snum+1)
	}
	pnum := c.paths[k.path]
	c.paths[k.path] = pnum + 1
	if pnum > 0 {
		k.path += "-" + strconv.Itoa(pnum+1)
	}
	anum := c.anchors[k.anchor]
	c.anchors[k.anchor] = anum + 1
	if anum > 0 {
		k.anchor += "-" + strconv.Itoa(anum+1)
	}
}

////////////////////////////////////////////////////////////////////////////////
// ids & hrefs
////////////////////////////////////////////////////////////////////////////////

// getIdForCodeSnippet returns a unique id for the given CodeSnippet.
func (c *build) getIdForCodeSnippet(cs page.CodeSnippet, tg *httptest.TestGroup, lang string) string {
	if id, ok := c.csIds[cs]; ok {
		return id
	}

	id := c.objkeys[tg].anchor + ".cs-lang-" + lang

	// make sure the id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}

	// cache the id
	c.csIds[cs] = id
	return id
}

////////////////////////////////////////////////////////////////////////////////
// mediatype handling
////////////////////////////////////////////////////////////////////////////////

var errNotSupportedMediaType = fmt.Errorf("httpdoc: media type is not supported")

// isSupportedMediaType reports whether or not the given mediatype is supported.
func isSupportedMediaType(mediatype string) bool {
	return mediatype == "application/json" ||
		mediatype == "application/xml"

	// TODO add support for the following:
	// - text/csv
	// - application/x-www-form-urlencoded
	// - text/plain
}

// getLangFromMediaType
func getLangFromMediaType(mediatype string) string {
	switch mediatype {
	case "application/json":
		return "json"
	case "application/xml":
		return "xml"
	}
	return ""
	// TODO add support for the following:
	// - text/csv
	// - application/x-www-form-urlencoded
	// - text/plain
}

// marshalValue marshals the given value according to the specified mediatype.
func marshalValue(value interface{}, mediatype string, withMarkup bool) (string, error) {
	if !isSupportedMediaType(mediatype) {
		return "", errNotSupportedMediaType
	}

	switch mediatype {
	case "application/json":
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", err
		}

		if withMarkup {
			return markup.JSON(data), nil
		}
		return string(data), nil
	case "application/xml":
		data, err := xml.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", err
		}

		if withMarkup {
			return markup.XML(data), nil
		}
		return string(data), nil
	}

	panic("shouldn't reach")
	return "", nil
}

// marshalBody
func marshalBody(body httptest.Body, withMarkup bool) (text string, mediatype string, numlines int, err error) {
	mediatype, _, err = mime.ParseMediaType(body.Type())
	if err != nil || !isSupportedMediaType(mediatype) {
		return "", "", 0, errNotSupportedMediaType
	}

	r, err := body.Reader()
	if err != nil {
		return "", "", 0, err
	}

	switch mediatype {
	case "application/json":
		raw, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", 0, err
		}
		data, err := json.MarshalIndent(json.RawMessage(raw), "", "  ")
		if err != nil {
			return "", "", 0, err
		}
		numlines = 1 + bytes.Count(data, []byte{'\n'})

		if withMarkup {
			text = markup.JSON(data)
		} else {
			text = string(data)
		}
	case "application/xml":
		raw, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", 0, err
		}
		// TODO(mkopriva): encoding/xml does not provide anything analogous
		// to `json.MarshalIndent(json.RawMessage(raw), "", "  ")` so to get
		// neatly formatted text write your own XML formatter for bytes

		numlines = 1 + bytes.Count(raw, []byte{'\n'})

		if withMarkup {
			text = markup.XML(raw)
		} else {
			text = string(raw)
		}
	}

	return text, mediatype, numlines, nil
}

////////////////////////////////////////////////////////////////////////////////
// helpers
////////////////////////////////////////////////////////////////////////////////

func getCamelCaseArticleId(id string) string {
	out := make([]byte, len(id))

	j, toupper := 0, false
	for _, c := range []byte(id) {
		if c == '-' {
			toupper = true
			continue
		} else if toupper {
			c = bytes.ToUpper([]byte{c})[0]
			toupper = false
		}

		out[j], j = c, j+1
	}

	return string(out[:j])
}

// getNearestStructType returns the nearest struct types.Type in the hierarchy
// of the given types.Type. If none is found, nil will be returned instead.
func getNearestStructType(t *types.Type) *types.Type {
	for t.Kind != types.KindStruct && t.Elem != nil {
		t = t.Elem
	}

	if t.Kind == types.KindStruct {
		return t
	}
	return nil
}

// getNearestNamedType returns the nearest name types.Type in the hierarchy
// of the given types.Type. If none is found, nil will be returned instead.
func getNearestNamedType(t *types.Type) *types.Type {
	for len(t.Name) == 0 && t.Elem != nil {
		t = t.Elem
	}

	if len(t.Name) > 0 {
		return t
	}
	return nil
}

func getTestGroupDesc(tg *httptest.TestGroup) string {
	if len(tg.Desc) > 0 {
		return tg.Desc
	}
	if len(tg.E) > 0 {
		return tg.E.String()
	}
	return ""
}

func getRequestPath(req httptest.Request, pattern string) (path string) {
	path = pattern
	if req.Params != nil {
		path = req.Params.SetParams(path)
	}
	if req.Query != nil {
		path += "?" + req.Query.GetQuery()
	}
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return path
}
