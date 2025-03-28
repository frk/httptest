package types

import (
	"path"
	"reflect"
	"strconv"
)

// TypeDecl holds basic info on a declared Go type.
type TypeDecl struct {
	// The position of a type's declaration in the source code.
	Pos Position
	// The raw documentation of the type.
	Doc []string
	// The declared type's name
	Name string
	// The type's package import path
	PkgPath string
}

// Type is the representation of a Go type.
type Type struct {
	// The position of the type's declaration in the source code.
	Pos Position
	// The raw documentation of a named type.
	Doc []string
	// The name of a named type or empty string for unnamed types
	Name string
	// The kind of the go type.
	Kind Kind
	// The package import path.
	PkgPath string
	// Indicates whether or not the type is the "byte" alias type.
	IsByte bool
	// Indicates whether or not the type is the "rune" alias type.
	IsRune bool
	// If the base type's an array type, this field will hold the array's length.
	ArrayLen int
	// If kind is map, key will hold the info on the map's key type.
	Key *Type
	// If kind is map, elem will hold the info on the map's value type.
	// If kind is ptr, elem will hold the info on pointed-to type.
	// If kind is slice/array, elem will hold the info on slice/array element type.
	// If kind is interface and the analyzed value was valid, elem will hold
	// the info on the dynamic type.
	Elem *Type
	// If kind is struct, Fields will hold the list of the struct's fields.
	Fields []*StructField
	// If Type is a basic, defined type then Values will hold the list of
	// declared constants of that type, if any.
	Values []*ConstValue
	// NOTE: Set only for types of the struct kind.
	ReflectType reflect.Type

	hasiface bool `cmp:"-"`
}

func (t *Type) isBuiltin() bool {
	return t.Kind.IsBasic() && (t.PkgPath == "" || t.Kind == KindUnsafePointer)
}

func (t *Type) isDefined() bool {
	return t.PkgPath != "" && t.Name != ""
}

func (t *Type) isUnnamed() bool {
	return t.Name == ""
}

func (t *Type) isConstable() bool {
	return t.isDefined() && t.Kind.IsBasic()
}

func (t *Type) String() string {
	if len(t.Name) > 0 {
		if len(t.PkgPath) > 0 {
			return path.Base(t.PkgPath) + "." + t.Name
		}
		return t.Name
	}

	switch t.Kind {
	default: // assume builtin basic
		return t.Kind.String()
	case KindArray:
		return "[" + strconv.FormatInt(int64(t.ArrayLen), 10) + "]" + t.Elem.String()
	case KindSlice:
		return "[]" + t.Elem.String()
	case KindMap:
		return "map[" + t.Key.String() + "]" + t.Elem.String()
	case KindPtr:
		return "*" + t.Elem.String()
	case KindUint8:
		if t.IsByte {
			return "byte"
		}
		return "uint8"
	case KindInt32:
		if t.IsRune {
			return "rune"
		}
		return "int32"
	case KindStruct:
		if len(t.Fields) > 0 {
			return "struct{ /* ... */ }"
		}
		return "struct{}"
	case KindInterface:
		return "interface{}"
	case KindChan, KindFunc:
		return "<unsupported>"
	}
	return "<unknown>"
}

// CanSelectFields reports whether a value of Type t can be used in a field selector expression.
func (t *Type) CanSelectFields() bool {
	if t.Kind == KindStruct {
		return len(t.Fields) > 0
	}
	if t.Kind == KindPtr && t.Elem.Kind == KindStruct {
		return len(t.Elem.Fields) > 0
	}
	return false
}

// HasConstValues reports whether the t has associated const values.
func (t *Type) HasConstValues() bool {
	return t.Kind.IsBasic() && len(t.Values) > 0
}

// ElemHasConstValues reports whether the t's Elem type has associated const values.
func (t *Type) ElemHasConstValues() bool {
	return t.Elem != nil && t.Elem.Kind.IsBasic() && len(t.Elem.Values) > 0
}

