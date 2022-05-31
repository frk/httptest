package httptest

import (
	"reflect"
	"testing"
)

func Test_E_Split(t *testing.T) {
	tests := []struct {
		E    E
		want []string
	}{{
		E:    "POST /foo/{bar}",
		want: []string{"POST", "/foo/{bar}"},
	}, {
		E:    "DELETE foos",
		want: []string{"DELETE", "/foos"},
	}, {
		E:    "GET /",
		want: []string{"GET", "/"},
	}, {
		E:    "/foo/{bar}",
		want: []string{"GET", "/foo/{bar}"},
	}, {
		E:    "/",
		want: []string{"GET", "/"},
	}, {
		E:    "",
		want: []string{"GET", "/"},
	}}

	for _, tt := range tests {
		method, pattern := tt.E.Split()
		got := []string{method, pattern}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q: got %q, want %q", string(tt.E), got, tt.want)
		}
	}
}

func Test_Params_SetParams(t *testing.T) {
	tests := []struct {
		patt   string
		params Params
		want   string
	}{
		{"", Params{}, ""},
		{"/foo/bar", Params{}, "/foo/bar"},
		{"/foo/{id}", Params{"id": 123}, "/foo/123"},
		{"/users/{user_id}/posts/{post_id}", Params{"user_id": 123, "post_id": 345}, "/users/123/posts/345"},
		{"/users/{user_id}/posts/{post_id}", Params{"user_id": 123}, "/users/123"},
		{"{a}{b}{c}", Params{"a": "foo", "b": "bar", "c": "baz"}, "foobarbaz"},

		{"{a}{b}c}", Params{"a": "foo", "b": "bar", "c": "baz"}, "foobarc}"},
		{"{a}{b}{c", Params{"a": "foo", "b": "bar", "c": "baz"}, "foobar{c"},
	}

	for _, tt := range tests {
		if got := tt.params.SetParams(tt.patt); got != tt.want {
			t.Errorf("%q: got %q, want %q", tt.patt, got, tt.want)
		}
	}
}
