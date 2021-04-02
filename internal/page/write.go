package page

import (
	"html/template"
	"io"
	"strings"
)

type TestMode string

const (
	SidebarTest          TestMode = "sidebar"
	ContentTest          TestMode = "content"
	ArticleTest          TestMode = "article"
	ArticleFieldListTest TestMode = "article_field_list"
	ArticleFieldItemTest TestMode = "article_field_item"
	EndpointOverviewTest TestMode = "endpoint_overview"
)

func Write(w io.Writer, p Page, m TestMode) error {
	tmpl := strings.Join([]string{
		page_root,
		sidebar,
		sidebar_header,
		sidebar_nav,
		sidebar_nav_item,
		content,
		content_header,
		content_sections,
		content_footer,
		article,
		article_field_list,
		article_field_list_sub,
		article_field_item,
		field_enum_list,
		example,
		endpoint_overview,
	}, "")

	t, err := template.New("t").Funcs(helpers).Parse(tmpl)
	if err != nil {
		return err
	}

	// NOTE(mkopriva): This is used purely for testing. If a non-zero TestMode
	// flag is passed in it is assumed that Write was executed inside a test.
	//
	// It is expected that the part of the Page identified by the TestMode
	// flag has been properly initialized by the test, if not the program may crash.
	if len(m) > 0 {
		name := string(m)
		data := interface{}(p)

		switch m {
		case SidebarTest:
			data = p.Sidebar
		case ContentTest:
			data = p.Content
		case ArticleTest:
			data = p.Content.Sections[0].Article
		case ArticleFieldListTest:
			data = p.Content.Sections[0].Article.FieldLists
		case ArticleFieldItemTest:
			data = p.Content.Sections[0].Article.FieldLists[0].Items[0]
		case EndpointOverviewTest:
			data = p.Content.Sections[0].Example.EndpointOverview
		}

		return t.ExecuteTemplate(w, name, data)
	}

	return t.Execute(w, p)
}

////////////////////////////////////////////////////////////////////////////////
// Page
////////////////////////////////////////////////////////////////////////////////

var page_root = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>{{ .Title }}</title>
	</head>
	<body>

		{{ template "sidebar" .Sidebar }}
		{{ template "content" .Content }}

		<footer>
		</footer>
	</body>
</html>
` //`

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

var sidebar = `{{ define "sidebar" -}}
<div class="sidebar">
	{{ template "sidebar_header" .Header }}
	{{ template "sidebar_nav" .NavGroups }}
</div>
{{ end }}
` //`

var sidebar_header = `{{ define "sidebar_header" -}}
<div class="sidebar-header">
	<h3 class="sidebar-heading">{{ .Title }}</h3>
</div>
{{- end }}
` //`

var sidebar_nav = `{{ define "sidebar_nav" -}}
<nav class="sidebar-nav">
{{ range . -}}
	<div class="nav-group">
		<h5 class="nav-heading">{{ .Heading }}</h5>
		<ul class="nav-items">
			{{ range .Items }}{{ template "sidebar_nav_item" . }}{{ end -}}
		</ul>
	</div>
{{ end -}}
</nav>
{{- end }}
` //`

var sidebar_nav_item = `{{ define "sidebar_nav_item" -}}
<li>
	<a href="{{ .Href }}" class="nav-item expandable">{{ .Text }}</a>
	{{- if .SubItems }}
	<ul class="nav-items">
		{{ range .SubItems }}{{ template "sidebar_nav_item" . }}{{ end -}}
	</ul>
	{{- end }}
</li>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

var content = `{{ define "content" -}}
<div class="content">
	{{ template "content_header" .Header }}
	{{ template "content_sections" .Sections }}
	{{ template "content_footer" .Footer }}
</div>
{{ end }}
` //`

var content_header = `{{ define "content_header" -}}
<div class="content-header">
</div>
{{- end }}
` //`

