package httpdoc

import (
	"bytes"
	"path/filepath"
	"runtime"

	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc/internal/page"
	"github.com/frk/httptest/httpdoc/internal/types"
)

type TopicGroup struct {
	// The name of the group, optional.
	Name string
	// The list of topics that belong to the group.
	Topics []*Topic
}

type Topic struct {
	// The name of the topic.
	Name string
	// ...
	Doc interface{}
	// Type can be used to set the topic's "primary" type.
	// If set to a non-nil value, that value's type information will
	// be used to generate the topic's "primary" type documentation.
	Type interface{}
	// A list of endpoints that are used to generate endpoint specific
	// documentation related to the topic.
	Epts []*httptest.Endpoint
	// The list of sub-topics that belong to this topic.
	SubTopics []*Topic
}

type Config struct {
	p   page.Page
	ng  *page.SidebarNavGroup
	buf bytes.Buffer
}

func (c *Config) Bytes() []byte {
	return c.buf.Bytes()
}

func (c *Config) Build(toc []*TopicGroup) {
	_, f, _, _ := runtime.Caller(1)
	dir := filepath.Dir(f)
	info, err := types.Load(dir)
	if err != nil {
		// ... TODO
	}

	_ = info
	////////////////////////////////////////////////////////////////////////
	// TODO write a main.go (package main) program that imports a package
	// that executes the code below and confirm that the result then is as
	// expected....
	//
	// pkg, err := go/build.ImportDir(dir, build.FindOnly) (*Package, error)
	//
	// ... = gcexportdata.Find(pkg.ImportPath, "")
	////////////////////////////////////////////////////////////////////////

	c.buildSidebar(toc)
	c.buildContent(toc)
}

func (c *Config) buildSidebar(toc []*TopicGroup) {
	for _, g := range toc {
		c.ng = new(page.SidebarNavGroup)
		c.ng.Heading = g.Name

		c.buildSidebarNavFromTopics(g.Topics, nil)

		c.p.Sidebar.NavGroups = append(c.p.Sidebar.NavGroups, c.ng)
	}
}

func (c *Config) buildSidebarNavFromTopics(topics []*Topic, parent *page.SidebarNavItem) {
	for _, t := range topics {
		item := new(page.SidebarNavItem)
		item.Text = t.Name

		if len(t.Epts) > 0 {
			c.buildSidebarNavFromEpts(t.Epts, item)
		}
		if len(t.SubTopics) > 0 {
			c.buildSidebarNavFromTopics(t.SubTopics, item)
		}

		if parent != nil {
			parent.SubItems = append(parent.SubItems, item)
		} else {
			c.ng.Items = append(c.ng.Items, item)
		}
	}
}

func (c *Config) buildSidebarNavFromEpts(epts []*httptest.Endpoint, parent *page.SidebarNavItem) {
	for _, e := range epts {
		item := new(page.SidebarNavItem)

		switch d := e.Doc.(type) {
		case string:
			item.Text = d
		}

		if parent != nil {
			parent.SubItems = append(parent.SubItems, item)
		} else {
			c.ng.Items = append(c.ng.Items, item)
		}
	}
}

func (c *Config) buildContent(toc []*TopicGroup) {
	for _, g := range toc {
		c.buildContentSectionsFromTopics(g.Topics, nil)
	}
}

func (c *Config) buildContentSectionsFromTopics(topics []*Topic, parent *page.ContentSection) {
	for _, t := range topics {
		section := new(page.ContentSection)

		switch d := t.Doc.(type) {
		case string:
			section.Heading = d
		}

		if t.Type != nil {
			section.Body.Text = ""  // TODO type's documentation if any
			section.Body.Type = nil // TODO type's field info if any
		}

		if len(t.Epts) > 0 {
			c.buildContentSectionsFromEpts(t.Epts, section)
		}
		if len(t.SubTopics) > 0 {
			c.buildContentSectionsFromTopics(t.SubTopics, section)
		}

		if parent != nil {
			parent.SubSections = append(parent.SubSections, section)
		} else {
			c.p.Content.Sections = append(c.p.Content.Sections, section)
		}
	}
}

func (c *Config) buildContentSectionsFromEpts(epts []*httptest.Endpoint, parent *page.ContentSection) {
	for _, e := range epts {
		section := new(page.ContentSection)

		switch d := e.Doc.(type) {
		case string:
			section.Heading = d
		}

		if e.Handler != nil {
			section.Body.Text = "" // TODO handler's documentation if any
		}

		c.buildExamplesFromEpt(e, section)

		if parent != nil {
			parent.SubSections = append(parent.SubSections, section)
		} else {
			c.p.Content.Sections = append(c.p.Content.Sections, section)
		}
	}
}

func (c *Config) buildExamplesFromEpt(e *httptest.Endpoint, parent *page.ContentSection) {
	for _, t := range e.Tests {
		if t.Request.Body != nil {
			typ := types.TypeOf(t.Request.Body.Value())
			_ = typ
		}
		if t.Response.Body != nil {
			typ := types.TypeOf(t.Response.Body.Value())
			_ = typ
		}
	}
}
