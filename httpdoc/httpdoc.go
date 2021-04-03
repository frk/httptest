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
	// the nav group currently being built, or nil
	ng *page.SidebarNavGroup

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
	// to not end in slash, which is why this is here.
	if l := len(c.RepositoryURL); l > 0 && c.RepositoryURL[l-1] == '/' {
		c.RepositoryURL = c.RepositoryURL[:l-1]
	}
	if l := len(c.ProjectRoot); l > 0 && c.ProjectRoot[l-1] == '/' {
		c.ProjectRoot = c.ProjectRoot[:l-1]
	}

	// build & write
	if err := c.buildSidebar(toc); err != nil {
		return err
	}
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

func (c *Config) buildSidebar(toc []*TopicGroup) error {
	for _, tg := range toc {
		c.ng = new(page.SidebarNavGroup)
		c.ng.Title = tg.Name

		if err := c.buildSidebarNavFromTopics(tg.Topics, nil); err != nil {
			return err
		}

		c.page.Sidebar.NavGroups = append(c.page.Sidebar.NavGroups, c.ng)
	}
	return nil
}

func (c *Config) buildSidebarNavFromTopics(topics []*Topic, parent *Topic) error {
	for _, t := range topics {
		c.idForTopic(t, parent)

		t.navItem = new(page.SidebarNavItem)
		t.navItem.Text = t.Name
		t.navItem.Href = c.hrefForTopic(t, parent)

		if err := c.buildSidebarNavFromTestGroups(t.TestGroups, t); err != nil {
			return err
		}
		if err := c.buildSidebarNavFromTopics(t.SubTopics, t); err != nil {
			return err
		}

		if parent != nil {
			parent.navItem.SubItems = append(parent.navItem.SubItems, t.navItem)
		} else {
			c.ng.Items = append(c.ng.Items, t.navItem)
		}
	}
	return nil
}

