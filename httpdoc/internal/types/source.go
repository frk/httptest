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

// source info of a type
type typeSource struct {
	// the type expression: *ast.Ident, *ast.ParenExpr, *ast.SelectorExpr,
	// *ast.StarExpr, or any of the *ast.XxxTypes
	Expr ast.Expr
	// documentation associated with the type declaration; or nil
	DeclDoc *ast.CommentGroup
	// documentation associated with the type spec; or nil
	SpecDoc *ast.CommentGroup
	// line comments from spec; or nil
	Comment *ast.CommentGroup
	// the type spec's position
	SpecPos token.Pos
}

func (s *Source) getTypeSourceByName(name, path string) *typeSource {
	pkg, ok := s.pkgs[path]
	if !ok {
		return nil
	}

	for _, syn := range pkg.Syntax {
		for _, d := range syn.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			for _, spec := range gd.Specs {
				s, ok := spec.(*ast.TypeSpec)
				if !ok || s.Name.Name != name {
					continue
				}

				return &typeSource{
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

func (s *Source) getTypeSourceById(id *ast.Ident) *typeSource {
	obj, ok := s.info.Defs[id]
	if !ok {
		return nil
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil
	}

	name, path := tn.Name(), tn.Pkg().Path()
	return s.getTypeSourceByName(name, path)
}

// source info of a const value
type constSource struct {
	Const *types.Const
	// documentation associated with the const declaration; or nil
	DeclDoc *ast.CommentGroup
	// documentation associated with the const spec; or nil
	SpecDoc *ast.CommentGroup
	// line comments from spec; or nil
	Comment *ast.CommentGroup
	// the const spec's position
	SpecPos token.Pos
}

// getConstSourceByTypeName scans the given Source.pkgs looking for all declared constants
// of the type identified by name and path (package path). On success the result will
// be a slice of go/types.Const instances that represent those constants.
func (s *Source) getConstSourceByTypeName(name, path string) (consts []*constSource) {
	for _, pkg := range s.pkgs {
		if pkg.PkgPath != path {
			if _, ok := pkg.Imports[path]; !ok {
				// If pkg is not the target package, and it also
				// does not import the target package, go to next
				continue
			}
		}

		for _, syn := range pkg.Syntax {
			for _, d := range syn.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok || gd.Tok != token.CONST {
					continue
				}

				for _, spec := range gd.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}

					for _, id := range vs.Names {
						if id.Name == "_" {
							continue
						}

						obj, ok := pkg.TypesInfo.Defs[id]
						if !ok {
							continue
						}

						if c, ok := obj.(*types.Const); ok {
							named, ok := c.Type().(*types.Named)
							if !ok {
								continue
							}

							tn := named.Obj()
							if tn.Name() != name || tn.Pkg().Path() != path {
								continue
							}

							consts = append(consts, &constSource{
								Const:   c,
								DeclDoc: gd.Doc,
								SpecDoc: vs.Doc,
								Comment: vs.Comment,
								SpecPos: vs.Pos(),
							})
						}
					}
				}
			}

		}
	}
	return consts
}

var cache struct {
	sync.RWMutex
	source map[string]*Source
}
