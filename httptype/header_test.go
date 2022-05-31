package httptype

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/frk/compare"
)

func TestHeader_panic(t *testing.T) {
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

			_ = Header(tt.v)
		})
	}
}

func TestHeader_GetHeader(t *testing.T) {
	type customString string
	type embeddedStruct struct {
		E1 string
		E2 []customString
		E3 [2]string

		E4 []string
		E5 string   `header:"X-4"`
		E6 []string `header:"X-4"`

		// pointers
		E1p *string
		E2p *[]customString
		E3p *[2]string

		E4p *[]string
		E5p *string   `header:"X-4"`
		E6p *[]string `header:"X-4"`
	}
	type headerStruct struct {
		h0 string
		H1 string
		H2 customString
		H3 []string
		H4 []customString
		H5 [2]string
		H6 [3]customString
		H7 []customString `header:"X-4"`

		embeddedStruct
		E4 string

		// pointers
		H1p *string
		H2p *customString
		H3p *[]string
		H4p *[]customString
		H5p *[2]string
		H6p *[3]customString
		H7p *[]customString `header:"X-4"`
		E4p *string
	}

	sptr := func(s string) *string { return &s }
	csptr := func(s customString) *customString { return &s }

	tests := []struct {
		v    interface{}
		want http.Header
	}{{
		// ignore empty fields
		v:    headerStruct{},
		want: http.Header{},
	}, {
		// ignore unexported fields
		v:    headerStruct{h0: "value0"},
		want: http.Header{},
	}, {
		v:    headerStruct{H1: "value1"},
		want: http.Header{"H1": {"value1"}},
	}, {
		// accept custom string types
		v:    headerStruct{H2: "value2"},
		want: http.Header{"H2": {"value2"}},
	}, {
		// accept string slices
		v: headerStruct{
			H3: []string{"value3a", "value3b"},
			H4: []customString{"value4a", "value4b"},
		},
		want: http.Header{
			"H3": {"value3a", "value3b"},
			"H4": {"value4a", "value4b"},
		},
	}, {
		// accept string arrays
		v: headerStruct{
			H5: [2]string{"value5a", "value5b"},
			H6: [3]customString{"value6a", "value6b", "value6c"},
		},
		want: http.Header{
			"H5": {"value5a", "value5b"},
			"H6": {"value6a", "value6b", "value6c"},
		},
	}, {
		// handle embedded structs recursively
		v: headerStruct{embeddedStruct: embeddedStruct{
			E1: "embe1",
			E2: []customString{"embe2a", "embe2b", "embe2c"},
			E3: [2]string{"embe3a", "embe3b"},
		}},
		want: http.Header{
			"E1": {"embe1"},
			"E2": {"embe2a", "embe2b", "embe2c"},
			"E3": {"embe3a", "embe3b"},
		},
	}, {
		// can aggregate values from multiple fields under one header key
		v: headerStruct{
			H7: []customString{"embeX4a", "embeX4b"},
			embeddedStruct: embeddedStruct{
				E4: []string{"embe4a", "embe4b"},
				E5: "embeX4c",
				E6: []string{"embeX4d", "embeX4e"},
			},
			E4: "embe4c",
		},
		want: http.Header{
			"E4":  {"embe4a", "embe4b", "embe4c"},
			"X-4": {"embeX4a", "embeX4b", "embeX4c", "embeX4d", "embeX4e"},
		},
	}, {
		////////////////////////////////////////////////////////////////////////
		// the rest is a more or less a copy of the above but with the pointer fields
		////////////////////////////////////////////////////////////////////////

		v:    headerStruct{H1p: sptr("value1")},
		want: http.Header{"H1p": {"value1"}},
	}, {
		// accept custom string types
		v:    headerStruct{H2p: csptr("value2")},
		want: http.Header{"H2p": {"value2"}},
	}, {
		// accept string slices
		v: headerStruct{
			H3p: &[]string{"value3a", "value3b"},
			H4p: &[]customString{"value4a", "value4b"},
		},
		want: http.Header{
			"H3p": {"value3a", "value3b"},
			"H4p": {"value4a", "value4b"},
		},
	}, {
		// accept string arrays
		v: headerStruct{
			H5p: &[2]string{"value5a", "value5b"},
			H6p: &[3]customString{"value6a", "value6b", "value6c"},
		},
		want: http.Header{
			"H5p": {"value5a", "value5b"},
			"H6p": {"value6a", "value6b", "value6c"},
		},
	}, {
		// handle embedded structs recursively
		v: headerStruct{embeddedStruct: embeddedStruct{
			E1p: sptr("embe1"),
			E2p: &[]customString{"embe2a", "embe2b", "embe2c"},
			E3p: &[2]string{"embe3a", "embe3b"},
		}},
		want: http.Header{
			"E1p": {"embe1"},
			"E2p": {"embe2a", "embe2b", "embe2c"},
			"E3p": {"embe3a", "embe3b"},
		},
	}, {
		// can aggregate values from multiple fields under one header key
		v: headerStruct{
			embeddedStruct: embeddedStruct{
				E4p: &[]string{"embe4a", "embe4b"},
				E5p: sptr("embeX4a"),
				E6p: &[]string{"embeX4b", "embeX4c"},
			},
			E4p: sptr("embe4c"),
			H7p: &[]customString{"embeX4d", "embeX4e"},
		},
		want: http.Header{
			"E4p": {"embe4a", "embe4b", "embe4c"},
			"X-4": {"embeX4a", "embeX4b", "embeX4c", "embeX4d", "embeX4e"},
		},
	}}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.v)
		t.Run(name, func(t *testing.T) {
			got := Header(tt.v).GetHeader()
			if e := compare.Compare(got, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}
