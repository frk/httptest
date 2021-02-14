package types

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest/httpdoc/internal/testdata/types"
)

func TestTypeOf(t *testing.T) {
	//pos := Position{}
	pos := Position{Filename: "/some/file", Line: 1234}
	pkg := "github.com/frk/httptest/httpdoc/internal/testdata/types"
	tests := []struct {
		v    interface{}
		want *Type
		flag bool
	}{{
		v:    "foobar",
		want: &Type{Name: "string", Kind: KindString},
	}, {
		v:    new(float64),
		want: &Type{Kind: KindPtr, Elem: &Type{Name: "float64", Kind: KindFloat64}},
	}, {
		v:    types.T1{},
		want: &Type{Pos: pos, Name: "T1", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{}},
	}, {
		v: types.T2{},
		want: &Type{Pos: pos, Name: "T2", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
		}},
	}, {
		v: types.T3{},
		want: &Type{Pos: pos, Name: "T3", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F2", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F3", Pos: pos, Type: &Type{Name: "float64", Kind: KindFloat64}},
			{Name: "F4", Pos: pos, Type: &Type{Name: "float64", Kind: KindFloat64}},
		}},
	}, {
		v: types.T4{},
		want: &Type{Pos: pos, Name: "T4", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F2", Pos: pos, Type: &Type{Name: "int", Kind: KindInt}},
			{Name: "F3", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
		}},
	}, {
		v: types.T5a{},
		want: &Type{Name: "T5a", Pos: pos, Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}},
			},
			{Name: "F2", Pos: pos, Type: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}},
			},
		}},
	}, {
		v: types.T5b{},
		want: &Type{Pos: pos, Name: "T5b", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
			{Name: "F2", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
		}},
	}, {
		v: types.T6{},
		want: func() *Type {
			t := &Type{Pos: pos, Name: "T6", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindPtr}},
			}}
			t.Fields[0].Type.Elem = t
			t.Fields[1].Type.Elem = t
			return t
		}(),
	}, {
		v: types.T7{F: "foo bar"},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
		}},
	}, {
		v: types.T7{F: types.T7{}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
				Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
					{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface}},
				},
			}}},
		}},
	}, {
		v: &types.T7{F: &types.T7{}},
		want: &Type{Kind: KindPtr, Elem: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}}}},
		}}},
	}, {
		v: types.T7{F: &types.T7{"foo bar"}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
				Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
					{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
				},
			}}}},
		}},
	}, {
		v: types.T7{F: &types.T8{F1: "foo bar"}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
				Pos: pos, Name: "T8", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
					{Name: "F1", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
					{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
				},
			}}}},
		}},
	}, {
		v: types.T9{},
		want: func() *Type {
			t := &Type{Name: "T9", Pos: pos, Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t
			return t
		}(),
	}, {
		v: types.T9{F1: &types.T9{F2: "foo bar"}},
		want: func() *Type {
			t := &Type{Name: "T9", Pos: pos, Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t

			return &Type{Name: "T9", Pos: pos, Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "string", Kind: KindString,
						}}},
					}},
				}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
		}(),
	}, {
		v: types.T9{
			F1: &types.T9{F2: "foo bar"},
			F2: &types.T9{F2: 123},
		},
		want: func() *Type {
			t := &Type{Pos: pos, Name: "T9", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t

			return &Type{Pos: pos, Name: "T9", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "string", Kind: KindString,
						}}},
					}},
				}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "int", Kind: KindInt,
						}}},
					},
				}}}},
			}}
		}(),
	}, {
		v: types.S1(""),
		want: &Type{Name: "S1", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S1 decl doc"},
		},
	}, {
		v: types.S2(""),
		want: &Type{Name: "S2", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S2 comment"},
		},
	}, {
		v: types.S3(""),
		want: &Type{Name: "S3", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"},
		},
	}, {
		v: types.S4(""),
		want: &Type{Name: "S4", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S4 spec doc"},
		},
	}, {
		v: types.S5(""),
		want: &Type{Name: "S5", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"},
		},
	}, {
		v: types.S6(""),
		want: &Type{Name: "S6", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"/*\n\tS6 decl doc line 1\n\tS6 decl doc line 2\n*/"},
		},
	}, {
		v: types.S7(""),
		want: &Type{Name: "S7", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{
				"/*\n\tS7 decl doc line 1\n\tS7 decl doc line 2\n*/",
				"// S7 decl doc line 3",
				"// S7 decl doc line 4",
			},
		},
	}, {
		v: types.S8{},
		want: &Type{Name: "S8", Pos: pos, Kind: KindStruct, PkgPath: pkg,
			Doc: []string{"// S8 decl doc"},
			Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Name: "string", Kind: KindString},
					Doc: []string{
						"// S8.F1 doc line 1",
						"// S8.F1 doc line 2",
					},
				},
				{Name: "F2", Pos: pos, Type: &Type{Name: "string", Kind: KindString},
					Doc: []string{
						"// S8.F2 comment",
					},
				},
				{Name: "F3", Pos: pos, Type: &Type{Name: "string", Kind: KindString},
					Doc: []string{
						"/* S8.F3 comment line 1\n\tS8.F3 comment line 2\n\tS8.F3 comment line 3\n\t*/",
					},
				},
			},
		},
	}}

	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filepath.Dir(f))
	src, err := Load(dir + "/testdata/types")
	if err != nil {
		t.Fatal(err)
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	for i, tt := range tests {
		if tt.flag {
			continue
		}
		got := src.TypeOf(tt.v)
		if err := cmp.Compare(got, tt.want); err != nil {
			t.Errorf("[%d] %T: %v", i, tt.v, err)
		}
	}
}
