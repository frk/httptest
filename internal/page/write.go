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
		case EndpointOverviewTest:
			data = p.Content.Sections[0].Example.EndpointOverview
		}

		return t.ExecuteTemplate(w, name, data)
	}

	return t.Execute(w, p)
}

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

var article_field_list = `{{ define "article_field_list" -}}
{{ range . -}}
<div class="article-field-list">
	<h5>{{ .Title }}</h5>
	<ul>
		{{ range .Items -}}
		{{ template "article_field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
{{ end -}}
` //`

var article_field_list_sub = `{{ define "article_field_list_sub" -}}
<div class="article-field-list">
	<ul>
		{{ range . -}}
		{{ template "article_field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

var article_field_item = `{{ define "article_field_item" -}}
<li>
	<h3>
		{{- with .Path }}
		<span class="field-path">{{ . }}</span>
		{{- end }}
		<span class="field-name">{{ .Name }}</span>
		<span class="field-type">{{ .Type }}</span>
	</h3>
	<div class="field-doc">
		{{- with .Text }}
		{{ . }}
		{{- end }}
	</div>

	{{- with .SubFields }}
	{{ template "article_field_list" . }}
	{{- end }}
</li>
{{ end -}}
` //`

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
