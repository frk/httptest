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
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/comment"
	"github.com/frk/httptest/internal/markup"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/types"
	"github.com/frk/tagutil"
)

// holds the state of the documentaiton build process
type build struct {
	Config
	// the input
	dir ArticleDirectory
	// source code info
	src *types.Source
	// the page being built
	page page.Page
	// the sidebar list currently being built, or nil
	sbls *page.SidebarList
	// buffer to write to
	buf bytes.Buffer

	// set of already generated ids
	ids map[string]int
	// cache of Article ids
	aIds map[*Article]string
	// cache of TestGroup ids
	tgIds map[*httptest.TestGroup]string
	// set of already generated hrefs
	hrefs map[string]int
	// cache of Article hrefs
	aHrefs map[*Article]string
	// cache of TestGroup hrefs
	tgHrefs map[*httptest.TestGroup]string

	// utilized by tests
	mode page.TestMode
}

// The argument skip is for the runtime.Caller func, it is the number of stack
// frames to ascend, with 0 identifying the caller of loadCallerSource.
func (c *build) loadCallerSource(skip int) error {
	_, f, _, _ := runtime.Caller(1 + skip)
	d := filepath.Dir(f)
	src, err := types.Load(d)
	if err != nil {
		return err
	}

	c.src = src
	return nil
}

