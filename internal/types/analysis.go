package types

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ = fmt.Println

func (s *Source) TypeOf(v interface{}) *Type {
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	a := &analysis{
		src:   s,
		ptrs:  make(map[interface{}]*Type),
		done:  make(map[reflect.Type]*Type),
		idone: make(map[reflect.Type]*Type),
	}

	typ := typeInfo(a, rt, rv, nil)
	s.analyzeTypeSource(typ, nil, nil)

	// TODO cache any *Type in typ's hierarchy that does not have an interface

	return typ
}

// holds the state of the analysis of a type
type analysis struct {
	src *Source
	// already visited pointers
	ptrs map[interface{}]*Type
	// Already analyzed types; only types that aren't,
	// and don't contain, an interface type.
	done map[reflect.Type]*Type
	// Already analyzed types; only of invalid values.
	idone map[reflect.Type]*Type
}

func typeInfo(a *analysis, rt reflect.Type, rv reflect.Value, rts []reflect.Type) *Type {
	kind := rt.Kind()
	if kind == reflect.Ptr && rv.IsValid() {
		if typ, ok := a.ptrs[rv.Interface()]; ok {
			return typ
		}
	}
	if kind != reflect.Interface {
		if typ, ok := a.done[rt]; ok {
			return typ
		}
	}
	if !rv.IsValid() {
		if typ, ok := a.idone[rt]; ok {
			return typ
		}
	}

	typ := new(Type)
	typ.Name = rt.Name()
	typ.Kind = reflectKindToKind[rt.Kind()]
	typ.PkgPath = rt.PkgPath()

	if kind == reflect.Ptr && rv.IsValid() {
		a.ptrs[rv.Interface()] = typ
	}
	if !rv.IsValid() {
		a.idone[rt] = typ
	}

	switch typ.Kind {
	case KindArray:
		rvElem := reflect.Value{}
		if rv.IsValid() && rv.Len() > 0 {
			rvElem = rv.Index(0)
			// TODO? if the element type's an interface it may be
			// desirable to produce a "union type" by analyzing each
			// individual element instead of just the first one.
		}
		typ.Elem = typeInfo(a, rt.Elem(), rvElem, rts)
		typ.ArrayLen = rt.Len()
		typ.hasiface = typ.Elem.hasiface
	case KindMap:
		rvKey := reflect.Value{}
		rvElem := reflect.Value{}
		if rv.IsValid() && rv.Len() > 0 {
			keys := rv.MapKeys()

			rvKey = keys[0]
			rvElem = rv.MapIndex(keys[0])

			// TODO? if the key or element type is an interface it
			// may be desirable to produce a "union type" by analyzing
			// each individual key/element instead of just the first one.
		}
		typ.Key = typeInfo(a, rt.Key(), rvKey, rts)
		typ.Elem = typeInfo(a, rt.Elem(), rvElem, rts)
		typ.hasiface = (typ.Key.hasiface || typ.Elem.hasiface)
	case KindPtr:
		rvElem := reflect.Value{}
		if rv.IsValid() {
			rvElem = rv.Elem()
		}
		typ.Elem = typeInfo(a, rt.Elem(), rvElem, rts)
		typ.hasiface = typ.Elem.hasiface
	case KindSlice:
		rvElem := reflect.Value{}
		if rv.IsValid() && rv.Len() > 0 {
			rvElem = rv.Index(0)

			// TODO? if the element type's an interface it may be
			// desirable to produce a "union type" by analyzing each
			// individual element instead of just the first one.
		}
		typ.Elem = typeInfo(a, rt.Elem(), rvElem, rts)
		typ.hasiface = typ.Elem.hasiface
	case KindStruct:
		num := rt.NumField()
		typ.Fields = make([]*StructField, num)

		for i := 0; i < num; i++ {
			ft := rt.Field(i)
			fv := reflect.Value{}
			if rv.IsValid() {
				fv = rv.Field(i)
			}

			sf := new(StructField)
			sf.Name = ft.Name
			sf.Tag = string(ft.Tag)
			sf.Type = typeInfo(a, ft.Type, fv, rts)
			sf.IsEmbedded = ft.Anonymous
			sf.IsExported = len(ft.PkgPath) > 0
			typ.Fields[i] = sf
			typ.hasiface = (typ.hasiface || sf.Type.hasiface)
		}
	case KindInterface:
		typ.hasiface = true
		if rv.IsValid() && rv.Elem().IsValid() {
			rv := rv.Elem()
			typ.Elem = typeInfo(a, rv.Type(), rv, rts)
		}
	}

	return typ
}

