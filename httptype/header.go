package httptype

import (
	"net/http"
	"reflect"

	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

func Header(v interface{}) httptest.HeaderGetter {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("httptest/httptype.Header: invalid argument type")
	}
	return headerGetter{v: v, rv: rv}
}

type headerGetter struct {
	v  interface{}
	rv reflect.Value
}

func (hg headerGetter) Value() (httpdoc.Value, error) { return hg.v, nil }

func (hg headerGetter) GetHeader() http.Header {
	h := make(http.Header)
	convertStructToHeader(hg.rv, h)
	return h
}

var stringType = reflect.TypeOf("")

func convertStructToHeader(s reflect.Value, h http.Header) {
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f, sf := s.Field(i), t.Field(i)
		if f.Kind() == reflect.Ptr {
			f = f.Elem()
		}

		key := sf.Tag.Get("header")
		if len(key) < 1 {
			key = sf.Name
		}

		switch f.Kind() {
		case reflect.String:
			val := f.Convert(stringType).Interface().(string)
			h[key] = append(h[key], val)
		case reflect.Slice:
			if f.Elem().Kind() == reflect.String {
				val := f.Interface().([]string)
				h[key] = append(h[key], val...)
			}
		case reflect.Struct:
			// Converting embedded struct fields is supported,
			// converting normal struct fields is not.
			if sf.Anonymous {
				convertStructToHeader(f, h)
			}
		}
	}
}