func (c *Config) buildSidebarNavFromTestGroups(tgs []*httptest.TestGroup, parent *Topic) error {
	for _, tg := range tgs {
		c.idForTestGroup(tg, parent)

		item := new(page.SidebarNavItem)
		item.Text = tg.Desc
		item.Href = c.hrefForTestGroup(tg, parent)

		if parent != nil {
			parent.navItem.SubItems = append(parent.navItem.SubItems, item)
		} else {
			c.ng.Items = append(c.ng.Items, item)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

func (c *Config) buildContent(toc []*TopicGroup) error {
	for _, tg := range toc {
		if err := c.buildContentSectionsFromTopics(tg.Topics, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) buildContentSectionsFromTopics(topics []*Topic, parent *Topic) error {
	for _, t := range topics {
		t.contentSection = new(page.ContentSection)
		t.contentSection.Id = c.idForTopic(t, parent)

		if err := c.buildArticleFromTopic(t, &t.contentSection.Article); err != nil {
			return err
		}
		if err := c.buildContentSectionsFromTopics(t.SubTopics, t); err != nil {
			return err
		}
		if err := c.buildEndpointOverview(t.TestGroups, t.contentSection); err != nil {
			return err
		}
		if err := c.buildContentSectionsFromTestGroups(t.TestGroups, t); err != nil {
			return err
		}

		if parent != nil {
			parent.contentSection.SubSections = append(parent.contentSection.SubSections, t.contentSection)
		} else {
			c.page.Content.Sections = append(c.page.Content.Sections, t.contentSection)
		}
	}
	return nil
}

func (c *Config) buildContentSectionsFromTestGroups(tgs []*httptest.TestGroup, parent *Topic) error {
	for _, tg := range tgs {
		section := new(page.ContentSection)
		section.Id = c.idForTestGroup(tg, parent)

		if err := c.buildArticleFromTestGroup(tg, &section.Article); err != nil {
			return err
		}
		if err := c.buildExamplesFromTestGroup(tg, section); err != nil {
			return err
		}

		if parent != nil {
			parent.contentSection.SubSections = append(parent.contentSection.SubSections, section)
		} else {
			c.page.Content.Sections = append(c.page.Content.Sections, section)
		}
	}
	return nil
}

func (c *Config) buildArticleFromTopic(t *Topic, a *page.Article) error {
	a.Title = t.Name
	a.Href = "#" + t.contentSection.Id

	if t.Text != nil {
		// NOTE(mkopriva): it may be useful to add support for the
		// *html/template.Template type and give the user the ability
		// to provide data through the config.

		switch v := t.Text.(type) {
		case template.HTML:
			a.Doc = v
		case string:
			a.Doc = template.HTML(v)
		case HTML:
			html, err := v.HTML()
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Text:(%T) error: %v", v, err)
			}
			a.Doc = html
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Text:(*os.File) read error: %v", err)
			}
			a.Doc = template.HTML(b)
		default:
			return fmt.Errorf("httpdoc: Topic.Text:(%T) unsupported type", v)
		}

	}

	if t.Parameters != nil {
		typ := c.src.TypeOf(t.Parameters)
		// must be struct, or ptr to struct
		if typ.Kind == types.KindPtr {
			typ = typ.Elem
		}
		if typ.Kind != types.KindStruct {
			return fmt.Errorf("httpdoc: Topic.Parameters:(%T) unsupported type kind", t.Parameters)
		}

		list, err := c.buildFieldListFromType(typ, t, true, nil)
		if err != nil {
			return err
		}
		list.Title = "Parameters"
		a.FieldLists = append(a.FieldLists, list)
	}

	if t.Attributes != nil {
		typ := c.src.TypeOf(t.Attributes)
		// must be struct, or ptr to struct
		if typ.Kind == types.KindPtr {
			typ = typ.Elem
		}
		if typ.Kind != types.KindStruct {
			return fmt.Errorf("httpdoc: Topic.Attributes:(%T) unsupported type kind", t.Attributes)
		}

		list, err := c.buildFieldListFromType(typ, t, false, nil)
		if err != nil {
			return err
		}
		list.Title = "Attributes"
		a.FieldLists = append(a.FieldLists, list)
	}

	if t.Returns != nil {
		a.Conclusion = new(page.Conclusion)
		a.Conclusion.Title = "Returns"

		// NOTE(mkopriva): it may be useful to add support for the
		// *html/template.Template type and give the user the ability
		// to provide data through the config.

		switch v := t.Returns.(type) {
		case template.HTML:
			a.Conclusion.Text = v
		case string:
			a.Conclusion.Text = template.HTML(v)
		case HTML:
			html, err := v.HTML()
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Returns:(%T) error: %v", v, err)
			}
			a.Conclusion.Text = html
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Returns:(*os.File) read error: %v", err)
			}
			a.Conclusion.Text = template.HTML(b)
		default:
			return fmt.Errorf("httpdoc: Topic.Returns:(%T) unsupported type", v)
		}
	}

	return nil
}

func (c *Config) buildArticleFromTestGroup(tg *httptest.TestGroup, a *page.Article) error {
	a.Title = tg.Desc

	//switch text := tg.Text.(type) {
	//case string:
	//	section.Article.HTML = template.HTML(text)
	//}
	return nil
}

func (c *Config) buildFieldListFromType(typ *types.Type, t *Topic, withValidation bool, path []string) (*page.FieldList, error) {
	list := new(page.FieldList)

	tId := c.idForTopic(t, nil)
	tHref := c.hrefForTopic(t, nil)
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
		item.Id = tId + "_" + item.Path + item.Name

		// the field's anchor
		item.Href = tHref
		if i := strings.IndexByte(item.Href, '#'); i > -1 {
			item.Href = item.Href[:i]
		}
		item.Href = item.Href + "#" + item.Id

		// the field's type
		item.Type = f.Type.String()
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
		if len(f.Type.Values) > 0 {
			enumList, err := c.buildEnumListFromType(f.Type)
			if err != nil {
				return nil, err
			}
			item.EnumList = enumList
		}

		// the field's sub fields
		if f.Type.Kind == types.KindStruct && len(f.Type.Fields) > 0 {
			subList, err := c.buildFieldListFromType(f.Type, t, withValidation, append(path, item.Name))
			if err != nil {
				return nil, err
			}
			item.SubFields = subList.Items
		}

		// the field's source link
		if len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
			file := strings.TrimPrefix(f.Pos.Filename, c.ProjectRoot)
			item.SourceLink = c.RepositoryURL + file + "#" + strconv.Itoa(f.Pos.Line)
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

func (c *Config) buildEnumListFromType(typ *types.Type) (*page.EnumList, error) {
	list := new(page.EnumList)
	list.Title = "Possible enum values"
	list.Items = make([]*page.EnumItem, len(typ.Values))

	for i, v := range typ.Values {
		enum := new(page.EnumItem)
		enum.Value = v.Value

		// the internal/types package returns const values of the
		// string kind quoted, unquote the value for display.
		if typ.Kind == types.KindString {
			value, err := strconv.Unquote(enum.Value)
			if err != nil {
				return nil, err
			}
			enum.Value = value
		}

		if len(v.Doc) > 0 {
			html, err := comment.ToHTML(v.Doc)
			if err != nil {
				return nil, err
			}
			enum.Doc = template.HTML(html)
		}

		if len(c.ProjectRoot) > 0 && len(c.RepositoryURL) > 0 {
			file := strings.TrimPrefix(v.Pos.Filename, c.ProjectRoot)
			enum.SourceLink = c.RepositoryURL + file + "#" + strconv.Itoa(v.Pos.Line)
		}

		list.Items[i] = enum
	}

	return list, nil
}

func (c *Config) buildEndpointOverview(tgs []*httptest.TestGroup, section *page.ContentSection) error {
	var overview []page.EndpointOverviewItem

	for _, tg := range tgs {
		item := page.EndpointOverviewItem{}
		item.Href = c.hrefForTestGroup(tg, nil)
		item.Method = tg.Endpoint.Method
		item.Pattern = tg.Endpoint.Pattern
		item.Tooltip = tg.Desc
		overview = append(overview, item)
	}

	section.Example.EndpointOverview = overview
	return nil
}

func (c *Config) buildExamplesFromTestGroup(tg *httptest.TestGroup, parent *page.ContentSection) error {
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
