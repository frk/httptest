package page

import (
	"html/template"
	"strings"
)

var T = template.Must(template.New("t").Funcs(helpers).Parse(strings.Join([]string{
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
	example_response,
	example_request,
	example_request_topbar,
	example_request_body,

	code_snippet_http,
	code_snippet_curl,

	code_block_pre,
	svg_code_icon,
	svg_code_icon_use,
	curl_data,
}, "")))

////////////////////////////////////////////////////////////////////////////////
// Page
////////////////////////////////////////////////////////////////////////////////

var page_root = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>{{ .Title }}</title>
		<link rel="stylesheet" href="/assets/css/main.css?{{ url .RandomHash }}">
		{{- if .AddCustomCSS }}
		<link rel="stylesheet" href="/assets/css/custom.css?{{ url .RandomHash }}">
		{{- end }}
	</head>
	<body>
		{{ template "svg_code_icon" }}

		{{ template "sidebar" .Sidebar }}
		{{ template "content" .Content }}

		{{ WITH_ID }}
		<script type="text/javascript">
			document.getElementById('{{ DOT }}').scrollIntoView();
		</script>
		{{ END }}

		<script type="text/javascript" src="/assets/js/main.js?{{ url .RandomHash }}"></script>
		<script type="text/javascript">
			httpdoc.init({
				lang: '{{ GET_LANG }}',
			});
		</script>
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
	<div class="sidebar-banner-box">
		{{- with (is_sidebar_banner_title .Banner) }}
		<h3 class="sidebar-heading">
			<a href="{{ .URL }}" class="">{{ .Text }}</a>
		</h3>
		{{- end }}

		{{- with (is_sidebar_banner_html .Banner) }}
			{{ .Text }}
		{{- end }}
	</div>
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
			{{ range .Items -}}
			{{ template "sidebar_item" . }}
			{{ end -}}
		</ul>
	</div>
{{ end -}}
</nav>
{{ end -}}
` //`

// NOTE(mkopriva): The correlated JavaScript relies on the structure of the
// sidebar_item DOM. Specifically it expects that the ul.sidebar-item-subitems
// is a direct child of the li.sidebar-list-item element. If the sidebar_item
// DOM is changed the JavaScript will most probably have to be changed too.
var sidebar_item = `{{ define "sidebar_item" -}}
<li class="{{ .ListItemClass }}{{ IS_ACTIVE .Href }} active{{ END }}" data-anchor="{{ .Anchor }}">
	<a href="{{ .Href }}" class="sidebar-item">{{ .Text }}</a>
	{{- if .SubItems }}
	<ul class="sidebar-item-subitems{{ IS_HIDDEN .Href }} hidden{{ END }}">
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
	{{- if .IsRoot }}
	<h2 class="article-section-lead-title">
		<a class="article-anchor" href="{{ .Href }}">{{ .Title }}</a>
	</h2>
	{{- else }}
	<h3 class="article-section-lead-title">
		<a class="article-anchor" href="{{ .Href }}">{{ .Title }}</a>
	</h3>
	{{- end }}

	{{- with .SourceLink }}
	<div class="article-source">
		<a class="article-source-link" href="{{ .Href }}">{{ .Text }}</a>
	</div>
	{{- end }}

	{{- with .Text }}
	<div class="article-text text-box">
		{{ . }}
	</div>
	{{- end }}
</section>
{{ end -}}
` //`

var article_text = `{{ define "article_text" -}}
<section class="article-section-text">
	<h3 class="article-section-text-title">{{ .Title }}</h3>
	<div class="article-text text-box">
		{{ .Text }}
	</div>
</section>
{{ end -}}
` //`