var content_sections = `{{ define "content_sections" -}}
<div class="content-sections">
	{{ range . -}}
	<section id="{{ .Id }}">
		{{ template "article" .Article }}
		{{ template "example" .Example }}

		{{- with .SubSections }}
		{{ template "content_sections" . }}
		{{- end }}
	</section>
	{{ end -}}
</div>
{{- end }}
` //`

var content_footer = `{{ define "content_footer" -}}
<div class="content-footer">
</div>
{{- end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Article
////////////////////////////////////////////////////////////////////////////////

var article = `{{ define "article" -}}
<div class="article">
	<h1>{{ .Heading }}</h1>

	{{- with .Text }}
	<div class="article-text">
		{{ . }}
	</div>
	{{- end }}

	{{- with .FieldLists }}
	{{ template "article_field_list" . }}
	{{- end }}
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Article Fields
////////////////////////////////////////////////////////////////////////////////

var article_field_list = `{{ define "article_field_list" -}}
{{ range . -}}
<div class="article-field-list-container">
	<h5 class="article-field-list-header">{{ .Title }}</h5>
	<ul class="article-field-list">
		{{ range .Items -}}
		{{ template "article_field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
{{ end -}}
` //`

var article_field_list_sub = `{{ define "article_field_list_sub" -}}
<div class="article-field-list-container">
	<ul class="article-field-list">
		{{ range . -}}
		{{ template "article_field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

var article_field_item = `{{ define "article_field_item" -}}
<li id="{{ .Id }}">
	<h3 class="field-header">
		<a class="field-anchor" href="{{ .Href }}">¶</a>
		{{- with .Path }}
		<span class="field-path">{{ . }}</span>
		{{- end }}
		<span class="field-name">{{ .Name }}</span>
		<span class="field-type">{{ .Type }}</span>
		{{- with .SourceLink }}
		<a class="field-source-link" href="{{ . }}">‹›</a>
		{{- end }}
	</h3>
	<div class="field-doc">
		{{- with .Text }}
		<div class="field-doc-text">
			{{ . }}
		</div>
		{{- end }}
	</div>

	{{- with .EnumList }}
	{{ template "field_enum_list" . }}
	{{- end }}
	{{- with .SubFields }}
	{{ template "article_field_list_sub" . }}
	{{- end }}
</li>
{{ end -}}
` //`

var field_enum_list = `{{ define "field_enum_list" -}}
<div class="field-enum-list-container">
	<h5 class="field-enum-list-header">{{ .Title }}</h5>
	<ul class="field-enum-list">
		{{ range .Items -}}
		<li class="field-enum-item">
			<div class="field-enum-item-header">
				<div class="field-enum-item-value">
					<code>{{ .Value }}</code>
				</div>
				<div class="field-enum-item-source-link">
					{{- with .SourceLink }}
					<a href="{{ . }}">‹›</a>
					{{- end }}
				</div>
			</div>
			<div class="field-enum-item-text">
				{{- with .Text }}
				{{ . }}
				{{- end }}
			</div>
		</li>
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Example
////////////////////////////////////////////////////////////////////////////////

var example = `{{ define "example" -}}
<div class="example">
	{{- if .EndpointOverview }}
	{{ template "endpoint_overview" .EndpointOverview }}
	{{- end }}
</div>
{{- end -}}
` //`

var endpoint_overview = `{{ define "endpoint_overview" -}}
<div class="endpoint-overview">
	<div class="ep-ov-topbar">ENDPOINTS</div>
	<div class="ep-ov-table">
	{{- range . }}
		<div class="ep-ov-row">
			<a href="{{ .Href }}">
				<span class="ep-ov-method method-{{ lower .Method }}"><code>{{ .Method }}</code></span>
				<span class="ep-ov-pattern"><code>{{ .Pattern }}</code></span>
			</a>
		</div>
	{{- end }}
	</div>
</div>
{{ end -}}
` //`
