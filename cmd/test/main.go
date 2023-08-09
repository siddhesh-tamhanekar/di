package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type GeneratedFile struct {
	imports []*ast.ImportSpec
	fns     []*ast.FuncDecl
}

func getDeclaredVariable(names []string, typ string, values []ast.Expr) *ast.ValueSpec {
	var nameIdentifers []*ast.Ident
	for _, v := range names {
		nameIdentifers = append(nameIdentifers, ast.NewIdent(v))
	}
	return &ast.ValueSpec{
		Names:  nameIdentifers,
		Type:   ast.NewIdent(typ),
		Values: values,
	}

}

func singleton() (GetDb func() Db) {
	var db *Db
	GetDb = func() Db {
		if db == nil {
			db1 := NewDb()
			db = &db1
		}
		return *db
	}
}

func main() {
	// fset := token.NewFileSet()
	file := ast.File{}

	file.Decls = append(file.Decls, &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{
		&ast.ImportSpec{
			Path: &ast.BasicLit{Value: "\"github.com/spew/spew\""},
		},
		&ast.ImportSpec{
			Path: &ast.BasicLit{Value: "\"github.com/spew/spew/di\""},
		},
	}})

	file.Decls = append(file.Decls, &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
		getDeclaredVariable([]string{"a"}, "", []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("NewDb"),
			}}),
	}})

	fn := ast.FuncDecl{
		Name: ast.NewIdent("singleton"),

		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.DeclStmt{
					Decl: &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							getDeclaredVariable([]string{"abc1"}, "string", nil),
							getDeclaredVariable([]string{"pqr"}, "*int", nil),
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, &fn)
	file.Name = ast.NewIdent("main")

	fset := token.NewFileSet()
	buf := new(strings.Builder)
	printer.Fprint(buf, fset, &file)
	fmt.Println(buf.String())
	fset1 := token.NewFileSet()

	f, err := parser.ParseFile(fset1, "./cmd/test/di_gen.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println("err", err)
	}
	_ = f
	spew.Dump(f)

	file1, err := os.OpenFile("./cmd/test/di_generated.go", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Write error", err)
	}
	fstr := strings.ReplaceAll(buf.String(), "EOF\n", "")
	file1.Write([]byte(fstr))
	file1.Close()

}

type Db int

func NewDb() Db {
	return 1
}

func NewToken() int {
	return 2
}
