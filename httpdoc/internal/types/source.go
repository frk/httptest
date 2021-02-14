package types

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sync"

	"golang.org/x/tools/go/packages"
)

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

type Source struct {
	fset *token.FileSet
	info *types.Info
	pkgs map[string]*packages.Package
	done map[reflect.Type]*Type
}

// not safe for concurrency
func Load(dir string) (*Source, error) {
	cache.Lock()
	defer cache.Unlock()

	if cache.source == nil {
		cache.source = make(map[string]*Source)
	}
	s, ok := cache.source[dir]
	if ok {
		return s, nil
	}

	conf := new(packages.Config)
	conf.Mode = loadMode
	conf.Dir = dir
	conf.Fset = token.NewFileSet()
	pkgs, err := packages.Load(conf, ".")
	if err != nil {
		return nil, err
	} else if len(pkgs) == 0 {
		return nil, errors.New("no package found in dir: " + dir)
	}
	s = &Source{
		fset: conf.Fset,
		info: pkgs[0].TypesInfo,
		pkgs: make(map[string]*packages.Package),
	}

	// save stores the given packages into cache. If the given packages
	// contain other imported packages then those will be stored as well,
	// and the imports of those packages will be stored too, and so on.
	var save func(pkgs ...*packages.Package)
	save = func(pkgs ...*packages.Package) {
		for _, pkg := range pkgs {
			if _, ok := s.pkgs[pkg.PkgPath]; ok {
				// skip if already present
				continue

			}

			s.pkgs[pkg.PkgPath] = pkg

			if len(pkg.Imports) > 0 {
				imports := make([]*packages.Package, 0, len(pkg.Imports))
				for _, pkg := range pkg.Imports {
					imports = append(imports, pkg)
				}
				save(imports...)
			}
		}
	}
	save(pkgs...)

	cache.source[dir] = s
	return s, nil
}

func (s *Source) position(pos token.Pos) Position {
	p := s.fset.Position(pos)
	return Position{Filename: p.Filename, Line: p.Line}
}

func (s *Source) getTypeSyntaxByName(name, path string) *TypeSyntax {
	pkg, ok := s.pkgs[path]
	if !ok {
		return nil
	}

	for _, syn := range pkg.Syntax {
		for _, dec := range syn.Decls {
			gd, ok := dec.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			for _, spec := range gd.Specs {
				s, ok := spec.(*ast.TypeSpec)
				if !ok || s.Name.Name != name {
					continue
				}

				return &TypeSyntax{
					Expr:    s.Type,
					DeclDoc: gd.Doc,
					SpecDoc: s.Doc,
					Comment: s.Comment,
					SpecPos: s.Pos(),
				}
			}
		}
	}
	return nil
}

func (s *Source) getTypeSyntaxById(id *ast.Ident) *TypeSyntax {
	obj, ok := s.info.Defs[id]
	if !ok {
		return nil
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil
	}

	name, path := tn.Name(), tn.Pkg().Path()
	return s.getTypeSyntaxByName(name, path)
}

var cache struct {
	sync.RWMutex
	source map[string]*Source
}
