package page

import (
	"html/template"
	"strings"
)

var helpers = template.FuncMap{
	"lower": strings.ToLower,
	"sh_line_break": func(numlines int) (f func() string) {
		return func() string {
			if numlines > 0 {
				numlines--
				return ` \`
			}
			return ""
		}
	},

	////////////////////////////////////////////////////////////////////////
	// sidebar banner type assertion
	////////////////////////////////////////////////////////////////////////
	"is_sidebar_banner_title": func(s SidebarBanner) *SidebarBannerTitle {
		if v, ok := s.(*SidebarBannerTitle); ok && v != nil {
			return v
		}
		return nil
	},
	"is_sidebar_banner_html": func(s SidebarBanner) *SidebarBannerHTML {
		if v, ok := s.(*SidebarBannerHTML); ok && v != nil {
			return v
		}
		return nil
	},

	////////////////////////////////////////////////////////////////////////
	// article section type assertion
	////////////////////////////////////////////////////////////////////////
	"is_article_text": func(s ArticleSection) *ArticleText {
		if v, ok := s.(*ArticleText); ok && v != nil {
			return v
		}
		return nil
	},
	"is_article_auth_info": func(s ArticleSection) *ArticleAuthInfo {
		if v, ok := s.(*ArticleAuthInfo); ok && v != nil {
			return v
		}
		return nil
	},
	"is_article_field_list": func(s ArticleSection) *ArticleFieldList {
		if v, ok := s.(*ArticleFieldList); ok && v != nil {
			return v
		}
		return nil
	},

	////////////////////////////////////////////////////////////////////////
	// example section type assertion
	////////////////////////////////////////////////////////////////////////
	"is_example_text": func(s ExampleSection) *ExampleText {
		if v, ok := s.(*ExampleText); ok && v != nil {
			return v
		}
		return nil
	},
	"is_example_endpoints": func(s ExampleSection) *ExampleEndpoints {
		if v, ok := s.(*ExampleEndpoints); ok && v != nil {
			return v
		}
		return nil
	},
	"is_example_object": func(s ExampleSection) *ExampleObject {
		if v, ok := s.(*ExampleObject); ok && v != nil {
			return v
		}
		return nil
	},
	"is_example_request": func(s ExampleSection) *ExampleRequest {
		if v, ok := s.(*ExampleRequest); ok && v != nil {
			return v
		}
		return nil
	},
	"is_example_response": func(s ExampleSection) *ExampleResponse {
		if v, ok := s.(*ExampleResponse); ok && v != nil {
			return v
		}
		return nil
	},

	////////////////////////////////////////////////////////////////////////
	// code snippet type assertion
	////////////////////////////////////////////////////////////////////////
	"is_code_snippet_http": func(s CodeSnippet) *CodeSnippetHTTP {
		if v, ok := s.(*CodeSnippetHTTP); ok && v != nil {
			return v
		}
		return nil
	},
	"is_code_snippet_curl": func(s CodeSnippet) *CodeSnippetCURL {
		if v, ok := s.(*CodeSnippetCURL); ok && v != nil {
			return v
		}
		return nil
	},

	////////////////////////////////////////////////////////////////////////
	// curl data type assertion
	////////////////////////////////////////////////////////////////////////
	"is_curl_data_text": func(t CURLDataType) CURLDataText {
		if v, ok := t.(CURLDataText); ok && len(v) > 0 {
			return v
		}
		return ""
	},
	"is_curl_data_key_value": func(t CURLDataType) *CURLDataKeyValue {
		if v, ok := t.(CURLDataKeyValue); ok {
			return &v
		}
		return nil
	},
}
