package httpdoc

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
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
// - fix hrefForTopic/hrefForTestGroup
// - build Article from TestGroups
// - build Example from TestGroups
//	- generate Request examples for raw HTTP, cURL, JavaScript, Go. These need
//        to be annotated with HTML tags for syntax highlighting
//	- json produced from httptest.Request/Response.Body needs to be annotated
//        with HTML tags for syntax highlighting
// - build Example from Topics
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
	// cache of Topic ids
	tIds map[*Topic]string
	// cache of TestGroup ids
	tgIds map[*httptest.TestGroup]string
	// set of already generated hrefs
	hrefs map[string]int
	// cache of Topic hrefs
	tHrefs map[*Topic]string
	// cache of TestGroup hrefs
	tgHrefs map[*httptest.TestGroup]string

	// utilized by tests
	mode page.TestMode
}

func (c *Config) Build(toc []*TopicGroup) error {
	_, f, _, _ := runtime.Caller(1)
	dir := filepath.Dir(f)
	src, err := types.Load(dir)
	if err != nil {
		return err
	}

	// initialize config
	c.src = src
	c.ids = make(map[string]int)
	c.tIds = make(map[*Topic]string)
	c.tgIds = make(map[*httptest.TestGroup]string)
	c.hrefs = make(map[string]int)
	c.tHrefs = make(map[*Topic]string)
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
	c.buildSidebar(toc)
	if err := c.buildContent(toc); err != nil {
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

func (c *Config) buildSidebar(toc []*TopicGroup) {
	lists := make([]*page.SidebarList, len(toc))
	for i, tg := range toc {
		list := new(page.SidebarList)
		list.Title = tg.Name
		list.Items = c.newSidebarItemsFromTopics(tg.Topics, nil)

		lists[i] = list
	}
	c.page.Sidebar.Lists = lists
}

func (c *Config) newSidebarItemsFromTopics(topics []*Topic, parent *Topic) []*page.SidebarItem {
	items := make([]*page.SidebarItem, len(topics))
	for i, t := range topics {
		c.idForTopic(t, parent)

		item := new(page.SidebarItem)
		item.Text = t.Name
		item.Href = c.hrefForTopic(t, parent)

		if len(t.TestGroups) > 0 {
			items := c.newSidebarItemsFromTestGroups(t.TestGroups, t)
			item.SubItems = append(item.SubItems, items...)
		}

		if len(t.SubTopics) > 0 {
			items := c.newSidebarItemsFromTopics(t.SubTopics, t)
			item.SubItems = append(item.SubItems, items...)
		}

		items[i] = item
	}
	return items
}

func (c *Config) newSidebarItemsFromTestGroups(tgs []*httptest.TestGroup, parent *Topic) []*page.SidebarItem {
	items := make([]*page.SidebarItem, len(tgs))
	for i, tg := range tgs {
		c.idForTestGroup(tg, parent)

		item := new(page.SidebarItem)
		item.Text = tg.Desc
		item.Href = c.hrefForTestGroup(tg, parent)

		items[i] = item
	}
	return items
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

func (c *Config) buildContent(toc []*TopicGroup) error {
	for _, tg := range toc {
		if len(tg.Topics) > 0 {
			list, err := c.newArticleListFromTopics(tg.Topics, nil)
			if err != nil {
				return err
			}

			c.page.Content.Articles = append(c.page.Content.Articles, list...)
		}
	}
	return nil
}

func (c *Config) newArticleListFromTopics(topics []*Topic, parent *Topic) ([]*page.Article, error) {
	list := make([]*page.Article, len(topics))
	for i, t := range topics {
		article, err := c.newArticleFromTopic(t, parent)
		if err != nil {
			return nil, err
		}

		if len(t.TestGroups) > 0 {
			section := c.newExampleSectionEndpointOverview(t.TestGroups)
			article.Example.Sections = append(article.Example.Sections, section)

			list, err := c.newArticleListFromTestGroups(t.TestGroups, t)
			if err != nil {
				return nil, err
			}
			article.SubArticles = append(article.SubArticles, list...)
		}

		if len(t.SubTopics) > 0 {
			list, err := c.newArticleListFromTopics(t.SubTopics, t)
			if err != nil {
				return nil, err
			}
			article.SubArticles = append(article.SubArticles, list...)
		}

		list[i] = article
	}
	return list, nil
}

func (c *Config) newArticleListFromTestGroups(tgs []*httptest.TestGroup, parent *Topic) ([]*page.Article, error) {
	list := make([]*page.Article, len(tgs))
	for i, tg := range tgs {
		article, err := c.newArticleFromTestGroup(tg, parent)
		if err != nil {
			return nil, err
		}

		if err := c.buildExamplesFromTestGroup(tg, article); err != nil {
			return nil, err
		}

		list[i] = article
	}
	return list, nil
}

func (c *Config) newArticleFromTopic(t *Topic, parent *Topic) (*page.Article, error) {
	article := new(page.Article)
	article.Id = c.idForTopic(t, parent)
	article.Href = "#" + article.Id
	article.Title = t.Name

	if t.Doc != nil {
		switch v := t.Doc.(type) {
		case string:
			article.Doc = template.HTML(v)
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return nil, fmt.Errorf("httpdoc: Topic.Text:(*os.File) read error: %v", err)
			}
			article.Doc = template.HTML(b)
		default:
			return nil, fmt.Errorf("httpdoc: Topic.Text:(%T) unsupported type", v)
		}

	}

	if t.Parameters != nil {
		stype := getNearestStructType(c.src.TypeOf(t.Parameters))
		if stype == nil {
			return nil, fmt.Errorf("httpdoc: Topic.Parameters:(%T) unsupported type kind", t.Parameters)
		}

		aId := c.idForTopic(t, nil)
		aHref := c.hrefForTopic(t, nil)
		list, err := c.newFieldList(stype, aId, aHref, true, nil)
		if err != nil {
			return nil, err
		}

		section := new(page.ArticleSection)
		section.Title = "Parameters"
		section.FieldLists = []*page.FieldList{list}
		article.Sections = append(article.Sections, section)
	}

	if t.Attributes != nil {
		stype := getNearestStructType(c.src.TypeOf(t.Attributes))
		if stype == nil {
			return nil, fmt.Errorf("httpdoc: Topic.Attributes:(%T) unsupported type kind", t.Attributes)
		}

		aId := c.idForTopic(t, nil)
		aHref := c.hrefForTopic(t, nil)
		list, err := c.newFieldList(stype, aId, aHref, false, nil)
		if err != nil {
			return nil, err
		}

		section := new(page.ArticleSection)
		section.Title = "Attributes"
		section.FieldLists = []*page.FieldList{list}
		article.Sections = append(article.Sections, section)
	}

	if t.Returns != nil {
		var text template.HTML
		switch v := t.Returns.(type) {
		case string:
			text = template.HTML(v)
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return nil, fmt.Errorf("httpdoc: Topic.Returns:(*os.File) read error: %v", err)
			}
			text = template.HTML(b)
		default:
			return nil, fmt.Errorf("httpdoc: Topic.Returns:(%T) unsupported type", v)
		}

		section := new(page.ArticleSection)
		section.Title = "Returns"
		section.Text = text
		article.Sections = append(article.Sections, section)
	}

	return article, nil
}

func (c *Config) newArticleFromTestGroup(tg *httptest.TestGroup, parent *Topic) (*page.Article, error) {
	article := new(page.Article)
	article.Id = c.idForTestGroup(tg, parent)
	article.Href = "#" + article.Id
	article.Title = tg.Desc

	if tg.Doc != nil {
		switch v := tg.Doc.(type) {
		case string:
			article.Doc = template.HTML(v)
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return nil, fmt.Errorf("httpdoc: httptest.TestGroup.Doc:(*os.File) read error: %v", err)
			}
			article.Doc = template.HTML(b)
		default:
			// If none of the above then assume it's an instance of the
			// handler that's registered to handle the endpoint and use
			// its type declaration info to produce the documentation.
			if decl := c.src.TypeDeclOf(v); decl != nil {
				html, err := comment.ToHTML(decl.Doc)
				if err != nil {
					return nil, err
				}
				article.Doc = template.HTML(html)

				if len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
					file := strings.TrimPrefix(decl.Pos.Filename, c.ProjectRoot)
					href := c.RepositoryURL + file + "#" + strconv.Itoa(decl.Pos.Line)
					text := decl.PkgPath + "." + decl.Name

					article.SourceLink = new(page.SourceLink)
					article.SourceLink.Href = href
					article.SourceLink.Text = text
				}
			}
		}

	}

	if len(tg.Tests) > 0 {
		t := tg.Tests[0]

		// TODO build additional parameter documentation from:
		// - t.Request.Params
		// - t.Request.Query
		// - t.Request.Header
		if t.Request.Body != nil {
			body := t.Request.Body.Value()
			stype := getNearestStructType(c.src.TypeOf(body))
			if stype == nil {
				return nil, fmt.Errorf("httpdoc: httptest.Request.Body:(%T) unsupported type kind", body)
			}

			aId := c.idForTestGroup(tg, nil)
			aHref := c.hrefForTestGroup(tg, nil)
			list, err := c.newFieldList(stype, aId, aHref, true, nil)
			if err != nil {
				return nil, err
			}

			section := new(page.ArticleSection)
			section.Title = "Parameters"
			section.FieldLists = []*page.FieldList{list}
			article.Sections = append(article.Sections, section)
		}

		// TODO build additional attribute documentation from:
		// - t.Response.Header
		// - t.Response.Doc?
		if t.Response.Body != nil {
			body := t.Response.Body.Value()
			stype := getNearestStructType(c.src.TypeOf(body))
			if stype == nil {
				return nil, fmt.Errorf("httpdoc: httptest.Response.Body:(%T) unsupported type kind", body)
			}

			aId := c.idForTestGroup(tg, nil)
			aHref := c.hrefForTestGroup(tg, nil)
			list, err := c.newFieldList(stype, aId, aHref, false, nil)
			if err != nil {
				return nil, err
			}

			section := new(page.ArticleSection)
			section.Title = "Attributes"
			section.FieldLists = []*page.FieldList{list}
			article.Sections = append(article.Sections, section)
		}

		// TODO Returns ...
	}

	return article, nil
}

