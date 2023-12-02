package lib

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"strings"
)

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func getVarName(name string) string {
	if name[0:1] == "*" {
		name = name[1:]

	}
	return strings.ToLower(name[0:1]) + name[1:]
}

func getCurrentModuleName() string {
	dir, _ := os.Getwd()
	gomodfile := dir + "/" + "go.mod"
	if _, err := os.Stat(gomodfile); errors.Is(err, os.ErrNotExist) {
		fmt.Println("go module does not exists")

	}
	gmodfile, err := os.ReadFile(gomodfile)
	if err != nil {
		log.Panic("go module file is not readable")
	}
	modline := strings.Split(string(gmodfile), "\n")[0]
	return strings.ReplaceAll(modline, "module ", "")

}

func WriteFile(fp string, b []byte) {
	fmt.Println("GENERATED CODE FOR", fp)

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Write error", err)
	}
	f.Write(b)
	f.Close()

}

func ContainsStr(strings []string, needle string) bool {
	for _, v := range strings {
		if v == needle {
			return true
		}
	}
	return false
}

func getModule() string {
	mod := ""
	if mod == "" {
		mod = getCurrentModuleName()
	}
	return mod
}

var Debug bool

func logMsg(msg string) {
	if Debug == true {
		fmt.Println(msg)

	}
}

// ---------------------------------
// Ast helper functions
// ---------------------------------
func getDeclaredVariable(names []string, typ string, values []ast.Expr) *ast.ValueSpec {
	var nameIdentifers []*ast.Ident
	for _, v := range names {
		nameIdentifers = append(nameIdentifers, ast.NewIdent(v))
	}
	if len(values) <= 0 {
		return &ast.ValueSpec{
			Names: nameIdentifers,
			Type:  ast.NewIdent(typ),
		}

	}
	return &ast.ValueSpec{
		Names:  nameIdentifers,
		Type:   ast.NewIdent(typ),
		Values: values,
	}

}

func getName(n ast.Node) string {
	switch n.(type) {
	case *ast.StarExpr:
		return getName(n.(*ast.StarExpr).X)
	case *ast.SelectorExpr:
		return getName(n.(*ast.SelectorExpr).Sel)
	case *ast.Ident:
		return n.(*ast.Ident).Name
	case *ast.UnaryExpr:
		return getName(n.(*ast.UnaryExpr).X)

	}
	return ""
}

func getPackageName(n ast.Node) string {
	switch n.(type) {
	case *ast.StarExpr:
		return getPackageName(n.(*ast.StarExpr).X)
	case *ast.UnaryExpr:
		return getPackageName(n.(*ast.UnaryExpr).X)
	case *ast.SelectorExpr:
		return n.(*ast.SelectorExpr).X.(*ast.Ident).Name
	}
	return ""
}

func functionCall(method *Method, isStoreResult bool, prefix string, errorName bool) ast.Node {
	p := strings.ReplaceAll(method.Name, prefix, "")
	call := &ast.CallExpr{
		Fun:  ast.NewIdent(method.Name),
		Args: []ast.Expr{},
	}
	for _, v := range method.Ast.Type.Params.List {
		name := getVarName(p) + strings.Title(getName(v.Names[0]))
		if strings.EqualFold(getName(v.Names[0]), p) {
			name = getVarName(getName(v.Names[0]))
		}
		call.Args = append(call.Args, ast.NewIdent(name))
	}

	if !isStoreResult {
		return &ast.ExprStmt{
			X: call,
		}
	}
	lhs := []ast.Expr{}
	for _, v := range method.Ast.Type.Results.List {
		name := getVarName(p) + strings.Title(getName(v.Type))
		if strings.EqualFold(getName(v.Type), p) {
			name = getVarName(getName(v.Type))
		}
		if !errorName && getName(v.Type) == "error" {
			name = "err"
		}
		lhs = append(lhs, ast.NewIdent(name))
	}

	return &ast.AssignStmt{
		Rhs: []ast.Expr{call},
		Lhs: lhs,
		Tok: token.DEFINE,
	}
}
