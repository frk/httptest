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
	FieldListTest        TestMode = "field_list"
	FieldItemTest        TestMode = "field_item"
	EndpointOverviewTest TestMode = "endpoint_overview"
)

func Write(w io.Writer, p Page, m TestMode) error {
	tmpl := strings.Join([]string{
		page_root,

		sidebar,
		sidebar_header,
		sidebar_footer,
		sidebar_lists,
		sidebar_item,

		content,
		content_header,
		content_footer,
		content_articles,

		article,
		article_section_lead,
		article_section_list,

		field_list,
		field_list_sub,
		field_item,

		value_list,

		example_section_list,
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
			data = p.Content.Articles[0]
		case FieldListTest:
			data = p.Content.Articles[0].Sections[0].(*FieldListArticleSection).Lists
		case FieldItemTest:
			data = p.Content.Articles[0].Sections[0].(*FieldListArticleSection).Lists[0].Items[0]
		case EndpointOverviewTest:
			data = p.Content.Articles[0].Example.Sections[0].(*EndpointsExampleSection)
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
	</body>
</html>
` //`

////////////////////////////////////////////////////////////////////////////////
// Sidebar
////////////////////////////////////////////////////////////////////////////////

var sidebar = `{{ define "sidebar" -}}
<div class="sidebar-container">
	{{ template "sidebar_header" .Header }}
	{{ template "sidebar_lists" .Lists }}
	{{ template "sidebar_footer" .Footer }}
</div>
{{ end }}
` //`

var sidebar_header = `{{ define "sidebar_header" -}}
<header class="sidebar-header">
	<h3 class="sidebar-heading">{{ .Title }}</h3>
</header>
{{ end -}}
` //`

var sidebar_footer = `{{ define "sidebar_footer" -}}
<footer class="sidebar-footer">
</footer>
{{ end -}}
` //`

var sidebar_lists = `{{ define "sidebar_lists" -}}
<nav class="sidebar">
{{ range . -}}
	<div class="sidebar-list-container">
		<h5 class="sidebar-list-title">{{ .Title }}</h5>
		<ul class="sidebar-list">
			{{ range .Items }}{{ template "sidebar_item" . }}{{ end -}}
		</ul>
	</div>
{{ end -}}
</nav>
{{ end -}}
` //`

var sidebar_item = `{{ define "sidebar_item" -}}
<li class="sidebar-list-item">
	<a href="{{ .Href }}" class="sidebar-item expandable">{{ .Text }}</a>
	{{- if .SubItems }}
	<ul class="sidebar-list-sub">
		{{ range .SubItems }}{{ template "sidebar_item" . }}{{ end -}}
	</ul>
	{{- end }}
</li>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Content
////////////////////////////////////////////////////////////////////////////////

var content = `{{ define "content" -}}
<div class="content-container">
	{{ template "content_header" .Header }}
	<main role="main">
		{{ template "content_articles" .Articles }}
	</main>
	{{ template "content_footer" .Footer }}
</div>
{{ end }}
` //`

var content_header = `{{ define "content_header" -}}
<header class="content-header">
</header>
{{ end -}}
` //`

var content_footer = `{{ define "content_footer" -}}
<footer class="content-footer">
</footer>
{{ end -}}
` //`

var content_articles = `{{ define "content_articles" -}}
<div class="articles-container">
	{{ range . -}}
	{{ template "article" . }}
	{{ end -}}
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Article
////////////////////////////////////////////////////////////////////////////////

var article = `{{ define "article" -}}
<article id="{{ .Id }}">
	<div class="article-content">
		<div class="article-text-column">
			{{ template "article_section_lead" . }}
			{{- with .Sections }}
			{{ template "article_section_list" . }}
			{{- end }}
		</div>
		<div class="article-code-column">
			{{- with .Example.Sections }}
			{{ template "example_section_list" . }}
			{{- end }}
		</div>
	</div>

	{{- with .SubArticles }}
	<div class="article-children">
		{{ range . -}}
		{{ template "article" . }}
		{{ end -}}
	</div>
	{{- end }}
</article>
{{ end -}}
` //`

var article_section_lead = `{{ define "article_section_lead" -}}
<section class="article-section-lead">
	<h2 class="article-section-lead-title">
		<a class="article-anchor" href="{{ .Href }}">{{ .Title }}</a>
	</h2>

	{{- with .SourceLink }}
	<div class="article-source">
		<a class="article-source-link" href="{{ .Href }}">{{ .Text }}</a>
	</div>
	{{- end }}

	{{- with .Text }}
	<div class="article-doc">
		{{ . }}
	</div>
	{{- end }}
</section>
{{ end -}}
` //`

var article_section_list = `{{ define "article_section_list" -}}
{{ range . -}}
{{ if (is_text_article_section .) -}}
<section class="article-section-text">
	<h3 class="article-section-text-title">{{ .Title }}</h3>
	<div class="article-doc">
		{{ .Text }}
	</div>
</section>
{{ else if (is_field_list_article_section .) }}
<section class="article-section-fields">
	<h3 class="article-section-fields-title">{{ .Title }}</h3>
	{{ range .Lists -}}
	{{ template "field_list" . }}
	{{ end -}}
</section>
{{ end -}}
{{ end -}}
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Fields
////////////////////////////////////////////////////////////////////////////////

var field_list = `{{ define "field_list" -}}
{{ range . -}}
<div class="field-list-container">
	{{- with .Title }}
	<h5 class="field-list-header">{{ . }}</h5>
	{{- end }}
	<ul class="field-list">
		{{ range .Items -}}
		{{ template "field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
{{ end -}}
` //`

var field_list_sub = `{{ define "field_list_sub" -}}
<div class="field-list-container">
	<ul class="field-list">
		{{ range . -}}
		{{ template "field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

var field_item = `{{ define "field_item" -}}
<li id="{{ .Id }}">
	<h3 class="field-header">
		<a class="field-anchor" href="{{ .Href }}">¶</a>
		{{- with .Path }}
		<span class="field-path">{{ . }}</span>
		{{- end }}
		<span class="field-name">{{ .Name }}</span>
		<span class="field-type">{{ .Type }}</span>
		{{- if .SettingText }}
		<span class="field-setting-{{ .SettingLabel }}">{{ .SettingText }}</span>
		{{- end }}
		{{- with .SourceLink }}
		<a class="field-source-link" href="{{ .Href }}">‹›</a>
		{{- end }}
	</h3>
	{{- with .Validation }}
	<div class="field-validation">
		{{ . }}
	</div>
	{{- end }}
	<div class="field-doc">
		{{- with .Doc }}
		<div class="field-doc-text">
			{{ . }}
		</div>
		{{- end }}
	</div>

	{{- with .ValueList }}
	{{ template "value_list" . }}
	{{- end }}
	{{- with .SubFields }}
	{{ template "field_list_sub" . }}
	{{- end }}
</li>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Example
////////////////////////////////////////////////////////////////////////////////

var example_section_list = `{{ define "example_section_list" -}}
{{ range . -}}
{{ if .Text }}
<section class="example-section-text">
	{{- with .Title }}
	<h3 class="example-section-text-title">{{ . }}</h3>
	{{- end }}
	<div class="example-doc">
		{{ .Text }}
	</div>
</section>
{{ else if .EndpointOverview }}
<section class="example-section-endpoint-overview">
	{{ template "endpoint_overview" .EndpointOverview }}
</section>
{{ end -}}
{{ end -}}
{{ end -}}
` //`

var endpoint_overview = `{{ define "endpoint_overview" -}}
<div class="endpoint-overview-container">
	<div class="endpoint-overview-topbar">
		<h3 class="endpoint-overview-title">{{ .Title }}</h3>
	</div>
	<div class="endpoint-overview-table">
	{{- range .Items }}
		<div class="endpoint-overview-row">
			<a href="{{ .Href }}">
				<span class="endpoint-overview-method method-{{ lower .Method }}"><code>{{ .Method }}</code></span>
				<span class="endpoint-overview-pattern"><code>{{ .Pattern }}</code></span>
			</a>
		</div>
	{{- end }}
	</div>
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Value List
////////////////////////////////////////////////////////////////////////////////

var value_list = `{{ define "value_list" -}}
{{ $class := .Class -}}
<div class="{{ $class }}-list-container">
	<h5 class="{{ $class }}-list-header">{{ .Title }}</h5>
	<ul class="{{ $class }}-list">
		{{ range .Items -}}
		<li class="{{ $class }}-item">
			<div class="{{ $class }}-item-header">
				<div class="{{ $class }}-item-value">
					<code>{{ .Text }}</code>
				</div>
				<div class="{{ $class }}-item-source-link">
					{{- with .SourceLink }}
					<a href="{{ .Href }}">‹›</a>
					{{- end }}
				</div>
			</div>
			<div class="{{ $class }}-item-doc">
				{{- with .Doc }}
				{{ . }}
				{{- end }}
			</div>
		</li>
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`
