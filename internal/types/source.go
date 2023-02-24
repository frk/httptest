package types

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
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

// not safe for concurrency
type Source struct {
	pkgs map[string]*packages.Package
}

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
	s = &Source{pkgs: make(map[string]*packages.Package)}

	s.savepkgs(pkgs...)

	cache.source[dir] = s
	return s, nil
}

// findpkg searches for a *packages.Package by the given pkgpath and returns it if a match is found.
func (s *Source) findpkg(pkgpath string) (*packages.Package, error) {
	pkg, ok := s.pkgs[pkgpath]
	if ok {
		return pkg, nil
	}

	conf := new(packages.Config)
	conf.Mode = loadMode
	conf.Fset = token.NewFileSet()
	pkgs, err := packages.Load(conf, pkgpath)
	if err != nil {
		return nil, err
	} else if len(pkgs) == 0 {
		return nil, errors.New("no package found with path: " + pkgpath)
	}

	s.savepkgs(pkgs...)

	return pkgs[0], nil
}

// savepkgs stores the given packages into cache. If the given packages
// contain other imported packages then those will be stored as well,
// and the imports of those packages will be stored too, and so on.
func (s *Source) savepkgs(pkgs ...*packages.Package) {
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
			s.savepkgs(imports...)
		}
	}
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
	// The package in which the type is declared.
	pkg *packages.Package
}

func (s *typeSource) position() Position {
	p := s.pkg.Fset.Position(s.SpecPos)
	return Position{Filename: p.Filename, Line: p.Line}
}

func (s *typeSource) positionForNode(node ast.Node) Position {
	p := s.pkg.Fset.Position(node.Pos())
	return Position{Filename: p.Filename, Line: p.Line}
}

func (s *Source) getTypeSourceByName(name, path string) *typeSource {
	pkg, err := s.findpkg(path)
	if err != nil {
		fmt.Printf("findpkg: %v\n", err)
		return nil
	}

	// remove type-args from generic-type name
	if i := strings.IndexByte(name, '['); i > -1 {
		name = name[:i]
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
					pkg:     pkg,
				}
			}
		}
	}

	return nil
}

func (s *Source) getTypeSourceById(id *ast.Ident, pkg *packages.Package) *typeSource {
	obj, ok := pkg.TypesInfo.Defs[id]
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
	// The package in which the constant is declared.
	pkg *packages.Package
}

func (s *constSource) position() Position {
	p := s.pkg.Fset.Position(s.SpecPos)
	return Position{Filename: p.Filename, Line: p.Line}
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
								pkg:     pkg,
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
