package httptype

import (
	"net/http"
	"reflect"

	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

// Header wraps the given value v and returns an httptest.HeaderGetter that
// can be used in the httptest.Request.Header and httptest.Response.Header
// fields. The v argument must be a named struct or a pointer to a named
// struct, otherwise Header will panic.
//
// The returned httptest.HeaderGetter implementation traverses the given struct
// and adds its fields into an http.Header value using the following rules:
//
// By default the name of the field is used as the http.Header key, however this
// can be overridden by adding a `header` tag to the field. For example:
//
//	Field string `header:"Header-Name"`
//
// An embedded struct field will be traversed recursively.
//
// A non-embedded struct field will be ignored.
//
// Only fields that are exported and have the following types are added to the http.Header value:
//	- string
//	- []string
//	- [N]string
//	- <a type whose underlying type is one of the above>
//	- <a pointer type to any of the above>
//
func Header(v interface{}) httptest.HeaderGetter {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		panic("httptest/httptype.Header: argument type invalid")
	} else if rv.Type().Name() == "" {
		panic("httptest/httptype.Header: argument type unnamed")
	}

	return headerGetter{v: v, rv: rv}
}

type headerGetter struct {
	v  interface{}
	rv reflect.Value
}

// Implements the httpdoc.Valuer interface.
func (h headerGetter) Value() (httpdoc.Value, error) { return h.v, nil }

// Implements the httptest.HeaderGetter interface.
func (h headerGetter) GetHeader() http.Header {
	m := make(http.Header)
	convertStructToHeader(h.rv, m)
	return m
}

func convertStructToHeader(s reflect.Value, h http.Header) {
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f, sf := s.Field(i), t.Field(i)
		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				continue
			}
			f = f.Elem()
		}
		if !sf.IsExported() && !sf.Anonymous && f.Kind() != reflect.Struct {
			continue
		}

		key := sf.Tag.Get("header")
		if len(key) < 1 {
			key = sf.Name
		}

		switch f.Kind() {
		case reflect.String:
			if s := f.String(); s != "" {
				h[key] = append(h[key], s)
			}
		case reflect.Slice, reflect.Array:
			if f.Type().Elem().Kind() == reflect.String {
				for i := 0; i < f.Len(); i++ {
					if s := f.Index(i).String(); s != "" {
						h[key] = append(h[key], s)
					}
				}
			}
		case reflect.Struct:
			// Converting embedded struct fields is supported.
			// Converting nested struct fields is not.
			if sf.Anonymous {
				convertStructToHeader(f, h)
			}
		}
	}
}