func (c *build) run() error {
	// initialize build
	c.ids = make(map[string]int)
	c.aIds = make(map[*Article]string)
	c.tgIds = make(map[*httptest.TestGroup]string)
	c.hrefs = make(map[string]int)
	c.aHrefs = make(map[*Article]string)
	c.tgHrefs = make(map[*httptest.TestGroup]string)

	// ensure the configured hrefs and the hrefs generated later don't collide
	c.hrefs[c.RootPath] = 0
	c.hrefs[c.logoURL] = 0
	c.hrefs[c.SigninPath] = 0

	// build & write
	if err := c.buildSidebar(); err != nil {
		return err
	}
	if err := c.buildContent(); err != nil {
		return err
	}
	if err := page.Write(&c.buf, c.page, c.mode); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

func (c *build) buildSidebar() error {
	c.page.Title = c.PageTitle
	c.page.Sidebar.Header.Title = c.PageTitle
	c.page.Sidebar.Header.RootURL = template.URL(c.RootPath)
	c.page.Sidebar.Header.LogoURL = template.URL(c.logoURL)
	c.page.Sidebar.Footer.SigninURL = template.URL(c.SigninPath)

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

var errNoArticleTitle = fmt.Errorf("httpdoc: Article.Title is required.")

func (c *build) newSidebarItemsFromArticles(articles []*Article, parent *Article) ([]*page.SidebarItem, error) {
	items := make([]*page.SidebarItem, len(articles))
	for i, a := range articles {
		// an article title is required, fail if none was provided
		if a.Title == "" {
			return nil, errNoArticleTitle
		}

		item := new(page.SidebarItem)
		item.Text = a.Title
		item.Href = c.getHrefForArticle(a, parent)

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
			item := new(page.SidebarItem)
			item.Text = desc
			item.Href = c.getHrefForTestGroup(g, parent)
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

		list, err := c.newArticleElementListFromArticles(g.Articles, nil)
		if err != nil {
			return err
		}

		c.page.Content.Articles = append(c.page.Content.Articles, list...)
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
	aElem.Id = c.getIdForArticle(a, parent)
	aElem.Href = c.getHrefForArticle(a, parent)
	aElem.Title = a.Title

	if a.Text != nil {
		html, err := c.newHTML(a.Text, nil)
		if err != nil {
			return nil, err
		}
		aElem.Text = html
	}

	if a.Code != nil {
		switch v := a.Code.(type) {
		case string, *os.File, HTMLer:
			html, err := c.newHTML(a.Code, nil)
			if err != nil {
				return nil, err
			}

			s := &page.ExampleText{Text: html}
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
	aElem.Id = c.getIdForTestGroup(tg, parent)
	aElem.Href = c.getHrefForTestGroup(tg, parent)
	aElem.Title = getTestGroupDesc(tg)

	if tg.DocA != nil {
		var decl types.TypeDecl
		html, err := c.newHTML(tg.DocA, &decl)
		if err != nil {
			return nil, err
		}
		aElem.Text = html

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
			html, err := c.newHTML(v, nil)
			if err != nil {
				return nil, err
			}
			section := &page.ArticleAuthInfo{Text: html}
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
		html, err := c.newHTML(tg.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ArticleText{Text: html}
		aElem.Sections = append(aElem.Sections, section)
	}
	return aElem, nil
}

////////////////////////////////////////////////////////////////////////////////
// Example Sections
////////////////////////////////////////////////////////////////////////////////

func (c *build) newExampleSectionsFromTestGroup(t *httptest.Test, tg *httptest.TestGroup) (sections []page.ExampleSection, err error) {
	if t.DocA != nil {
		html, err := c.newHTML(t.DocA, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleText{Text: html}
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
		html, err := c.newHTML(t.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleText{Text: html}
		sections = append(sections, section)
	}
	return sections, nil
}

func (c *build) newExampleEndpoints(tgs []*httptest.TestGroup) *page.ExampleEndpoints {
	section := new(page.ExampleEndpoints)
	section.Title = "ENDPOINTS"

	for _, tg := range tgs {
		item := new(page.EndpointItem)
		item.Href = c.getHrefForTestGroup(tg, nil)
		item.Method = tg.Endpoint.Method
		item.Pattern = tg.Endpoint.Pattern
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

	section := new(page.ExampleObject)
	section.Title = title
	section.Lang = getLangFromMediaType(mediatype)
	section.Text = template.HTML(text)
	return section, nil
}

func (c *build) newExampleRequest(req httptest.Request, tg *httptest.TestGroup) (*page.ExampleRequest, error) {
	section := new(page.ExampleRequest)
	section.Title = "REQUEST"
	section.Method = tg.Endpoint.Method
	section.Pattern = tg.Endpoint.Pattern

	csr, err := c.newCodeSnippetRequest(tg.Endpoint, req)
	if err != nil {
		return nil, err
	}
	for _, st := range c.SnippetTypes {
		switch st {
		case SNIPP_HTTP:
			snip := c.newCodeSnippetHTTP(*csr)
			section.Snippets = append(section.Snippets, snip)
		case SNIPP_CURL:
			snip := c.newCodeSnippetCURL(*csr)
			section.Snippets = append(section.Snippets, snip)
		}
		// TODO snippets for vanilla-js, Go
	}

	return section, nil
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
		text, mediatype, err := marshalBody(resp.Body, true)
		if err != nil {
			return nil, err
		}

		section.Body = template.HTML(text)
		section.Lang = getLangFromMediaType(mediatype)

	}

	return section, nil
}

////////////////////////////////////////////////////////////////////////////////
// Code Snippets
////////////////////////////////////////////////////////////////////////////////

func (c *build) newCodeSnippetRequest(ep httptest.Endpoint, req httptest.Request) (*page.CodeSnippetRequest, error) {
	csr := new(page.CodeSnippetRequest)
	csr.Method = ep.Method
	csr.Host = trimURLScheme(c.ExampleHost)

	// prepare the URL's path
	csr.Path = ep.Pattern
	if req.Params != nil {
		csr.Path = req.Params.SetParams(csr.Path)
	}
	if req.Query != nil {
		csr.Path += "?" + req.Query.GetQuery()
	}
	if len(csr.Path) > 0 && csr.Path[0] != '/' {
		csr.Path = "/" + csr.Path
	}

	// the target URL
	csr.URL = c.ExampleHost + csr.Path

	// headers
	if req.Header != nil {
		for key, values := range req.Header.GetHeader() {
			// if body's present the content type and length
			// will be set automatically so skip them here.
			if req.Body != nil && (key == "Content-Type" || key == "Content-Length") {
				continue
			}

			for _, val := range values {
				item := page.HeaderItem{}
				item.Key = key
				item.Value = val
				csr.Header = append(csr.Header, item)
			}
		}
	}

	// prep the body
	if req.Body != nil {
		text, _, err := marshalBody(req.Body, false)
		if err != nil {
			return nil, err
		}

		h1 := page.HeaderItem{Key: "Content-Type", Value: req.Body.Type()}
		h2 := page.HeaderItem{Key: "Content-Length", Value: strconv.Itoa(len(text))}
		csr.Header = append(csr.Header, h1, h2)
		csr.Body = template.HTML(text)
	}

	// finally sort the headers, if any are present
	if len(csr.Header) > 0 {
		sort.Sort(&headerSorter{csr.Header})
	}
	return csr, nil
}

func (c *build) newCodeSnippetHTTP(csr page.CodeSnippetRequest) *page.CodeSnippetHTTP {
	snip := new(page.CodeSnippetHTTP)
	snip.CodeSnippetRequest = csr
	return snip
}

func (c *build) newCodeSnippetCURL(csr page.CodeSnippetRequest) *page.CodeSnippetCURL {
	snip := new(page.CodeSnippetCURL)
	snip.CodeSnippetRequest = csr

	// if the method is GET we can omit it since GET is the default cURL
	// method and is seldom if ever used explicitly with the -X option
	if strings.ToUpper(snip.Method) == "GET" {
		snip.Method = ""
	}

	return snip
}

////////////////////////////////////////////////////////////////////////////////
// Misc.
////////////////////////////////////////////////////////////////////////////////

var errNotNamedType = fmt.Errorf("httpdoc: type is not named")

func (c *build) newHTML(value interface{}, decl *types.TypeDecl) (html template.HTML, err error) {
	switch v := value.(type) {
	case string:
		html = template.HTML(v)
	case *os.File:
		text, err := ioutil.ReadAll(v)
		if err != nil {
			return "", err
		}
		html = template.HTML(text)
	case HTMLer:
		text, err := v.HTML()
		if err != nil {
			return "", err
		}
		html = template.HTML(text)
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
		text, err := comment.ToHTML(d.Doc)
		if err != nil {
			return "", err
		}
		html = template.HTML(text)

		if decl != nil { // retain the type declaration?
			*decl = *d
		}
	}

	return html, nil
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
		item.Href = aElem.Href
		if i := strings.IndexByte(item.Href, '#'); i > -1 {
			item.Href = item.Href[:i]
		}
		item.Href = item.Href + "#" + item.Id

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
			text, err := comment.ToHTML(f.Doc)
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
			subList, err := c._newFieldList(stype, aElem, class, isInput, idpfx, append(path, item.Name))
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
			text, err := comment.ToHTML(v.Doc)
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
// helpers
////////////////////////////////////////////////////////////////////////////////

// getIdForArticle returns a unique id value for the given Article.
func (c *build) getIdForArticle(a *Article, parent *Article) string {
	if id, ok := c.aIds[a]; ok {
		return id
	}

	id := strings.Map(func(r rune) rune {
		// TODO(mkopriva): handle non ascii characters, e.g. japanese, chinese, arabic, etc.
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(a.Title))

	if parent != nil {
		id = c.getIdForArticle(parent, nil) + "." + id
	}

	// make sure the id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}

	// cache the id
	c.aIds[a] = id
	return id
}

// getIdForTestGroup returns a unique id for the given TestGroup.
func (c *build) getIdForTestGroup(tg *httptest.TestGroup, parent *Article) string {
	if id, ok := c.tgIds[tg]; ok {
		return id
	}

	var id string
	if len(tg.Desc) > 0 {
		id = strings.Map(func(r rune) rune {
			// TODO(mkopriva): handle non ascii characters, e.g. japanese, chinese, arabic, etc.
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return '-'
		}, strings.ToLower(tg.Desc))
	} else {
		// default to tg.Endpoint if no tg.Desc was set
		method := strings.ToLower(strings.Trim(tg.Endpoint.Method, "/-"))
		pattern := strings.ToLower(strings.Trim(tg.Endpoint.Pattern, "/-"))
		id = method + "-" + pattern

		// replace "/" with "-", and remove placeholder delimiters
		id = strings.Map(func(r rune) rune {
			// TODO(mkopriva): allow for handling of more complex endpoint
			// patterns, i.e. some routers allow specifying regular expressions,
			// or they use delimiters different from "{" and "}".
			//
			// Consider providing a Config.<Field> so that the user can supply
			// their own "getHrefForTestGroup" implementation based on their own need.
			if r == '{' || r == '}' {
				return -1
			}
			if r == '/' {
				return '-'
			}
			return r
		}, id)
	}

	if parent != nil {
		id = c.getIdForArticle(parent, nil) + "." + id
	}

	// make sure the id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}

	// cache the id
	c.tgIds[tg] = id
	return id
}

// getHrefForArticle returns an href string for the given Article.
func (c *build) getHrefForArticle(a *Article, parent *Article) string {
	if href, ok := c.aHrefs[a]; ok {
		return href
	}

	id := c.getIdForArticle(a, parent)
	href := c.RootPath + "/" + id
	if parent != nil {
		href = c.getHrefForArticle(parent, nil)
		if i := strings.IndexByte(href, '#'); i > -1 {
			href = href[:i]
		}
		href += "#" + id
	}

	// make sure the href is unique
	count := c.hrefs[href]
	c.hrefs[href] = count + 1
	if count > 0 {
		href += "-" + strconv.Itoa(count+1)
	}

	// cache the href
	c.aHrefs[a] = href
	return href
}

// getHrefForTestGroup returns an href string for the given TestGroup.
func (c *build) getHrefForTestGroup(tg *httptest.TestGroup, parent *Article) string {
	if href, ok := c.tgHrefs[tg]; ok {
		return href
	}

	id := c.getIdForTestGroup(tg, parent)
	href := c.RootPath + "/" + id
	if parent != nil {
		href = c.getHrefForArticle(parent, nil)
		if i := strings.IndexByte(href, '#'); i > -1 {
			href = href[:i]
		}
		href += "#" + id
	}

	// make sure the returned href is unique
	count := c.hrefs[href]
	c.hrefs[href] = count + 1
	if count > 0 {
		href += "-" + strconv.Itoa(count+1)
	}

	// cache the href
	c.tgHrefs[tg] = href
	return href
}

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
func marshalBody(body httptest.Body, withMarkup bool) (text string, mediatype string, err error) {
	mediatype, _, err = mime.ParseMediaType(body.Type())
	if err != nil || !isSupportedMediaType(mediatype) {
		return "", "", errNotSupportedMediaType
	}

	r, err := body.Reader()
	if err != nil {
		return "", "", err
	}

	switch mediatype {
	case "application/json":
		raw, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", err
		}
		data, err := json.MarshalIndent(json.RawMessage(raw), "", "  ")
		if err != nil {
			return "", "", err
		}

		if withMarkup {
			text = markup.JSON(data)
		} else {
			text = string(data)
		}
	case "application/xml":
		raw, err := ioutil.ReadAll(r)
		if err != nil {
			return "", "", err
		}

		// TODO(mkopriva): encoding/xml does not provide anything analogous
		// to `json.MarshalIndent(json.RawMessage(raw), "", "  ")` so to get
		// neatly formatted text write your own XML formatter for bytes

		if withMarkup {
			text = markup.XML(raw)
		} else {
			text = string(raw)
		}
	}

	return text, mediatype, nil
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

// getNearestStructType returns the nearest struct types.Type in the hierarchy
// of the given types.Type. If none is found, nil will be returned instead.
func getNearestStructType(t *types.Type) *types.Type {
	for t.Kind != types.KindString && t.Elem != nil {
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
	if ep := tg.Endpoint; len(ep.Pattern) > 0 {
		if len(ep.Method) > 0 {
			return ep.Method + " " + ep.Pattern
		}
		return ep.Pattern
	}
	return ""
}
