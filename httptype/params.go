package httptype

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

// Params wraps the given value v and returns an httptest.ParamSetter that can
// be used in the httptest.Request.Params field. The v argument must be a named
// struct or a pointer to a named struct, otherwise Params will panic.
//
// The SetParams(pattern string) (path string) implementation of the returned
// httptest.ParamSetter traverses the given struct and, using the rules outlined
// below, replaces any placeholders in the given pattern with the values of the
// struct's fields.
//
// The placeholders in the given pattern are expected to be demarcated with curly
// braces. For example:
//
//	"/users/{user_id}".
//
// By default the name of the field is used to match a placeholder, however this
// can be overridden by adding a `param` tag to the field. For example:
//
//	ID int `param:"user_id"`
//
// An embedded struct field will be traversed recursively.
//
// A non-embedded struct field will be ignored.
//
// Only fields that are exported and have the following types will be used to replace the placeholders:
//	- string
//	- bool
//	- int, int8, int16, int32, int64
//	- uint, uint8, uint16, uint32, uint64
//	- float32, float64
//	- <types whose underlying type is one of the above>
//	- <pointer types to any of the above>
//
func Params(v interface{}) httptest.ParamSetter {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		panic("httptest/httptype.Params: argument type invalid")
	} else if rv.Type().Name() == "" {
		panic("httptest/httptype.Params: argument type unnamed")
	}

	return paramSetter{v: v, rv: rv}
}

type paramSetter struct {
	v  interface{}
	rv reflect.Value
}

// Implements the httpdoc.Valuer interface.
func (p paramSetter) Value() (httpdoc.Value, error) { return p.v, nil }

// Implements the httptest.ParamSetter interface.
func (p paramSetter) SetParams(pattern string) (path string) {
	m := make(map[string]string)
	convertStructToMap(p.rv, m)

	var i, j int

	for {
		if i = strings.IndexByte(pattern, '{'); i > -1 {
			if j = strings.IndexByte(pattern, '}'); j > -1 && j > i {
				if v, ok := m[pattern[i+1:j]]; ok {
					path += pattern[:i] + v
				} else {
					path += pattern[:j+1]
				}
				pattern = pattern[j+1:]
				continue
			}
		}
		break
	}
	return path + pattern
}

func convertStructToMap(s reflect.Value, m map[string]string) {
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

		omitempty := false
		key := sf.Tag.Get("param")
		if i := strings.Index(key, ",omitempty"); i > -1 {
			omitempty = true
			key = strings.TrimSpace(key[:i])
		}
		if omitempty && f.IsZero() {
			continue
		}
		if len(key) < 1 {
			key = sf.Name
		}

		switch k := f.Kind(); k {
		case reflect.Bool:
			val := f.Bool()
			m[key] = strconv.FormatBool(val)
		case reflect.String:
			val := f.String()
			m[key] = val
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i64 := f.Int()
			m[key] = strconv.FormatInt(i64, 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u64 := f.Uint()
			m[key] = strconv.FormatUint(u64, 10)
		case reflect.Float32:
			f64 := f.Float()
			m[key] = strconv.FormatFloat(f64, 'f', -1, 32)
		case reflect.Float64:
			f64 := f.Float()
			m[key] = strconv.FormatFloat(f64, 'f', -1, 64)
		case reflect.Struct:
			// Converting embedded struct fields is supported.
			// Converting nested struct fields is not.
			if sf.Anonymous {
				convertStructToMap(f, m)
			}
		}
	}
}
