package httpdoc

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/comment"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/types"
	"github.com/frk/tagutil"
)

// TODO
// - build ArticleElement from TestGroups
// - build Example from TestGroups
//	- generate Request examples for raw HTTP, cURL, JavaScript, Go. These need
//        to be annotated with HTML tags for syntax highlighting
//	- json produced from httptest.Request/Response.Body needs to be annotated
//        with HTML tags for syntax highlighting
// - build Example from Articles
// -

type Config struct {
	// ProjectRoot and RepositoryURL are used to generate source links
	// for handlers, struct types, fields, and enums. If one or both of
	// them are left unset then the links will not be generated.
	//
	// The ProjectRoot field should be set to the local (i.e. on the host machine),
	// root directory of the project for which the documentation is being generated.
	//
	// The RepositoryURL field should be set to the remote, web-accessible, root
	// location of the project for which the documentation is being generated.
	// For example:
	//	// for a github repo the url should have the following format.
	//	RepositoryURL: "https://github.com/<user>/<project>/tree/<branch>/",
	//	// for a bitbucket repo the url should have the following format.
	//	RepositoryURL: "https://bitbucket.org/<user>/<project>/src/<branch>/",
	//
	ProjectRoot, RepositoryURL string
	// The tag to be used to resolve a field's name for the documentation, defaults to "json".
	// If no name is present in the tag's value the field's name will be used as fallback.
	FieldNameTag string
	// FieldType returns the name for a specific field's type based on
	// given reflect.StructField value.
	//
	// If FieldType is nil or it returns false as the second return
	// value (ok) it will fall back to the default behaviour.
	FieldType func(field reflect.StructField) (typeName string, ok bool)
	// FieldSetting returns values that are used to document whether the given
	// field is required, optional, or something else. The structType argument
	// represents the type of struct to which the field belongs.
	//
	// The returned label is used in the corresponding element's class.
	// The returned text is used as the corresponding element's content.
	// The returned ok value indicates whether or not the field's setting
	// documentation should not be generated.
	//
	// If FieldSetting is nil then the documentation will be generated based
	// on the field's "doc" tag and if the field doesn't have a "doc" tag then
	// the field will be documented as optional.
	FieldSetting func(field reflect.StructField, structType reflect.Type) (label, text string, ok bool)
	// FieldValidation returns the documentation on the given field's validity
	// requirements. The structType argument represents the type of struct
	// to which the field belongs. If the returned text is empty then no
	// documentation will be rendered.
	//
	// If FieldValidation is nil then no documentation on the field validity
	// requirements will be generated.
	FieldValidation func(field reflect.StructField, structType reflect.Type) (text template.HTML)

	//
	buf bytes.Buffer
	// source code info
	src *types.Source
	// the page being built
	page page.Page
	// the sidebar list currently being built, or nil
	sbls *page.SidebarList

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

func (c *Config) Build(dir ArticleDirectory) error {
	_, f, _, _ := runtime.Caller(1)
	fdir := filepath.Dir(f)
	src, err := types.Load(fdir)
	if err != nil {
		return err
	}

	// initialize config
	c.src = src
	c.ids = make(map[string]int)
	c.aIds = make(map[*Article]string)
	c.tgIds = make(map[*httptest.TestGroup]string)
	c.hrefs = make(map[string]int)
	c.aHrefs = make(map[*Article]string)
	c.tgHrefs = make(map[*httptest.TestGroup]string)

	// defaults
	if len(c.FieldNameTag) == 0 {
		c.FieldNameTag = "json"
	}
	if c.FieldSetting == nil {
		c.FieldSetting = defaultFieldSetting
	}

	// NOTE: The code that constructs source links requires both of these
	// to *not* end in slash, which is why this is here.
	if l := len(c.RepositoryURL); l > 0 && c.RepositoryURL[l-1] == '/' {
		c.RepositoryURL = c.RepositoryURL[:l-1]
	}
	if l := len(c.ProjectRoot); l > 0 && c.ProjectRoot[l-1] == '/' {
		c.ProjectRoot = c.ProjectRoot[:l-1]
	}

	// build & write
	c.buildSidebar(dir)
	if err := c.buildContent(dir); err != nil {
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

func (c *Config) buildSidebar(dir ArticleDirectory) {
	lists := make([]*page.SidebarList, len(dir))
	for i, g := range dir {
		list := new(page.SidebarList)
		list.Title = g.Name
		list.Items = c.newSidebarItemsFromArticles(g.Articles, nil)

		lists[i] = list
	}
	c.page.Sidebar.Lists = lists
}

func (c *Config) newSidebarItemsFromArticles(articles []*Article, parent *Article) []*page.SidebarItem {
	items := make([]*page.SidebarItem, len(articles))
	for i, a := range articles {
		item := new(page.SidebarItem)
		item.Text = a.Title
		item.Href = c.getHrefForArticle(a, parent)

		if len(a.TestGroups) > 0 {
			items := c.newSidebarItemsFromTestGroups(a.TestGroups, a)
			item.SubItems = append(item.SubItems, items...)
		}
		if len(a.SubArticles) > 0 {
			items := c.newSidebarItemsFromArticles(a.SubArticles, a)
			item.SubItems = append(item.SubItems, items...)
		}

		items[i] = item
	}
	return items
}

func (c *Config) newSidebarItemsFromTestGroups(tgs []*httptest.TestGroup, parent *Article) []*page.SidebarItem {
	items := make([]*page.SidebarItem, len(tgs))
	for i, g := range tgs {
		item := new(page.SidebarItem)
		item.Text = g.Desc
		item.Href = c.getHrefForTestGroup(g, parent)

		items[i] = item
	}
	return items
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

func (c *Config) buildContent(dir ArticleDirectory) error {
	for _, g := range dir {
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

func (c *Config) newArticleElementListFromArticles(articles []*Article, parent *Article) ([]*page.ArticleElement, error) {
	list := make([]*page.ArticleElement, len(articles))
	for i, a := range articles {
		aElem, err := c.newArticleElementFromArticle(a, parent)
		if err != nil {
			return nil, err
		}

		if len(a.TestGroups) > 0 {
			ov := c.newEndpointOverview(a.TestGroups)
			section := &page.ExampleSection{EndpointOverview: ov}
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

func (c *Config) newArticleElementListFromTestGroups(tgs []*httptest.TestGroup, parent *Article) ([]*page.ArticleElement, error) {
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

func (c *Config) newArticleElementFromArticle(a *Article, parent *Article) (*page.ArticleElement, error) {
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
			section := &page.ExampleSection{Text: html}
			aElem.Example.Sections = append(aElem.Example.Sections, section)
		case Valuer, interface{}:
			obj, err := c.newExampleObject(v, a.Type)
			if err != nil {
				return nil, err
			}
			section := &page.ExampleSection{ExampleObject: obj}
			aElem.Example.Sections = append(aElem.Example.Sections, section)

			// Optionally generate a section of field docs to add
			// to the article's text-column. If not struct, ignore.
			list, err := c.newFieldList(v, aElem, false)
			if err != nil && err != errNotStructType {
				return nil, err
			}
			if err != errNotStructType {
				section := new(page.ArticleSection)
				section.Title = "Fields"
				section.FieldLists = []*page.FieldList{list}
				aElem.Sections = append(aElem.Sections, section)
			}
		}
	}

	return aElem, nil
}

func (c *Config) newArticleElementFromTestGroup(tg *httptest.TestGroup, parent *Article) (*page.ArticleElement, error) {
	aElem := new(page.ArticleElement)
	aElem.Id = c.getIdForTestGroup(tg, parent)
	aElem.Href = c.getHrefForTestGroup(tg, parent)
	aElem.Title = tg.Desc

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
		if len(decl.Name) > 0 && len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
			file := strings.TrimPrefix(decl.Pos.Filename, c.ProjectRoot)
			href := c.RepositoryURL + file + "#" + strconv.Itoa(decl.Pos.Line)
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
			section := &page.ArticleSection{AuthInfo: html}
			aElem.Sections = append(aElem.Sections, section)
		}

		// input field lists
		var inputFields []*page.FieldList
		if v, ok := t.Request.Header.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, true)
			if err != nil {
				return nil, err
			}
			list.Title = "Header"
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Params.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, true)
			if err != nil {
				return nil, err
			}
			list.Title = "Path Params"
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Query.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, true)
			if err != nil {
				return nil, err
			}
			list.Title = "Query"
			inputFields = append(inputFields, list)
		}
		if v, ok := t.Request.Body.(Valuer); ok && v != nil {
			list, err := c.newFieldList(v, aElem, true)
			if err != nil {
				return nil, err
			}
			list.Title = "Body"
			inputFields = append(inputFields, list)
		}
		if len(inputFields) > 0 {
			section := new(page.ArticleSection)
			section.Title = "INPUT"
			section.FieldLists = inputFields
			aElem.Sections = append(aElem.Sections, section)
		}

		// NOTE(mkopriva): for now create an output section only for
		// endpoints where the test request doesn't have a body.
		// Consider, however, to change this in the future, maybe by
		// using a tabbed view of input and output for every endpoint.
		if v, ok := t.Request.Body.(Valuer); !ok || v == nil {
			var outputFields []*page.FieldList
			if v, ok := t.Response.Header.(Valuer); ok && v != nil {
				list, err := c.newFieldList(v, aElem, true)
				if err != nil {
					return nil, err
				}
				list.Title = "Header"
				outputFields = append(outputFields, list)
			}
			if v, ok := t.Response.Body.(Valuer); ok && v != nil {
				list, err := c.newFieldList(v, aElem, false)
				if err != nil {
					return nil, err
				}
				list.Title = "Body"
				outputFields = append(outputFields, list)
			}
			if len(outputFields) > 0 {
				section := new(page.ArticleSection)
				section.Title = "OUTPUT"
				section.FieldLists = outputFields
				aElem.Sections = append(aElem.Sections, section)
			}
		}
	}

	if tg.DocB != nil {
		html, err := c.newHTML(tg.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ArticleSection{Text: html}
		aElem.Sections = append(aElem.Sections, section)
	}
	return aElem, nil
}

func (c *Config) newExampleSectionsFromTestGroup(t *httptest.Test, tg *httptest.TestGroup) (sections []*page.ExampleSection, err error) {
	if t.DocA != nil {
		html, err := c.newHTML(t.DocA, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleSection{Text: html}
		sections = append(sections, section)
	}

	req := &page.ExampleRequest{}
	req.Method = tg.Endpoint.Method
	req.Pattern = tg.Endpoint.Pattern
	// TODO create snippets from request

	section := &page.ExampleSection{ExampleRequest: req}
	sections = append(sections, section)

	// Currently this adds an ExampleResponse section only if a body is present
	// and it implements the Valuer interface, however it may be useful to generate
	// a response example in other cases as well, e.g. to at least document the
	// status code, or the http headers of the response...
	if v, ok := t.Response.Body.(Valuer); ok && v != nil {
		// TODO value, err := v.Value()
		// TODO if err != nil {
		// TODO 	return nil, err
		// TODO }

		// TODO // resolve the body type first, if can't handle the type then skip the section
		// TODO typ := t.Response.Body.Type()
		// TODO if typ, _, err = mime.ParseMediaType(typ); err != nil {
		// TODO 	return nil, err
		// TODO }

		// TODO if isSupportedMediaType(typ) {
		// TODO 	resp := &page.ExampleResponse{Status: t.Response.StatusCode}
		// TODO 	if t.Response.Header != nil {
		// TODO 		resp.Header = nil // TODO htmlify.Header(t.Response.Header.GetHeader())
		// TODO 	}

		// TODO 	text, err := addMarkupToValue(value, typ)
		// TODO 	if err != nil {
		// TODO 		return nil, err
		// TODO 	}
		// TODO 	resp.Code = text

		// TODO 	section := &page.ExampleSection{ExampleResponse: resp}
		// TODO 	section.Title = "RESPONSE"
		// TODO 	sections = append(sections, section)
		// TODO }
	}

	if t.DocB != nil {
		html, err := c.newHTML(t.DocB, nil)
		if err != nil {
			return nil, err
		}
		section := &page.ExampleSection{Text: html}
		sections = append(sections, section)
	}
	return sections, nil
}

var errNotNamedType = fmt.Errorf("httpdoc: type is not named")

func (c *Config) newHTML(value interface{}, decl *types.TypeDecl) (html template.HTML, err error) {
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
		if d != nil {
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

func (c *Config) newFieldList(value interface{}, aElem *page.ArticleElement, isInput bool) (list *page.FieldList, err error) {
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

	return c._newFieldList(typ, aElem, isInput, nil)
}

func (c *Config) _newFieldList(typ *types.Type, aElem *page.ArticleElement, isInput bool, path []string) (*page.FieldList, error) {
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
		item.Id = aElem.Id + "_" + item.Path + item.Name

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
			html, err := comment.ToHTML(f.Doc)
			if err != nil {
				return nil, err
			}
			item.Doc = template.HTML(html)
		}

		// the field type's enum values
		if len(ftype.Values) > 0 {
			enumList, err := c.newValueListEnums(ftype)
			if err != nil {
				return nil, err
			}
			item.ValueList = enumList
		}

		// the field's sub fields
		if stype := getNearestStructType(ftype); stype != nil && len(stype.Fields) > 0 {
			subList, err := c._newFieldList(stype, aElem, isInput, append(path, item.Name))
			if err != nil {
				return nil, err
			}
			item.SubFields = subList.Items
		}

		// the field's source link
		if len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
			file := strings.TrimPrefix(f.Pos.Filename, c.ProjectRoot)
			href := c.RepositoryURL + file + "#" + strconv.Itoa(f.Pos.Line)

			item.SourceLink = new(page.SourceLink)
			item.SourceLink.Href = href
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

func (c *Config) newValueListEnums(typ *types.Type) (*page.ValueList, error) {
	list := new(page.ValueList)
	list.Class = "enum"
	list.Title = "Possible enum values"
	list.Items = make([]*page.ValueItem, len(typ.Values))

	for i, v := range typ.Values {
		value := new(page.ValueItem)
		value.Text = v.Value

		// the internal/types package returns const values of the
		// string kind quoted, unquote the value for display.
		if typ.Kind == types.KindString {
			text, err := strconv.Unquote(value.Text)
			if err != nil {
				return nil, err
			}
			value.Text = text
		}

		if len(v.Doc) > 0 {
			html, err := comment.ToHTML(v.Doc)
			if err != nil {
				return nil, err
			}
			value.Doc = template.HTML(html)
		}

		if len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
			file := strings.TrimPrefix(v.Pos.Filename, c.ProjectRoot)
			href := c.RepositoryURL + file + "#" + strconv.Itoa(v.Pos.Line)

			value.SourceLink = new(page.SourceLink)
			value.SourceLink.Href = href
		}

		list.Items[i] = value
	}

	return list, nil
}

func (c *Config) newExampleObject(value interface{}, typ string) (obj *page.ExampleObject, err error) {
	// resolve the mime type
	if typ == "" {
		if body, ok := value.(httptest.Body); ok && body != nil {
			if typ, _, err = mime.ParseMediaType(body.Type()); err != nil {
				return nil, err
			}
		}
	}

	// TODO if !isSupportedMediaType(typ) {
	// TODO 	typ = "application/json" // default
	// TODO }

	// TODO // if this is a Valuer then get the underlying value
	// TODO if v, ok := value.(Valuer); ok {
	// TODO 	if value, err = v.Value(); err != nil {
	// TODO 		return nil, err
	// TODO 	}
	// TODO }

	var text template.HTML
	// TODO // marshal the value and decorate it with html for syntax highlighting
	// TODO text, err := addMarkupToValue(value, typ)
	// TODO if err != nil {
	// TODO 	return nil, err
	// TODO }

	return &page.ExampleObject{Type: typ, Code: text}, nil
}

func (c *Config) newEndpointOverview(tgs []*httptest.TestGroup) *page.EndpointOverview {
	ov := new(page.EndpointOverview)
	ov.Title = "ENDPOINTS"

	for _, tg := range tgs {
		item := new(page.EndpointItem)
		item.Href = c.getHrefForTestGroup(tg, nil)
		item.Method = tg.Endpoint.Method
		item.Pattern = tg.Endpoint.Pattern
		item.Tooltip = tg.Desc
		ov.Items = append(ov.Items, item)
	}

	return ov
}

// getIdForArticle returns a unique id value for the given Article.
func (c *Config) getIdForArticle(a *Article, parent *Article) string {
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
		id = c.getIdForArticle(parent, nil) + "_" + id
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
func (c *Config) getIdForTestGroup(tg *httptest.TestGroup, parent *Article) string {
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
		pattern := strings.Map(func(r rune) rune {
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
		}, tg.Endpoint.Pattern)

		id = strings.ToLower(pattern + "-" + tg.Endpoint.Method)
	}

	if parent != nil {
		id = c.getIdForArticle(parent, nil) + "_" + id
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
func (c *Config) getHrefForArticle(a *Article, parent *Article) string {
	if href, ok := c.aHrefs[a]; ok {
		return href
	}

	id := c.getIdForArticle(a, parent)
	href := "/" + id
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
func (c *Config) getHrefForTestGroup(tg *httptest.TestGroup, parent *Article) string {
	if href, ok := c.tgHrefs[tg]; ok {
		return href
	}

	id := c.getIdForTestGroup(tg, parent)
	href := "/" + id
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

func defaultFieldSetting(s reflect.StructField, t reflect.Type) (label, text string, ok bool) {
	const (
		required = "required"
		optional = "optional"
	)

	tag := tagutil.New(string(s.Tag))
	if tag.Contains("doc", required) {
		return required, required, true
	}
	return optional, optional, true
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