func (s *Source) analyzeTypeSource(t *Type, src *typeSource, visited map[*Type]bool) {
	if visited == nil {
		visited = make(map[*Type]bool)
	}

	if t == nil || visited[t] || t.isBuiltin() { // nothing to do?
		return
	} else if src == nil && t.PkgPath != "" && t.Name != "" {
		src = s.getTypeSourceByName(t.Name, t.PkgPath)
	}
	visited[t] = true

	if src == nil {
		switch t.Kind {
		case KindPtr, KindArray, KindSlice:
			s.analyzeTypeSource(t.Elem, nil, visited)
		case KindMap:
			s.analyzeTypeSource(t.Key, nil, visited)
			s.analyzeTypeSource(t.Elem, nil, visited)
		case KindInterface:
			if t.Elem != nil {
				s.analyzeTypeSource(t.Elem, nil, visited)
			}
		case KindStruct:
			for _, f := range t.Fields {
				s.analyzeTypeSource(f.Type, nil, visited)
			}
		}
		return
	}

	// source code position & documentation of the type
	if src.SpecPos != token.NoPos {
		pos := src.position()
		cg, doc := src.SpecDoc, []string(nil)
		if cg == nil || len(cg.List) == 0 {
			cg = src.DeclDoc
		}
		if cg == nil || len(cg.List) == 0 {
			cg = src.Comment
		}
		if cg != nil {
			for _, s := range cg.List {
				doc = append(doc, s.Text)
			}
		}
		t.Pos = pos
		t.Doc = doc
	}

	switch x := src.Expr.(type) {
	case *ast.ParenExpr:
		s.analyzeTypeSource(t, &typeSource{Expr: x.X, pkg: src.pkg}, visited)
	case *ast.Ident:
		if src := s.getTypeSourceById(x, src.pkg); src != nil {
			s.analyzeTypeSource(t, &typeSource{Expr: src.Expr, pkg: src.pkg}, visited)
		}
	case *ast.SelectorExpr:
		if src := s.getTypeSourceById(x.Sel, src.pkg); src != nil {
			s.analyzeTypeSource(t, &typeSource{Expr: src.Expr, pkg: src.pkg}, visited)
		}
	case *ast.StarExpr:
		if t.Kind != KindPtr {
			panic("shouldn't happen")
		}
		if t.Elem.isDefined() {
			s.analyzeTypeSource(t.Elem, nil, visited)
		} else {
			s.analyzeTypeSource(t.Elem, &typeSource{Expr: x.X, pkg: src.pkg}, visited)
		}
	case *ast.InterfaceType:
		if t.Elem != nil {
			s.analyzeTypeSource(t.Elem, nil, visited)
		}
	case *ast.ArrayType:
		if t.Elem.isDefined() {
			s.analyzeTypeSource(t.Elem, nil, visited)
		} else {
			s.analyzeTypeSource(t.Elem, &typeSource{Expr: x.Elt, pkg: src.pkg}, visited)
		}
	case *ast.MapType:
		if t.Key.isDefined() {
			s.analyzeTypeSource(t.Key, nil, visited)
		} else {
			s.analyzeTypeSource(t.Key, &typeSource{Expr: x.Key, pkg: src.pkg}, visited)
		}
		if t.Elem.isDefined() {
			s.analyzeTypeSource(t.Elem, nil, visited)
		} else {
			s.analyzeTypeSource(t.Elem, &typeSource{Expr: x.Value, pkg: src.pkg}, visited)
		}
	case *ast.StructType:
		i := 0
		for _, fi := range x.Fields.List {
			// source code position & documentation of the field
			pos := src.positionForNode(fi)
			cg, doc := fi.Doc, []string(nil)
			if cg == nil || len(cg.List) == 0 {
				cg = fi.Comment
			}
			if cg != nil {
				for _, s := range cg.List {
					doc = append(doc, s.Text)
				}
			}

			for range fi.Names {
				t.Fields[i].Pos = pos
				t.Fields[i].Doc = doc
				if t.Fields[i].Type.isDefined() {
					s.analyzeTypeSource(t.Fields[i].Type, nil, visited)
				} else {
					s.analyzeTypeSource(t.Fields[i].Type, &typeSource{Expr: fi.Type, pkg: src.pkg}, visited)
				}
				i += 1
			}
		}
	}

	if t.isConstable() {
		s.analyzeConstSource(t)
	}
}

func (s *Source) analyzeConstSource(t *Type) {
	consts := s.getConstSourceByTypeName(t.Name, t.PkgPath)
	if len(consts) == 0 {
		return
	}

	t.Values = make([]*ConstValue, len(consts))
	for i, src := range consts {
		t.Values[i] = new(ConstValue)
		t.Values[i].Name = src.Const.Name()
		t.Values[i].Value = src.Const.Val().ExactString()

		// source code position & documentation of the constant
		pos := src.position()
		cg, doc := src.SpecDoc, []string(nil)
		if cg == nil || len(cg.List) == 0 {
			cg = src.Comment
		}
		if cg == nil || len(cg.List) == 0 {
			cg = src.DeclDoc
		}
		if cg != nil {
			for _, s := range cg.List {
				doc = append(doc, s.Text)
			}
		}
		t.Values[i].Pos = pos
		t.Values[i].Doc = doc
	}
}
