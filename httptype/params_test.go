package httptype

import (
	"fmt"
	"testing"

	"github.com/frk/httptest"
)

func TestParams_panic(t *testing.T) {
	type NamedStruct struct{}

	tests := []struct {
		v         interface{}
		wantPanic bool
	}{
		// bad
		{v: nil, wantPanic: true},
		{v: "foo bar", wantPanic: true},
		{v: []string{}, wantPanic: true},
		{v: map[string][]string{}, wantPanic: true},
		{v: struct{}{}, wantPanic: true},
		{v: &struct{}{}, wantPanic: true},

		// ok
		{v: NamedStruct{}, wantPanic: false},
		{v: &NamedStruct{}, wantPanic: false},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.v)
		t.Run(name, func(t *testing.T) {
			defer func() {
				x := recover()
				if tt.wantPanic != (x != nil) {
					t.Errorf("want_panic=%t; got='%v';", tt.wantPanic, x)
				}
			}()

			_ = Params(tt.v)
		})
	}
}

func TestParams_SetParams(t *testing.T) {
	type paramStruct struct {
		p0 string
		P1 string
		P2 string `param:",omitempty"`
		P3 string `param:"tag3"`

		B bool

		I0  int
		I8  int8
		I16 int16
		I32 int32
		I64 int64

		U0  uint
		U8  uint8
		U16 uint16
		U32 uint32
		U64 uint64

		F32 float32
		F64 float64

		*paramStruct // embedded

		// TODO add tests for pointer fields
	}

	tests := []struct {
		params        httptest.ParamSetter
		pattern, want string
	}{{
		params:  Params(paramStruct{}),
		pattern: "/foo/bar", want: "/foo/bar",
	}, {
		params:  Params(paramStruct{}),
		pattern: "/foo/{x}", want: "/foo/{x}",
	}, {
		params:  Params(paramStruct{}),
		pattern: "/foo/{x}/", want: "/foo/{x}/",
	}, {
		params:  Params(paramStruct{}),
		pattern: "/foo/{x}/bar/{y}/baz", want: "/foo/{x}/bar/{y}/baz",
	}, {
		// ignore unexported fields
		params:  Params(paramStruct{p0: "v0"}),
		pattern: "/foo/{p0}/bar", want: "/foo/{p0}/bar",
	}, {
		params:  Params(paramStruct{P1: "v1"}),
		pattern: "/foo/{P1}/bar", want: "/foo/v1/bar",
	}, {
		// omitempty
		params:  Params(paramStruct{P2: ""}),
		pattern: "/foo/{P2}/bar", want: "/foo/{P2}/bar",
	}, {
		// use tag to match placeholder
		params:  Params(paramStruct{P3: "v3"}),
		pattern: "/foo/{P3}/bar/{tag3}/baz", want: "/foo/{P3}/bar/v3/baz",
	}, {
		// bool
		params:  Params(paramStruct{B: true}),
		pattern: "/foo/{B}/bar", want: "/foo/true/bar",
	}, {
		// ints
		params:  Params(paramStruct{I0: 1234567890, I8: -8, I16: 16, I32: 32, I64: -64}),
		pattern: "/foo/{I0}/{I8}/bar/{I16}/{I32}/baz/{I64}",
		want:    "/foo/1234567890/-8/bar/16/32/baz/-64",
	}, {
		// uints
		params:  Params(paramStruct{U0: 1234567890, U8: 8, U16: 16, U32: 32, U64: 64}),
		pattern: "/foo/{U0}/{U8}/bar/{U16}/{U32}/baz/{U64}",
		want:    "/foo/1234567890/8/bar/16/32/baz/64",
	}, {
		// floats
		params:  Params(paramStruct{F32: -0.354, F64: 0.0456789}),
		pattern: "/foo/{F32}/bar/{F64}/",
		want:    "/foo/-0.354/bar/0.0456789/",
	}, {

		//////////////
		// embedded field
		//////////////

		// ignore unexported fields
		params:  Params(paramStruct{paramStruct: &paramStruct{p0: "v0"}}),
		pattern: "/foo/{p0}/bar", want: "/foo/{p0}/bar",
	}, {
		params:  Params(paramStruct{paramStruct: &paramStruct{P1: "v1"}}),
		pattern: "/foo/{P1}/bar", want: "/foo/v1/bar",
	}, {
		// omitempty
		params:  Params(paramStruct{paramStruct: &paramStruct{P2: ""}}),
		pattern: "/foo/{P2}/bar", want: "/foo/{P2}/bar",
	}, {
		// use tag to match placeholder
		params:  Params(paramStruct{paramStruct: &paramStruct{P3: "v3"}}),
		pattern: "/foo/{P3}/bar/{tag3}/baz", want: "/foo/{P3}/bar/v3/baz",
	}, {
		// bool
		params:  Params(paramStruct{paramStruct: &paramStruct{B: true}}),
		pattern: "/foo/{B}/bar", want: "/foo/true/bar",
	}, {
		// ints
		params:  Params(paramStruct{paramStruct: &paramStruct{I0: 1234567890, I8: -8, I16: 16, I32: 32, I64: -64}}),
		pattern: "/foo/{I0}/{I8}/bar/{I16}/{I32}/baz/{I64}",
		want:    "/foo/1234567890/-8/bar/16/32/baz/-64",
	}, {
		// uints
		params:  Params(paramStruct{paramStruct: &paramStruct{U0: 1234567890, U8: 8, U16: 16, U32: 32, U64: 64}}),
		pattern: "/foo/{U0}/{U8}/bar/{U16}/{U32}/baz/{U64}",
		want:    "/foo/1234567890/8/bar/16/32/baz/64",
	}, {
		// floats
		params:  Params(paramStruct{paramStruct: &paramStruct{F32: -0.354, F64: 0.0456789}}),
		pattern: "/foo/{F32}/bar/{F64}/",
		want:    "/foo/-0.354/bar/0.0456789/",
	}}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := tt.params.SetParams(tt.pattern)
			if got != tt.want {
				t.Errorf("want=%q; got=%q", tt.want, got)
			}
		})
	}
}