// StructField describes a single struct field.
type StructField struct {
	// Name of the field.
	Name string
	// The field's type.
	Type *Type
	// The field's raw tag.
	Tag string
	// Indicates whether or not the field is embedded.
	IsEmbedded bool
	// Indicates whether or not the field is exported.
	IsExported bool
	// The position of the field's declaration in the source code.
	Pos Position
	// The raw documentation of a struct field.
	Doc []string
}

// ConstValue represents a typed constant.
type ConstValue struct {
	// The constant's identifier.
	Name string
	// The value of the constant.
	Value string
	// The position of the constant's declaration in the source code.
	Pos Position
	// The raw documentation of the constant.
	Doc []string
}

// The position of a token in the source code.
type Position struct {
	// The name of the file in the token is declared.
	Filename string `cmp:"+"`
	// The line at which the token is declared in the file.
	Line int `cmp:"+"`
}

// IsZero reports whether or not p is empty.
func (p Position) IsZero() bool { return p == Position{} }

// Kind indicates the specific kind of a Go type.
type Kind uint

const (
	// basic
	KindInvalid Kind = iota

	_basic_kind_start
	KindBool
	KindInt
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	KindUintptr
	KindFloat32
	KindFloat64
	KindComplex64
	KindComplex128
	KindString
	KindUnsafePointer
	_basic_kind_end

	// non-basic
	KindArray
	KindInterface
	KindMap
	KindPtr // 24
	KindSlice
	KindStruct

	// not supported
	KindChan
	KindFunc

	// alisases (basic)
	KindByte = KindUint8
	KindRune = KindInt32
)

// Reports whether or not k is of a basic kind.
func (k Kind) IsBasic() bool { return _basic_kind_start < k && k < _basic_kind_end }

// IsArrayOrSlice indicates whether or not k is an Array kind or a Slice kind.
func (k Kind) IsArrayOrSlice() bool { return k == KindArray || k == KindSlice }

func (k Kind) String() string {
	if int(k) < len(kindString) {
		return kindString[k]
	}
	return "<unknown> (types.Kind.String)"
}

var reflectKindToKind = [...]Kind{
	reflect.Invalid:       KindInvalid,
	reflect.Bool:          KindBool,
	reflect.Int:           KindInt,
	reflect.Int8:          KindInt8,
	reflect.Int16:         KindInt16,
	reflect.Int32:         KindInt32,
	reflect.Int64:         KindInt64,
	reflect.Uint:          KindUint,
	reflect.Uint8:         KindUint8,
	reflect.Uint16:        KindUint16,
	reflect.Uint32:        KindUint32,
	reflect.Uint64:        KindUint64,
	reflect.Uintptr:       KindUintptr,
	reflect.Float32:       KindFloat32,
	reflect.Float64:       KindFloat64,
	reflect.Complex64:     KindComplex64,
	reflect.Complex128:    KindComplex128,
	reflect.Array:         KindArray,
	reflect.Chan:          KindChan,
	reflect.Func:          KindFunc,
	reflect.Interface:     KindInterface,
	reflect.Map:           KindMap,
	reflect.Ptr:           KindPtr,
	reflect.Slice:         KindSlice,
	reflect.String:        KindString,
	reflect.Struct:        KindStruct,
	reflect.UnsafePointer: KindUnsafePointer,
}

var kindString = [...]string{
	KindInvalid:       "<invalid>",
	KindBool:          "bool",
	KindInt:           "int",
	KindInt8:          "int8",
	KindInt16:         "int16",
	KindInt32:         "int32",
	KindInt64:         "int64",
	KindUint:          "uint",
	KindUint8:         "uint8",
	KindUint16:        "uint16",
	KindUint32:        "uint32",
	KindUint64:        "uint64",
	KindUintptr:       "uintptr",
	KindFloat32:       "float32",
	KindFloat64:       "float64",
	KindComplex64:     "complex64",
	KindComplex128:    "complex128",
	KindArray:         "array",
	KindChan:          "chan",
	KindFunc:          "func",
	KindInterface:     "interface",
	KindMap:           "map",
	KindPtr:           "ptr",
	KindSlice:         "slice",
	KindString:        "string",
	KindStruct:        "struct",
	KindUnsafePointer: "unsafe.Pointer",
}
