package page

import (
	"html/template"
	"io"
	"strings"
)

type TestMode string

const (
	SidebarTest       TestMode = "sidebar"
	SidebarHeaderTest TestMode = "sidebar_header"
	SidebarFooterTest TestMode = "sidebar_footer"
	SidebarListsTest  TestMode = "sidebar_lists"

	ContentTest TestMode = "content"

	ArticleTest              TestMode = "article"
	ArticlePrimaryColumnTest TestMode = "article_primary_column"
	ArticleExampleColumnTest TestMode = "article_example_column"

	ArticleSectionListTest TestMode = "article_section_list"
	ArticleTextTest        TestMode = "article_text"
	ArticleAuthInfoTest    TestMode = "article_auth_info"
	ArticleFieldListTest   TestMode = "article_field_list"

	ExampleTextTest      TestMode = "example_text"
	ExampleEndpointsTest TestMode = "example_endpoints"
	ExampleObjectTest    TestMode = "example_object"
	ExampleRequestTest   TestMode = "example_request"
	ExampleResponseTest  TestMode = "example_response"

	FieldItemTest TestMode = "field_item"
	EnumListTest  TestMode = "enum_list"
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
		article_primary_column,
		article_example_column,
		article_section_list,
		article_lead,
		article_text,
		article_auth_info,
		article_field_list,

		field_list,
		field_item,
		field_children,

		enum_list,

		example_section_list,
		example_endpoints,
		example_text,
		example_object,
		example_request,
		example_response,

		code_snippet_lang_list,
		code_snippet_http,
		code_snippet_curl,
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
		case SidebarHeaderTest:
			data = p.Sidebar.Header
		case SidebarFooterTest:
			data = p.Sidebar.Footer
		case SidebarListsTest:
			data = p.Sidebar.Lists
		case ContentTest:
			data = p.Content
		case ArticleTest, ArticlePrimaryColumnTest, ArticleExampleColumnTest:
			data = p.Content.Articles[0]

		// article section tests
		case ArticleSectionListTest:
			data = p.Content.Articles[0].SubArticles[0].Sections
		case ArticleTextTest:
			data = p.Content.Articles[0].SubArticles[0].Sections[0].(*ArticleText)
		case ArticleAuthInfoTest:
			data = p.Content.Articles[0].SubArticles[0].Sections[0].(*ArticleAuthInfo)
		case ArticleFieldListTest:
			data = p.Content.Articles[0].SubArticles[0].Sections[0].(*ArticleFieldList)

		// example section tests
		case ExampleEndpointsTest:
			data = p.Content.Articles[0].Example.Sections[0].(*ExampleEndpoints)
		case ExampleTextTest:
			data = p.Content.Articles[0].SubArticles[0].Example.Sections[0].(*ExampleText)
		case ExampleObjectTest:
			data = p.Content.Articles[0].SubArticles[0].Example.Sections[0].(*ExampleObject)
		case ExampleRequestTest:
			data = p.Content.Articles[0].SubArticles[0].Example.Sections[0].(*ExampleRequest)
		case ExampleResponseTest:
			data = p.Content.Articles[0].SubArticles[0].Example.Sections[1].(*ExampleResponse)

		// single item tests
		case FieldItemTest:
			data = p.Content.Articles[0].SubArticles[0].Sections[0].(*ArticleFieldList).Lists[0].Items[0]
		case EnumListTest:
			data = p.Content.Articles[0].Sections[0].(*ArticleFieldList).Lists[0].Items[0].EnumList
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
{{ $root_url := .RootURL -}}
<header class="sidebar-header">
	{{- with .LogoURL }}
	<div class="sidebar-logo-container">
		<a href="{{ $root_url }}" class="">
			<img src="{{ . }}" alt="" class="logo">
		</a>
	</div>
	{{- end }}
	<h3 class="sidebar-heading">
		<a href="{{ $root_url }}" class="">{{ .Title }}</a>
	</h3>
</header>
{{ end -}}
` //`

var sidebar_footer = `{{ define "sidebar_footer" -}}
<footer class="sidebar-footer">
	{{- with .SigninURL }}
	<div class="signin-container">
		<a href="{{ . }}" class="signin-link">Sign In</a>
	</div>
	{{- end }}
</footer>
{{ end -}}
` //`

var sidebar_lists = `{{ define "sidebar_lists" -}}
<nav class="sidebar">
{{ range . -}}
	<div class="sidebar-list-container">
		<h5 class="sidebar-list-title">{{ .Title }}</h5>
		<ul class="sidebar-list">
			{{ range .Items -}}
			{{ template "sidebar_item" . }}
			{{ end -}}
		</ul>
	</div>
{{ end -}}
</nav>
{{ end -}}
` //`

var sidebar_item = `{{ define "sidebar_item" -}}
<li class="sidebar-list-item">
	<a href="{{ .Href }}" class="{{ .AnchorClass }}">{{ .Text }}</a>
	{{- if .SubItems }}
	<ul class="sidebar-list-child">
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
	{{ template "content_articles" .Articles }}
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
<main role="main">
	<div class="article-column">
		{{ range . -}}
		{{ template "article" . }}
		{{ end -}}
	</div>
</main>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Article
////////////////////////////////////////////////////////////////////////////////

var article = `{{ define "article" -}}
<article id="{{ .Id }}">
	<div class="article-content">
		{{ template "article_primary_column" . }}
		{{ template "article_example_column" . }}
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

var article_primary_column = `{{ define "article_primary_column" -}}
<div class="article-primary-column">
	{{ template "article_lead" . }}
	{{- with .Sections }}
	{{ template "article_section_list" . }}
	{{- end }}
</div>
{{ end -}}
` //`

var article_example_column = `{{ define "article_example_column" -}}
<div class="article-example-column">
	{{- with .Example.Sections }}
	{{ template "example_section_list" . }}
	{{- end }}
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Article Sections
////////////////////////////////////////////////////////////////////////////////

var article_section_list = `{{ define "article_section_list" -}}
{{ range . -}}
{{ if (is_article_text .) -}}
	{{ template "article_text" . }}
{{ else if (is_article_auth_info .) -}}
	{{ template "article_auth_info" . }}
{{ else if (is_article_field_list .) -}}
	{{ template "article_field_list" . }}
{{ end -}}
{{ end -}}
{{ end -}}
` //`

var article_lead = `{{ define "article_lead" -}}
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
	<div class="article-text">
		{{ . }}
	</div>
	{{- end }}
</section>
{{ end -}}
` //`

var article_text = `{{ define "article_text" -}}
<section class="article-section-text">
	<h3 class="article-section-text-title">{{ .Title }}</h3>
	<div class="article-text">
		{{ .Text }}
	</div>
</section>
{{ end -}}
` //`

var article_auth_info = `{{ define "article_auth_info" -}}
<section class="article-section-auth-info">
	<h3 class="article-section-auth-info-title">{{ .Title }}</h3>
	<div class="auth-info-text">
		{{ .Text }}
	</div>
</section>
{{ end -}}
` //`

var article_field_list = `{{ define "article_field_list" -}}
<section class="article-section-field-list">
	<h3 class="article-section-field-list-title">{{ .Title }}</h3>
	{{ range .Lists -}}
	{{ template "field_list" . }}
	{{ end -}}
</section>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Example
////////////////////////////////////////////////////////////////////////////////

var example_section_list = `{{ define "example_section_list" -}}
{{ range . -}}
{{ if (is_example_endpoints .) -}}
	{{ template "example_endpoints" . }}
{{ else if (is_example_text .) -}}
	{{ template "example_text" . }}
{{ else if (is_example_object .) -}}
	{{ template "example_object" . }}
{{ else if (is_example_request .) -}}
	{{ template "example_request" . }}
{{ else if (is_example_response .) -}}
	{{ template "example_response" . }}
{{ end -}}
{{ end -}}
{{ end -}}
` //`

var example_endpoints = `{{ define "example_endpoints" -}}
<section class="example-section-endpoint-list">
	<div class="xs-endpoint-list-container">
		<div class="xs-endpoint-list-topbar">
			<h3 class="xs-endpoint-list-title">{{ .Title }}</h3>
		</div>
		<div class="xs-endpoint-list">
		{{- range .Endpoints }}
			<div class="xs-endpoint-item">
				<a href="{{ .Href }}">
					<span class="xs-endpoint-method method-{{ lower .Method }}"><code>{{ .Method }}</code></span>
					<span class="xs-endpoint-pattern"><code>{{ .Pattern }}</code></span>
				</a>
			</div>
		{{- end }}
		</div>
	</div>
</section>
{{ end -}}
` //`

var example_text = `{{ define "example_text" -}}
<section class="example-section-text">
	{{- with .Title }}
	<h3 class="xs-text-title">{{ . }}</h3>
	{{- end }}
	<div class="xs-text-container">
		{{ .Text }}
	</div>
</section>
{{ end -}}
` //`

var example_object = `{{ define "example_object" -}}
<section class="example-section-object">
	<div class="xs-object-container">
		<div class="xs-object-topbar">
			<h3 class="xs-object-title">{{ .Title }}</h3>
		</div>
		<div class="xs-object-text code-block">
			<div class="code-block-scroll">
				<pre class="code-block-pre">
					<code class="lang-{{ .Lang }}">
						{{ .Text }}
					</code>
				</pre>
			</div>
		</div>
	</div>
</section>
{{ end -}}
` //`

var example_request = `{{ define "example_request" -}}
<section class="example-section-request">
</section>
{{ end -}}
` //`

var example_response = `{{ define "example_response" -}}
<section class="example-section-response">
	<div class="xs-response-container">
		<div class="xs-response-topbar">
			<h3 class="xs-response-title">
				{{ .Title }}:<code class="xs-response-status status-{{ .Status }}"> {{ .Status }}</code>
			</h3>
			{{- with .Header }}
			<ul class="xs-response-header-list">
				{{ range . -}}
				<li class="xs-response-header-item">
					<code class="xs-response-header-key">{{ .Key }}: </code>
					<code class="xs-response-header-value">{{ .Value }}</code>
				</li>
				{{ end -}}
			</ul>
			{{- end }}
		</div>
		<div class="xs-response-body code-block">
			{{- if .Body }}
			<div class="code-block-scroll">
				<pre class="code-block-pre">
					<code class="lang-{{ .Lang }}">
						{{ .Body }}
					</code>
				</pre>
			</div>
			{{- end }}
		</div>
	</div>
</section>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Code Snippets
////////////////////////////////////////////////////////////////////////////////

var code_snippet_lang_list = `{{ define "code_snippet_lang_list" -}}
{{ range . -}}
{{ if (is_code_snippet_http .) -}}
	{{ template "code_snippet_http" . }}
{{ else if (is_code_snippet_curl .) -}}
	{{ template "code_snippet_curl" . }}
{{ end -}}
{{ end -}}
{{ end -}}
` //`

var code_snippet_http = `{{ define "code_snippet_http" -}}
{{ end -}}
` //`

var code_snippet_curl = `{{ define "code_snippet_curl" -}}
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Fields
////////////////////////////////////////////////////////////////////////////////

var field_list = `{{ define "field_list" -}}
<div class="field-list-container">
	{{- with .Title }}
	<h5 class="field-list-heading">{{ . }}</h5>
	{{- end }}
	<ul class="field-list">
		{{ range .Items -}}
		{{ template "field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

var field_item = `{{ define "field_item" -}}
<li id="{{ .Id }}" class="field-item">
	<h3 class="field-heading">
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
	<div class="field-text-container">
		{{- with .Text }}
		<div class="field-text">
			{{ . }}
		</div>
		{{- end }}
	</div>

	{{- with .EnumList }}
	{{ template "enum_list" . }}
	{{- end }}
	{{- with .SubFields }}
	{{ template "field_children" . }}
	{{- end }}
</li>
{{ end -}}
` //`

var field_children = `{{ define "field_children" -}}
<div class="field-list-container">
	<ul class="field-list">
		{{ range . -}}
		{{ template "field_item" . }}
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Enum List
////////////////////////////////////////////////////////////////////////////////

var enum_list = `{{ define "enum_list" -}}
<div class="enum-list-container">
	<h5 class="enum-list-heading">{{ .Title }}</h5>
	<ul class="enum-list">
		{{ range .Items -}}
		<li class="enum-item">
			<div class="enum-heading">
				<div class="enum-value">
					<code>{{ .Value }}</code>
				</div>
				<div class="enum-source-link">
					{{- with .SourceLink }}
					<a href="{{ .Href }}">‹›</a>
					{{- end }}
				</div>
			</div>
			<div class="enum-text">
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
