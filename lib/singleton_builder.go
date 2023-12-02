package lib

import (
	"go/ast"
	"go/token"
	"strings"
)

type SingletonBuilder struct {
	Singleton        *ast.FuncDecl
	SingletonMethods []*Method
}

func (s *SingletonBuilder) IsRequired() bool {
	return s.Singleton != nil
}

func (s *SingletonBuilder) Build() []ast.Decl {
	s.Singleton.Body.List = append(s.Singleton.Body.List, &ast.ReturnStmt{})

	var names []string
	for _, v := range s.Singleton.Type.Results.List {
		names = append(names, v.Names[0].Name)
	}
	var decls []ast.Decl
	decls = append(decls, s.Singleton)
	decls = append(decls, &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			getDeclaredVariable(names, "", []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("diSingleton"),
				},
			}),
		},
	})
	return decls
}

func (s *SingletonBuilder) Add(src string, pkg *Package) {

	if s.Singleton == nil {
		pkg.addImport("sync", "")
		s.Singleton = &ast.FuncDecl{
			Name: ast.NewIdent("diSingleton"),
			Type: &ast.FuncType{
				Params:  &ast.FieldList{},
				Results: &ast.FieldList{},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{},
			},
		}
	}
	fnName := "Get" + src
	method := GetMethod("New"+src, pkg.name)
	if method == nil {
		panic("the specified method doesn't exists: " + "New" + src)

	}
	if len(method.Ast.Type.Params.List) != 0 {
		panic("method used for singleton should not have any parameters" + method.Name)
	}
	s.SingletonMethods = append(s.SingletonMethods, &Method{
		Name:    fnName,
		Results: method.Results,
		Ast: &ast.FuncDecl{
			Name: ast.NewIdent(fnName),
			Type: method.Ast.Type,
		},
	})

	s.Singleton.Type.Results.List = append(s.Singleton.Type.Results.List, &ast.Field{
		Names: []*ast.Ident{
			ast.NewIdent(fnName),
		},
		Type: method.Ast.Type,
	})
	specs := []ast.Spec{
		getDeclaredVariable([]string{getVarName("sync" + src)}, "sync.Once", []ast.Expr{}),
	}
	for _, v := range method.Ast.Type.Results.List {
		n := getName(v.Type)
		if n != src {
			n = src + strings.Title(n)
		}
		specs = append(specs, &ast.ValueSpec{
			Type:  v.Type,
			Names: []*ast.Ident{ast.NewIdent(getVarName(n))},
		})
		// specs = append(specs, getDeclaredVariable([]string{getVarName(n)}, getName(v.Type), []ast.Expr{}))
	}

	// generate var stmt
	s.Singleton.Body.List = append(s.Singleton.Body.List, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok:   token.VAR,
			Specs: specs,
		},
	})
	fcall := functionCall(method, true, "New", true)
	fcall.(*ast.AssignStmt).Tok = token.ASSIGN
	// generate singleton func for given variable
	s.Singleton.Body.List = append(s.Singleton.Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(fnName),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.FuncLit{
				Type: method.Ast.Type,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent(getVarName("sync" + src)),
									Sel: ast.NewIdent("Do"),
								},
								Args: []ast.Expr{
									&ast.FuncLit{
										Type: &ast.FuncType{
											Results: &ast.FieldList{},
											Params:  &ast.FieldList{},
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												fcall.(*ast.AssignStmt),
											},
										},
									},
								},
							},
						},
						&ast.ReturnStmt{
							Results: fcall.(*ast.AssignStmt).Lhs,
						},
					},
				},
			},
		},
	})
}
