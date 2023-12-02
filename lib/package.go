package lib

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

var pkgs map[string]*Package

func GetStructOrInterface(s string, p string) (*Struct, *Interface) {
	pkg, ok := pkgs[p]
	if ok {

		for _, st := range pkg.Structs {
			if st.Name == s {
				return st, nil
			}
		}
		for _, st := range pkg.Interfaces {
			if st.Name == s {
				return nil, st
			}
		}
	}

	return nil, nil
}

func GetStructMethod(s string, method string, packageName string) *Method {
	pkg, ok := pkgs[packageName]
	if ok {
		for _, v := range pkg.Methods {
			if v.Name == method && v.Reciever != nil {
				recv := v.Reciever.T
				if recv[0:1] == "*" && recv[1:] == s {
					return v
				} else if recv == s {
					return v
				}
			}
		}
	}
	return nil
}

func GetMethod(method string, packageName string) *Method {
	pkg, ok := pkgs[packageName]
	if ok {
		for _, v := range pkg.Methods {
			if v.Name == method && v.Reciever == nil {
				return v
			}
		}
	}
	return nil
}

type Package struct {
	name       string
	path       string
	Structs    []*Struct
	Interfaces []*Interface
	Methods    []*Method

	Imports []string // depreciated
	Vars    []string
	Shared  map[string]string // depreciated

	ImportSpecs []ast.Spec
	SharedVars  []ast.Spec
	Funcs       []*ast.FuncDecl
	singleton   SingletonBuilder
}

func NewPackage(file *CodeFile) *Package {
	return &Package{
		name:        file.Package,
		path:        file.Path,
		Shared:      make(map[string]string),
		Vars:        make([]string, 0),
		ImportSpecs: make([]ast.Spec, 0),
	}

}
func (pkg *Package) AddParsedFile(file *CodeFile) {
	pkg.Structs = append(pkg.Structs, file.Structs...)
	pkg.Methods = append(pkg.Methods, file.Methods...)
	pkg.Interfaces = append(pkg.Interfaces, file.Interfaces...)

}
func (pkg *Package) addImport(path string, name string) {
	imp := path
	mod := getModule()
	if strings.Contains(path, ".go") {
		imp = path[:strings.LastIndex(path, "/")]
		imp = strings.ReplaceAll(imp, "./", mod+"/")

	}
	if name != "" && pkg.name == name {
		return
	}
	// fmt.Println("adding import", imp, "into", pkg.name)
	importStmt := &ast.ImportSpec{Path: &ast.BasicLit{Value: "\"" + imp + "\""}}
	if name != "" {
		importStmt.Name = ast.NewIdent(name)
	}
	pkg.ImportSpecs = append(pkg.ImportSpecs, importStmt)

}

func (pkg *Package) generateFile() {
	file := ast.File{}
	fset := token.NewFileSet()
	file.Name = ast.NewIdent(pkg.name)

	// add imports
	if len(pkg.ImportSpecs) > 0 {
		file.Decls = append(file.Decls, &ast.GenDecl{Tok: token.IMPORT, Specs: pkg.ImportSpecs})
	}

	// add package level variables
	if len(pkg.SharedVars) > 0 {
		file.Decls = append(file.Decls, &ast.GenDecl{Tok: token.VAR, Specs: pkg.SharedVars})
	}
	if pkg.singleton.IsRequired() {
		// fmt.Println("adding singleton function")
		file.Decls = append(file.Decls, pkg.singleton.Build()...)
	}
	if len(pkg.Funcs) > 0 {
		for _, v := range pkg.Funcs {
			file.Decls = append(file.Decls, v)
		}
	}
	fp := pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_generated.go"
	if len(file.Decls) == 0 {
		// fmt.Println("No Need to generate file", fp)
		return
	}

	pkg.addGeneratedComments(&file)
	buf := new(strings.Builder)
	// fmt.Println(buf.String())

	printer.Fprint(buf, fset, &file)
	fstr := strings.ReplaceAll(buf.String(), "EOF\n", "")
	fstr = strings.ReplaceAll(fstr, "}\nfunc ", "}\n\nfunc ")
	// fmt.Println(fstr)
	b := []byte(fstr)
	emptyfile := fmt.Sprintf("package %s\n\n// Code generated by DI library. DO NOT EDIT.\n// To generate file use <path_to_di>/di --path= --module=\n", pkg.name)

	// for _, v := range pkg.Fns {
	// 	b = append(b, []byte(v.code)...)
	// }
	if emptyfile == string(b) {
		fmt.Println("No Need to generate file v2 ", fp)
		return
	}
	// fmt.Println(string(b))
	b, err := format.Source(b)
	if err != nil {
		fmt.Println("format error", err)
		return
	}

	os.Remove(fp)
	WriteFile(fp, b)

}

func (pkg *Package) addGeneratedComments(file *ast.File) {

	file.Decls = append([]ast.Decl{&ast.GenDecl{Tok: token.EOF, Specs: []ast.Spec{
		&ast.ValueSpec{
			Comment: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// Code generated by DI library. DO NOT EDIT.\n"},
					{Text: "// To generate file use <path_to_di>/di --path= --module="},
				},
			},
		},
	}}}, file.Decls...)
}

func (pkg *Package) cleanup() {
	fp := pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_gen.go"
	os.Remove(fp)
	fp = pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_generated.go"
	os.Remove(fp)

}
