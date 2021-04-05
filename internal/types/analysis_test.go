package types

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest/internal/testdata/types"
)

func TestTypeOf(t *testing.T) {
	//pos := Position{}
	pos := Position{Filename: "/some/file", Line: 1234}
	pkg := "github.com/frk/httptest/internal/testdata/types"
	tests := []struct {
		v    interface{}
		want *Type
		flag bool
	}{0: {
		v:    "foobar",
		want: &Type{Name: "string", Kind: KindString},
	}, 1: {
		v:    new(float64),
		want: &Type{Kind: KindPtr, Elem: &Type{Name: "float64", Kind: KindFloat64}},
	}, 2: {
		v:    types.T1{},
		want: &Type{Pos: pos, Name: "T1", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T1{}), PkgPath: pkg, Fields: []*StructField{}},
	}, 3: {
		v: types.T2{},
		want: &Type{Pos: pos, Name: "T2", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
		}},
	}, 4: {
		v: types.T3{},
		want: &Type{Pos: pos, Name: "T3", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T3{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F2", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F3", Pos: pos, Type: &Type{Name: "float64", Kind: KindFloat64}},
			{Name: "F4", Pos: pos, Type: &Type{Name: "float64", Kind: KindFloat64}},
		}},
	}, 5: {
		v: types.T4{},
		want: &Type{Pos: pos, Name: "T4", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T4{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Name: "string", Kind: KindString}},
			{Name: "F2", Pos: pos, Type: &Type{Name: "int", Kind: KindInt}},
			{Name: "F3", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
		}},
	}, 6: {
		v: types.T5a{},
		want: &Type{Name: "T5a", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T5a{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}},
			},
			{Name: "F2", Pos: pos, Type: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}},
			},
		}},
	}, 7: {
		v: types.T5b{},
		want: &Type{Pos: pos, Name: "T5b", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T5b{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
			{Name: "F2", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
				Name: "T2", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T2{}), PkgPath: pkg,
				Fields: []*StructField{{Name: "F", Pos: pos, Type: &Type{Name: "string", Kind: KindString}}}}},
			},
		}},
	}, 8: {
		v: types.T6{},
		want: func() *Type {
			t := &Type{Pos: pos, Name: "T6", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T6{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindPtr}},
			}}
			t.Fields[0].Type.Elem = t
			t.Fields[1].Type.Elem = t
			return t
		}(),
	}, 9: {
		v: types.T7{F: "foo bar"},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
		}},
	}, 10: {
		v: types.T7{F: types.T7{}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
				Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
					{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface}},
				},
			}}},
		}},
	}, 11: {
		v: &types.T7{F: &types.T7{}},
		want: &Type{Kind: KindPtr, Elem: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}}}},
		}}},
	}, 12: {
		v: types.T7{F: &types.T7{"foo bar"}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
				Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
					{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
				},
			}}}},
		}},
	}, 13: {
		v: types.T7{F: &types.T8{F1: "foo bar"}},
		want: &Type{Pos: pos, Name: "T7", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T7{}), PkgPath: pkg, Fields: []*StructField{
			{Name: "F", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
				Pos: pos, Name: "T8", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T8{}), PkgPath: pkg, Fields: []*StructField{
					{Name: "F1", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Name: "string", Kind: KindString}}},
					{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
				},
			}}}},
		}},
	}, 14: {
		v: types.T9{},
		want: func() *Type {
			t := &Type{Name: "T9", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t
			return t
		}(),
	}, 15: {
		v: types.T9{F1: &types.T9{F2: "foo bar"}},
		want: func() *Type {
			t := &Type{Name: "T9", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t

			return &Type{Name: "T9", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "string", Kind: KindString,
						}}},
					}},
				}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
		}(),
	}, 16: {
		v: types.T9{
			F1: &types.T9{F2: "foo bar"},
			F2: &types.T9{F2: 123},
		},
		want: func() *Type {
			t := &Type{Pos: pos, Name: "T9", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface}},
			}}
			t.Fields[0].Type.Elem = t

			return &Type{Pos: pos, Name: "T9", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
				{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "string", Kind: KindString,
						}}},
					}},
				}},
				{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{Kind: KindPtr, Elem: &Type{
					Pos: pos, Name: "T9", Kind: KindStruct, ReflectType: reflect.TypeOf(types.T9{}), PkgPath: pkg, Fields: []*StructField{
						{Name: "F1", Pos: pos, Type: &Type{Kind: KindPtr, Elem: t}},
						{Name: "F2", Pos: pos, Type: &Type{Kind: KindInterface, Elem: &Type{
							Name: "int", Kind: KindInt,
						}}},
					},
				}}}},
			}}
		}(),
	}, 17: {
		v: types.S1(""),
		want: &Type{Name: "S1", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S1 decl doc"},
		},
	}, 18: {
		v: types.S2(""),
		want: &Type{Name: "S2", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S2 comment"},
		},
	}, 19: {
		v: types.S3(""),
		want: &Type{Name: "S3", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"},
		},
	}, 20: {
		v: types.S4(""),
		want: &Type{Name: "S4", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S4 spec doc"},
		},
	}, 21: {
		v: types.S5(""),
		want: &Type{Name: "S5", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"},
		},
	}, 22: {
		v: types.S6(""),
		want: &Type{Name: "S6", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{"/*\n\tS6 decl doc line 1\n\tS6 decl doc line 2\n*/"},
		},
	}, 23: {
		v: types.S7(""),
		want: &Type{Name: "S7", Pos: pos, Kind: KindString, PkgPath: pkg,
			Doc: []string{
				"/*\n\tS7 decl doc line 1\n\tS7 decl doc line 2\n*/",
				"// S7 decl doc line 3",
				"// S7 decl doc line 4",
			},
		},
	}, 24: {
		v: types.S8{},
		want: &Type{Name: "S8", Pos: pos, Kind: KindStruct, ReflectType: reflect.TypeOf(types.S8{}), PkgPath: pkg,
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
	}, 25: {
		v: types.C1(0),
		want: &Type{Name: "C1", Pos: pos, Kind: KindInt, PkgPath: pkg,
			Values: []*ConstValue{
				{Name: "c1a", Pos: pos, Value: `-1`, Doc: []string{"// c1a spec doc line 1", "// c1a spec doc line 2"}},
				{Name: "c1b", Pos: pos, Value: `1`, Doc: []string{"// C1 const decl doc"}},
				{Name: "c1c", Pos: pos, Value: `13`, Doc: []string{"// c1c comment line"}},
			},
		},
	}, 26: {
		v: types.C2(""),
		want: &Type{Name: "C2", Pos: pos, Kind: KindString, PkgPath: pkg,
			Values: []*ConstValue{
				{Name: "c2a", Pos: pos, Value: `"foo"`},
				{Name: "c2b", Pos: pos, Value: `"bar"`},
				{Name: "c2d", Pos: pos, Value: `"baz"`, Doc: []string{"// c2d comment line"}},
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

func TestTypeDeclOf(t *testing.T) {
	pos := Position{Filename: "/some/file", Line: 1234}
	pkg := "github.com/frk/httptest/internal/testdata/types"
	tests := []struct {
		v    interface{}
		want *TypeDecl
	}{0: {
		v:    types.T1{},
		want: &TypeDecl{Pos: pos, Name: "T1", PkgPath: pkg},
	}, 1: {
		v:    &types.T9{},
		want: &TypeDecl{Pos: pos, Name: "T9", PkgPath: pkg},
	}, 2: {
		v:    types.S1(""),
		want: &TypeDecl{Name: "S1", PkgPath: pkg, Pos: pos, Doc: []string{"// S1 decl doc"}},
	}, 3: {
		v: types.S2(""),
		want: &TypeDecl{Name: "S2", Pos: pos, PkgPath: pkg,
			Doc: []string{"// S2 comment"}},
	}, 4: {
		v: types.S3(""),
		want: &TypeDecl{Name: "S3", Pos: pos, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"}},
	}, 5: {
		v: types.S4(""),
		want: &TypeDecl{Name: "S4", Pos: pos, PkgPath: pkg,
			Doc: []string{"// S4 spec doc"}},
	}, 6: {
		v: types.S5(""),
		want: &TypeDecl{Name: "S5", Pos: pos, PkgPath: pkg,
			Doc: []string{"// S3,S4,S5 decl doc"}},
	}, 7: {
		v: types.S6(""),
		want: &TypeDecl{Name: "S6", Pos: pos, PkgPath: pkg,
			Doc: []string{"/*\n\tS6 decl doc line 1\n\tS6 decl doc line 2\n*/"}},
	}, 8: {
		v: types.S7(""),
		want: &TypeDecl{Name: "S7", Pos: pos, PkgPath: pkg,
			Doc: []string{
				"/*\n\tS7 decl doc line 1\n\tS7 decl doc line 2\n*/",
				"// S7 decl doc line 3",
				"// S7 decl doc line 4",
			}},
	}}

	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filepath.Dir(f))
	src, err := Load(dir + "/testdata/types")
	if err != nil {
		t.Fatal(err)
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	for i, tt := range tests {
		got := src.TypeDeclOf(tt.v)
		if err := cmp.Compare(got, tt.want); err != nil {
			t.Errorf("[%d] %T: %v", i, tt.v, err)
		}
	}
}
