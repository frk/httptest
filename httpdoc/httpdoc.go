package httpdoc

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/comment"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/types"
)

type Config struct {
	//
	buf bytes.Buffer
	// source code info
	src *types.Source
	// the page being built
	page page.Page
	// the nav group currently being built, or nil
	ng *page.SidebarNavGroup
	// set of already generated hrefs
	hrefs map[string]int
	// map of endpoints to their respective hrefs
	ep2href map[httptest.Endpoint]string
	// set of already generated ids for html elements
	ids map[string]int
	// utilized by tests
	mode page.TestMode
}

func (c *Config) Build(toc []*TopicGroup) error {
	c.hrefs = make(map[string]int)
	c.ep2href = make(map[httptest.Endpoint]string)
	c.ids = make(map[string]int)

	_, f, _, _ := runtime.Caller(1)
	dir := filepath.Dir(f)
	src, err := types.Load(dir)
	if err != nil {
		return err
	}
	c.src = src

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

func (c *Config) buildSidebarNavFromTopics(topics []*Topic, parent *page.SidebarNavItem) error {
	for _, t := range topics {
		item := new(page.SidebarNavItem)
		item.Text = t.Name
		item.Href = c.hrefFromTopic(t, parent)

		if err := c.buildSidebarNavFromTestGroups(t.TestGroups, item); err != nil {
			return err
		}
		if err := c.buildSidebarNavFromTopics(t.SubTopics, item); err != nil {
			return err
		}

		if parent != nil {
			parent.SubItems = append(parent.SubItems, item)
		} else {
			c.ng.Items = append(c.ng.Items, item)
		}
	}
	return nil
}

func (c *Config) buildSidebarNavFromTestGroups(tgs []*httptest.TestGroup, parent *page.SidebarNavItem) error {
	for _, tg := range tgs {
		item := new(page.SidebarNavItem)
		item.Text = tg.Desc
		item.Href = c.hrefFromEndpoint(tg.Endpoint)

		if parent != nil {
			parent.SubItems = append(parent.SubItems, item)
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

func (c *Config) buildContentSectionsFromTopics(topics []*Topic, parent *page.ContentSection) error {
	for _, t := range topics {
		section := new(page.ContentSection)
		section.Id = c.idFromTopic(t, parent)

		if err := c.buildArticleFromTopic(t, &section.Article); err != nil {
			return err
		}
		if err := c.buildContentSectionsFromTopics(t.SubTopics, section); err != nil {
			return err
		}
		if err := c.buildEndpointOverview(t.TestGroups, section); err != nil {
			return err
		}
		if err := c.buildContentSectionsFromTestGroups(t.TestGroups, section); err != nil {
			return err
		}

		if parent != nil {
			parent.SubSections = append(parent.SubSections, section)
		} else {
			c.page.Content.Sections = append(c.page.Content.Sections, section)
		}
	}
	return nil
}

func (c *Config) buildContentSectionsFromTestGroups(tgs []*httptest.TestGroup, parent *page.ContentSection) error {
	for _, tg := range tgs {
		section := new(page.ContentSection)
		section.Id = c.idFromTestGroup(tg, parent)

		if err := c.buildArticleFromTestGroup(tg, &section.Article); err != nil {
			return err
		}
		if err := c.buildExamplesFromTestGroup(tg, section); err != nil {
			return err
		}

		if parent != nil {
			parent.SubSections = append(parent.SubSections, section)
		} else {
			c.page.Content.Sections = append(c.page.Content.Sections, section)
		}
	}
	return nil
}

func (c *Config) buildArticleFromTopic(t *Topic, a *page.Article) error {
	a.Heading = t.Name

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

		list, err := c.buildFieldListFromType("Attributes", typ)
		if err != nil {
			return err
		}
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

func (c *Config) buildFieldListFromType(title string, typ *types.Type) (*page.FieldList, error) {
	list := new(page.FieldList)
	list.Title = title

	for _, f := range typ.Fields {
		var text string
		if len(f.Doc) > 0 {
			html, err := comment.ToHTML(f.Doc)
			if err != nil {
				return nil, err
			}
			text = html
		}

		item := new(page.FieldListItem)
		item.Name = f.Name
		item.Path = ""
		item.Type = f.Type.GetName()
		item.Text = template.HTML(text)
		list.Items = append(list.Items, item)
	}

	return list, nil
}

func (c *Config) buildEndpointOverview(tgs []*httptest.TestGroup, section *page.ContentSection) error {
	var overview []page.EndpointOverviewItem

	for _, tg := range tgs {
		item := page.EndpointOverviewItem{}
		item.Href = c.hrefFromEndpoint(tg.Endpoint)
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

// idFromTopic returns an "id" tag attribute for the given topic.
func (c *Config) idFromTopic(t *Topic, parent *page.ContentSection) string {
	prefix := ""
	if parent != nil {
		prefix = parent.Id
	}

	id := strings.Map(func(r rune) rune {
		// TODO(mkopriva): handle non ascii characters, e.g. japanese chinese, arabic, etc.
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(t.Name))

	if len(prefix) > 0 {
		id = prefix + "-" + id
	}

	// make sure the returned id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}
	return id
}

// idFromTestGroup returns an "id" tag attribute for the given topic.
func (c *Config) idFromTestGroup(tg *httptest.TestGroup, parent *page.ContentSection) string {
	prefix := ""
	if parent != nil {
		prefix = parent.Id
	}

	id := strings.Map(func(r rune) rune {
		// TODO(mkopriva): handle non ascii characters, e.g. japanese chinese, arabic, etc.
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(tg.Desc))

	if len(prefix) > 0 {
		id = prefix + "-" + id
	}

	// make sure the returned id is unique
	count := c.ids[id]
	c.ids[id] = count + 1
	if count > 0 {
		id += "-" + strconv.Itoa(count+1)
	}
	return id
}

// hrefFromTopic returns an href string for the given topic.
func (c *Config) hrefFromTopic(t *Topic, parent *page.SidebarNavItem) string {
	prefix := ""
	if parent != nil {
		prefix = parent.Href
	}

	href := strings.Map(func(r rune) rune {
		// TODO(mkopriva): handle non ascii characters, e.g. japanese chinese, arabic, etc.
		//
		// Whatever is valid in path (without the need of encoding it)
		// should be left unchanged, the rest should be replaced by a hyphen.
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, strings.ToLower(t.Name))
	href = prefix + "/" + href

	// make sure the returned href is unique
	count := c.hrefs[href]
	c.hrefs[href] = count + 1
	if count > 0 {
		href += "-" + strconv.Itoa(count+1)
	}
	return href
}

// hrefFromEndpoint returns an href string for the given test group.
func (c *Config) hrefFromEndpoint(ep httptest.Endpoint) string {
	if href, ok := c.ep2href[ep]; ok {
		return href
	}

	method, pattern := ep.Method, ep.Pattern
	pattern = strings.Map(func(r rune) rune {
		// TODO(mkopriva): allow for handling of more complex endpoint
		// patterns, i.e. some routers allow specifying regular expressions,
		// or they use delimiters different from "{" and "}".
		//
		// Consider providing a Config.<Field> so that the user can supply
		// their own "hrefFromEndpoint" implementation based on their own need.
		if r == '{' || r == '}' {
			return -1
		}
		return r
	}, pattern)

	href := strings.ToLower(pattern + "/" + method)

	// make sure the returned href is unique
	count := c.hrefs[href]
	c.hrefs[href] = count + 1
	if count > 0 {
		href += "-" + strconv.Itoa(count+1)
	}

	c.ep2href[ep] = href
	return href
}
