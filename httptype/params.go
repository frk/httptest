package httptype

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

func Params(v interface{}) httptest.ParamSetter {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("httptest/httptype.Params: invalid argument type")
	}
	return paramsSetter{v: v, rv: rv}
}

type paramsSetter struct {
	v  interface{}
	rv reflect.Value
}

func (ps paramsSetter) Value() (httpdoc.Value, error) { return ps.v, nil }

func (ps paramsSetter) SetParams(pattern string) (path string) {
	m := make(map[string]string)
	convertStructToMap(ps.rv, m)

	var i, j int

	for {
		if i = strings.IndexByte(pattern, '{'); i > -1 {
			if j = strings.IndexByte(pattern, '}'); j > -1 && j > i {
				if v, ok := m[pattern[i+1:j]]; ok {
					path += pattern[:i] + v
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
			f = f.Elem()
		}

		key := sf.Tag.Get("param")
		if len(key) < 1 {
			key = sf.Name
		}

		switch k := f.Kind(); k {
		case reflect.Bool:
			val := f.Bool()
			m[key] = strconv.FormatBool(val)
		case reflect.String:
			val := f.Convert(stringType).Interface().(string)
			m[key] = val
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i64 := f.Int()
			m[key] = strconv.FormatInt(i64, 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u64 := f.Uint()
			m[key] = strconv.FormatUint(u64, 10)
		case reflect.Float32, reflect.Float64:
			f64 := f.Float()
			m[key] = strconv.FormatFloat(f64, 'f', -1, 64)
		case reflect.Struct:
			// Converting embedded struct fields is supported,
			// converting normal struct fields is not.
			if sf.Anonymous {
				convertStructToMap(f, m)
			}
		}
	}
}