var article_auth_info = `{{ define "article_auth_info" -}}
<section class="article-section-auth-info">
	<h3 class="article-section-auth-info-title">{{ .Title }}</h3>
	<div class="auth-info-text text-box">
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

// NOTE(mkopriva): The correlated JavaScript relies on the structure of the
// xs-endpoint-item DOM. Specifically it expects that the div.xs-endpoint-item
// has the anchor as the first child. If the xs-endpoint-item DOM is changed the
// JavaScript will most probably have to be changed too.
var example_endpoints = `{{ define "example_endpoints" -}}
<section class="example-section-endpoint-list">
	<div class="xs-endpoint-list-container">
		<div class="xs-endpoint-list-topbar">
			<h3 class="xs-endpoint-list-title">{{ .Title }}</h3>
		</div>
		<div class="xs-endpoint-list">
		{{- range .Endpoints }}
			<div class="xs-endpoint-item" data-tooltip="{{ .Tooltip }}">
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
	<div class="xs-text-container text-box">
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
				{{ template "code_block_pre" . }}
			</div>
		</div>
	</div>
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
			{{- if .Code }}
			<div class="code-block-scroll">
				{{ template "code_block_pre" . }}
			</div>
			{{- end }}
		</div>
	</div>
</section>
{{ end -}}
` //`

var example_request = `{{ define "example_request" -}}
<section class="example-section-request">
	<div class="xs-request-container">
		{{ template "example_request_topbar" . }}
		{{ template "example_request_body" . }}
	</div>
</section>
{{ end -}}
` //`

var example_request_topbar = `{{ define "example_request_topbar" -}}
<div class="xs-request-topbar">
	<div class="xs-request-title-container">
		<h3 class="xs-request-title">
			<code>
				<span class="xs-request-endpoint-method {{ lower .Method }}">{{ .Method }} </span>
				<span class="xs-request-endpoint-pattern">{{ .Pattern }}</span>
			</code>
		</h3>
	</div>
	<div class="xs-request-lang-select-container">
		<select name="lang" autocomplete="off">
		{{ range .Options -}}
			<option value="{{ .Value }}"{{ IS_LANG .Value }} selected{{ END }}>{{ .Text }}</option>
		{{ end -}}
		</select>
	</div>
</div>
{{ end -}}
` //`

var example_request_body = `{{ define "example_request_body" -}}
<div class="xs-request-body">
	{{ range .Snippets -}}
	<div class="code-snippet-container lang-{{ .Lang }}{{ IS_LANG .Lang }} selected{{ END }}" data-lang="{{ .Lang }}">
		<div class="cs-lines-container">
			{{ range $i, $_ := .Lines -}}
			<div>{{ $i }}</div>
			{{ end -}}
		</div>
		<div class="cs-code-container">
		{{ if (is_code_snippet_http .Snippet) -}}
			{{ template "code_snippet_http" .Snippet }}
		{{ else if (is_code_snippet_curl .Snippet) -}}
			{{ template "code_snippet_curl" .Snippet }}
		{{ end -}}
		</div>
	</div>
	{{ end -}}
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Code Snippets
////////////////////////////////////////////////////////////////////////////////

var code_snippet_http = `{{ define "code_snippet_http" -}}
<pre class="cs-pre lang-http">
<code class="lang-http"><span class="token http-method {{ .ClassMethod }}">{{ .Method }}</span> <span class="token http-uri">{{ .RequestURI }}</span> <span class="token http-version">{{ .HTTPVersion }}</span>
{{ range .Headers -}}
<span class="token http-header-key">{{ .Key }}:</span> <span class="token http-header-value">{{ .Value }}</span>
{{ end }}
{{- with .Body }}
<span class="token http-body">{{ . }}</span>
{{- end -}}
</code></pre>
{{ end -}}
` //`

var code_snippet_curl = `{{ define "code_snippet_curl" -}}
{{ $LB := (sh_line_break .NumOpts) -}}
<pre class="cs-pre lang-curl">
<code class="lang-curl"><span class="token curl-cmd">curl</span> <span class="token curl-flag">-X</span> <span class="token curl-flag-value {{ .ClassMethod }}">{{ .X }}</span> <span class="token curl-url">"{{ .URL }}"</span>{{ call $LB }}
{{- range .H }}
    <span class="token curl-flag">-H</span> <span class="token curl-header-value">'{{ . }}'</span>{{ call $LB }}
{{- end }}
{{- range .Data }}
    <span class="token curl-flag">-d</span> {{ template "curl_data" . }}{{ call $LB }}
{{- end }}</code>
</pre>
{{ end -}}
` //`

