package httptype

import (
	"reflect"
	"strings"

	"github.com/frk/form"
	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

func Query(v interface{}) httptest.QueryGetter {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("httptest/httptype.Query: invalid argument type")
	}
	return queryGetter{v: v, rv: rv}
}

type queryGetter struct {
	v  interface{}
	rv reflect.Value
}

func (qg queryGetter) Value() (httpdoc.Value, error) { return qg.v, nil }

func (qg queryGetter) GetQuery() string {
	var b strings.Builder
	if err := form.NewEncoder(&b).WithTagKey("query").Encode(qg.v); err != nil {
		panic("httptest/httptype.GetQuery: " + err.Error())
	}
	return b.String()
}
