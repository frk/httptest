package httptype

import (
	"testing"

	"github.com/frk/compare"
)

func Test_halfInit(t *testing.T) {
	type T1 struct {
		F1 string
		F2 bool
	}

	type T2 struct {
		F1 string
		F2 any
	}

	type T3[T any] struct {
		F1 int
		F2 T
	}

	tests := []struct {
		in   any
		want any
	}{
		{"foo", ""},
		{T1{"foo", true}, T1{}},
		{&T1{"foo", true}, (*T1)(nil)},
		{T2{"foo", nil}, T2{}},
		{&T2{"foo", nil}, (*T2)(nil)},
		{&T2{"foo", true}, &T2{"", false}},
		{&T2{"foo", &T1{}}, &T2{"", &T1{}}},
		{
			&T2{"foo", &T2{"", map[string]any{"foo": 9.9, "bar": new(string)}}},
			&T2{"", &T2{"", map[string]any{"foo": float64(0), "bar": new(string)}}},
		},
		{
			&T3[[]*T2]{123, []*T2{{}, {"", []byte(`heeloo`)}}},
			&T3[[]*T2]{0, []*T2{nil, {"", []byte(nil)}}},
		},
	}

	for _, tt := range tests {
		got := halfInit(tt.in)
		if e := compare.Compare(got, tt.want); e != nil {
			t.Error(e)
		}
	}
}