var curl_data = `{{ define "curl_data" -}}
{{- if (is_curl_data_text .) -}}
<span class="token curl-data-text">'{{ .HTML }}'</span>
{{- else if (is_curl_data_key_value .) -}}
<span class="token curl-data-key">{{ .Key }}</span><span class="token curl-data-op">=</span><span class="token curl-data-value">{{ .Value }}</span>
{{- end -}}
{{- end -}}
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
		<a class="field-anchor" href="{{ .Href }}"><span class="field-anchor-icon">#</span></a>
		{{ with .Path -}}
		<span class="field-path">{{ . }}</span>
		{{- end -}}
		<span class="field-name">{{ .Name }}</span>
		<span class="field-type">{{ .Type }}</span>
		{{- if .SettingText }}
		<span class="field-setting {{ .SettingLabel }}">{{ .SettingText }}</span>
		{{- end }}
		{{- if .ExpandableText }}
		<span class="field-expandability {{ .ExpandableLabel }}">{{ .ExpandableText }}</span>
		{{- end }}
		{{- with .SourceLink }}
		<a class="field-source-link" href="{{ .Href }}">{{ template "svg_code_icon_use" }}</a>
		{{- end }}
	</h3>
	<div class="field-text-container">
		{{- with .Text }}
		<div class="field-text text-box">
			{{ . }}
		</div>
		{{- end }}
	</div>
	{{- with .Validation }}
	<div class="field-validation text-box">
		{{ . }}
	</div>
	{{- end }}

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
<div class="field-list-container child collapsed">
	<h5 class="field-list-heading child">Child fields</h5>
	<ul class="field-list child">
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
<div class="enum-list-container collapsed">
	<h5 class="enum-list-heading">{{ .Title }}</h5>
	<ul class="enum-list">
		{{ range .Items -}}
		<li class="enum-item">
			<div class="enum-text text-box">
				<span class="enum-value">
					<code>{{ .Value }}</code>
				</span>
				<span class="enum-doc">
					{{- with .Text }}
					{{ . }}
					{{- end }}
				</span>
				{{- with .SourceLink }}
				<a href="{{ .Href }}" class="enum-source-link">{{ template "svg_code_icon_use" }}</a>
				{{- end }}
			</div>
		</li>
		{{ end -}}
	</ul>
</div>
{{ end -}}
` //`

////////////////////////////////////////////////////////////////////////////////
// Misc.
////////////////////////////////////////////////////////////////////////////////

var code_block_pre = `{{ define "code_block_pre" -}}
<pre class="code-block-pre">
<code class="lang-{{ .Lang }}">{{ .Code }}</code>
</pre>
{{ end -}}` //`

var svg_code_icon = `{{ define "svg_code_icon" -}}
<svg xmlns="http://www.w3.org/2000/svg" style="display:none">
	<symbol id="svg-code-icon" viewBox="0 0 16 16" stroke="currentcolor" fill="none">
		<path d="M15.466,4.241c-0,-2.065 -1.677,-3.741 -3.741,-3.741l-7.484,0c-2.065,0 -3.741,1.676 -3.741,3.741l-0,7.482c-0,2.065 1.676,3.741 3.741,3.741l7.484,0c2.064,0 3.741,-1.676 3.741,-3.741l-0,-7.482Z" style="stroke-width:1px;"/>
		<path d="M6.483,11l-2.985,-3l2.985,-3" style="stroke-width:1.5px;"/>
		<path d="M9.498,5l2.985,3l-2.985,3" style="stroke-width:1.5px;"/>
	</symbol>
</svg>
{{- end -}}` //`

var svg_code_icon_use = `{{ define "svg_code_icon_use" -}}
<svg width="16px" height="16px"><use href="#svg-code-icon"/></svg>
{{- end -}}` //`
