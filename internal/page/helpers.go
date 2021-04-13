package page

import (
	"html/template"
	"strings"
)

var helpers = template.FuncMap{
	"lower": strings.ToLower,

	////////////////////////////////////////////////////////////////////////
	// article section type assertion
	////////////////////////////////////////////////////////////////////////
	"is_text_article_section": func(s ArticleSection) *TextArticleSection {
		if v, ok := s.(*TextArticleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_auth_info_article_section": func(s ArticleSection) *AuthInfoArticleSection {
		if v, ok := s.(*AuthInfoArticleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_field_list_article_section": func(s ArticleSection) *FieldListArticleSection {
		if v, ok := s.(*FieldListArticleSection); ok && v != nil {
			return v
		}
		return nil
	},

	////////////////////////////////////////////////////////////////////////
	// example section type assertion
	////////////////////////////////////////////////////////////////////////
	"is_text_example_section": func(s ExampleSection) *TextExampleSection {
		if v, ok := s.(*TextExampleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_endpoints_example_section": func(s ExampleSection) *EndpointsExampleSection {
		if v, ok := s.(*EndpointsExampleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_object_example_section": func(s ExampleSection) *ObjectExampleSection {
		if v, ok := s.(*ObjectExampleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_request_example_section": func(s ExampleSection) *RequestExampleSection {
		if v, ok := s.(*RequestExampleSection); ok && v != nil {
			return v
		}
		return nil
	},
	"is_response_example_section": func(s ExampleSection) *ResponseExampleSection {
		if v, ok := s.(*ResponseExampleSection); ok && v != nil {
			return v
		}
		return nil
	},
}