func (c *Config) newFieldList(typ *types.Type, aId string, aHref string, withValidation bool, path []string) (*page.FieldList, error) {
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
		item.Id = aId + "_" + item.Path + item.Name

		// the field's anchor
		item.Href = aHref
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
			subList, err := c.newFieldList(stype, aId, aHref, withValidation, append(path, item.Name))
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

		if withValidation {
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

func (c *Config) newExampleSectionEndpointOverview(tgs []*httptest.TestGroup) *page.ExampleSection {
	overview := new(page.EndpointOverview)
	overview.Title = "ENDPOINTS"

	for _, tg := range tgs {
		item := new(page.EndpointItem)
		item.Href = c.hrefForTestGroup(tg, nil)
		item.Method = tg.Endpoint.Method
		item.Pattern = tg.Endpoint.Pattern
		item.Tooltip = tg.Desc
		overview.Items = append(overview.Items, item)
	}

	return &page.ExampleSection{EndpointOverview: overview}
}

func (c *Config) buildExamplesFromTestGroup(tg *httptest.TestGroup, a *page.Article) error {
	for _, t := range tg.Tests {
		if t.Request.Body != nil {
			typ := c.src.TypeOf(t.Request.Body.Value())
			_ = typ
		}
		if t.Response.Body != nil {
			typ := c.src.TypeOf(t.Response.Body.Value())
			_ = typ
		}
	}
	return nil
}

// idForTopic returns a unique id value for the given Topic.
func (c *Config) idForTopic(t *Topic, parent *Topic) string {
	if id, ok := c.tIds[t]; ok {
		return id
	}

	id := strings.Map(func(r rune) rune {
		// TODO(mkopriva): handle non ascii characters, e.g. japanese, chinese, arabic, etc.
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(t.Name))

	if parent != nil {
		id = c.idForTopic(parent, nil) + "_" + id
	}

	// make sure the id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}

	// cache the id
	c.tIds[t] = id
	return id
}

// idForTestGroup returns a unique id for the given TestGroup.
func (c *Config) idForTestGroup(tg *httptest.TestGroup, parent *Topic) string {
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
			// their own "hrefForTestGroup" implementation based on their own need.
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
		id = c.idForTopic(parent, nil) + "_" + id
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

// hrefForTopic returns an href string for the given Topic.
func (c *Config) hrefForTopic(t *Topic, parent *Topic) string {
	if href, ok := c.tHrefs[t]; ok {
		return href
	}

	id := c.idForTopic(t, parent)
	href := "/" + id
	if parent != nil {
		href = c.hrefForTopic(parent, nil)
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
	c.tHrefs[t] = href
	return href
}

// hrefForTestGroup returns an href string for the given TestGroup.
func (c *Config) hrefForTestGroup(tg *httptest.TestGroup, parent *Topic) string {
	if href, ok := c.tgHrefs[tg]; ok {
		return href
	}

	id := c.idForTestGroup(tg, parent)
	href := "/" + id
	if parent != nil {
		href = c.hrefForTopic(parent, nil)
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
