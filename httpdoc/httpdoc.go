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

const DefaultFieldNameTagKey = "json"

type Config struct {
	// The tag key to be used to retrieve a field's name, defaults to "json".
	//
	// If no name is present in the tag value associated with the key,
	// the field's name will be used as fallback.
	FieldNameTagKey string
	// FieldTypeName returns the name for a specific field's type based on
	// given reflect.StructField value.
	//
	// If FieldTypeName is nil or it returns false as the second return
	// value (ok) it will fall back to the default behaviour.
	FieldTypeName func(reflect.StructField) (typeName string, ok bool)

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

	if len(c.FieldNameTagKey) == 0 {
		c.FieldNameTagKey = DefaultFieldNameTagKey
	}

	////////////////////////////////////////////////////////////////////////
	// TODO write a main.go (package main) program that imports a package
	// that executes the code below and confirm that the result then is as
	// expected....
	//
	// pkg, err := go/build.ImportDir(dir, build.FindOnly) (*Package, error)
	//
	// ... = gcexportdata.Find(pkg.ImportPath, "")
	////////////////////////////////////////////////////////////////////////

	if err := c.buildSidebar(toc); err != nil {
		return err
	}
	if err := c.buildContent(toc); err != nil {
		return err
	}
	if err := page.Write(&c.buf, c.page, c.mode); err != nil {
		return err
	}

	// ...

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

func (c *Config) buildSidebar(toc []*TopicGroup) error {
	for _, tg := range toc {
		c.ng = new(page.SidebarNavGroup)
		c.ng.Heading = tg.Name

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
	a.Heading = t.Name

	// TODO href

	if t.Text != nil {
		// NOTE(mkopriva): it may be useful to add support for the
		// *html/template.Template type and give the user the ability
		// to provide data through the config.

		switch v := t.Text.(type) {
		case template.HTML:
			a.Text = v
		case string:
			a.Text = template.HTML(v)
		case HTML:
			html, err := v.HTML()
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Text:(%T) error: %v", v, err)
			}
			a.Text = html
		case *os.File:
			b, err := ioutil.ReadAll(v)
			if err != nil {
				return fmt.Errorf("httpdoc: Topic.Text:(*os.File) read error: %v", err)
			}
			a.Text = template.HTML(b)
		default:
			return fmt.Errorf("httpdoc: Topic.Text:(%T) unsupported type", v)
		}

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

		list, err := c.buildFieldListFromType(typ, t, nil)
		if err != nil {
			return err
		}
		list.Title = "Attributes"
		a.FieldLists = append(a.FieldLists, list)
	}

	// if t.Type != nil {
	// 	section.Article.Text = ""  // TODO type's documentation if any
	// 	section.Article.Type = nil // TODO type's field info if any
	// }
	return nil
}

func (c *Config) buildArticleFromTestGroup(tg *httptest.TestGroup, a *page.Article) error {
	a.Heading = tg.Desc

	//switch text := tg.Text.(type) {
	//case string:
	//	section.Article.HTML = template.HTML(text)
	//}
	return nil
}

func (c *Config) buildFieldListFromType(typ *types.Type, t *Topic, path []string) (*page.FieldList, error) {
	list := new(page.FieldList)

	tId := c.idForTopic(t, nil)
	tHref := c.hrefForTopic(t, nil)
	tagKey := c.FieldNameTagKey
	for _, f := range typ.Fields {
		tag := tagutil.New(f.Tag)

		// skip field?
		if tag.Contains(tagKey, "-") || tag.Contains("doc", "-") {
			continue
		}

		var fieldName = f.Name
		if name := tag.First(tagKey); name != "" {
			fieldName = name
		}

		var fieldType = f.Type.String()
		if c.FieldTypeName != nil {
			sf, ok := typ.ReflectType.FieldByName(f.Name)
			if !ok {
				// this should not happen
				panic(fmt.Sprintf("httpdoc: reflect.Type.FieldByName(%q) failed.", f.Name))
			}
			if name, ok := c.FieldTypeName(sf); len(name) > 0 || ok {
				fieldType = name
			}
		}

		var fieldDoc string
		if len(f.Doc) > 0 {
			html, err := comment.ToHTML(f.Doc)
			if err != nil {
				return nil, err
			}
			fieldDoc = html
		}

		var fieldPath string
		if len(path) > 0 {
			fieldPath = strings.Join(path, ".") + "."
		}

		var fieldId = tId + "_" + fieldPath + fieldName
		var fieldHref = tHref
		if i := strings.IndexByte(fieldHref, '#'); i > -1 {
			fieldHref = fieldHref[:i]
		}
		fieldHref = fieldHref + "#" + fieldId

		var subFields []*page.FieldListItem
		if f.Type.Kind == types.KindStruct {
			list, err := c.buildFieldListFromType(f.Type, t, append(path, fieldName))
			if err != nil {
				return nil, err
			}
			subFields = list.Items
		}

		if false { // Parameters?
			// TODO required directive
			// TODO validation directive
		}

		item := new(page.FieldListItem)
		item.Id = fieldId
		item.Href = fieldHref
		item.Name = fieldName
		item.Type = fieldType
		item.Text = template.HTML(fieldDoc)
		item.Path = fieldPath
		item.SubFields = subFields
		list.Items = append(list.Items, item)
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
