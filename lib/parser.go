package lib

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

type CodeFile struct {
	Package    string
	Path       string
	Name       string
	Structs    []*Struct
	Methods    []*Method
	Interfaces []*Interface
	Imports    map[string]string
}

type Interface struct {
	File    *CodeFile
	Name    string
	Methods []*Method
}

type Type struct {
	Name string
	T    string
}

type Field struct {
	Struct *Struct
	Name   string
	Type   string
	Tag    string
}

type Method struct {
	File     *CodeFile
	Name     string
	Params   []*Type
	Results  []*Type
	Reciever *Type
}

type Struct struct {
	Name   string
	File   *CodeFile
	Fields []*Field
}
type ParseVisitor struct {
	v    int
	file *CodeFile
}

func getTypes(list *ast.FieldList) []*Type {
	var ts []*Type
	if list == nil {
		return ts
	}
	for _, f := range list.List {
		fset := token.NewFileSet()
		buf := new(strings.Builder)
		printer.Fprint(buf, fset, f.Type)
		n := ""
		if len(f.Names) > 0 {
			n = f.Names[0].Name
		}
		t := Type{
			Name: n,
			T:    buf.String(),
		}
		ts = append(ts, &t)
	}
	return ts

}
func (v ParseVisitor) Visit(n ast.Node) ast.Visitor {

	if n == nil {
		return nil
	}
	switch n.(type) {
	case *ast.ImportSpec:
		name := ""
		if n.(*ast.ImportSpec).Name != nil {
			name = n.(*ast.ImportSpec).Name.Name

		}
		path := strings.Trim(n.(*ast.ImportSpec).Path.Value, "\"")
		if name == "" {
			idx := strings.LastIndex(path, "/") + 1
			if idx < 0 {
				name = path
			} else {
				name = path[idx:]

			}
		}
		v.file.Imports[name] = path
	case *ast.FuncDecl:
		f := n.(*ast.FuncDecl)
		m := Method{
			Name:    f.Name.Name,
			File:    v.file,
			Params:  getTypes(f.Type.Params),
			Results: getTypes(f.Type.Results),
		}
		ts := getTypes(f.Recv)

		if len(ts) > 0 {
			m.Reciever = ts[0]
		}

		v.file.Methods = append(v.file.Methods, &m)
	case *ast.TypeSpec:
		tSpec := n.(*ast.TypeSpec)
		sType, ok := tSpec.Type.(*ast.StructType)
		if ok {
			s := Struct{
				Name:   tSpec.Name.Name,
				File:   v.file,
				Fields: make([]*Field, 0),
			}
			for _, f := range sType.Fields.List {
				fset := token.NewFileSet()
				buf := new(strings.Builder)
				printer.Fprint(buf, fset, f.Type)

				fs := Field{
					Struct: &s,
					Name:   f.Names[0].Name,
					Type:   buf.String(),
				}
				if f.Tag != nil {
					fs.Tag = f.Tag.Value
				}
				s.Fields = append(s.Fields, &fs)
			}
			v.file.Structs = append(v.file.Structs, &s)

		}
		iType, ok := tSpec.Type.(*ast.InterfaceType)
		if ok {
			s := Interface{
				Name:    tSpec.Name.Name,
				File:    v.file,
				Methods: make([]*Method, 0),
			}
			for _, f := range iType.Methods.List {
				m, ok := f.Type.(*ast.FuncType)
				if ok {
					fs := Method{
						Name:    f.Names[0].Name,
						Params:  getTypes(m.Params),
						Results: getTypes(m.Results),
					}
					s.Methods = append(s.Methods, &fs)
				}
			}
			v.file.Interfaces = append(v.file.Interfaces, &s)

		}

	}
	v.v = v.v + 1
	return v

}

func parseFile(file string) *ast.File {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("parsing file err", err, f)
	}
	return f
}

func parseGoFile(file string) *CodeFile {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("parsing file err", err, f)
	}

	return NewCodeFile(f, file)
}

func NewCodeFile(f *ast.File, path string) *CodeFile {
	codeFile := CodeFile{
		Package: f.Name.Name,
		Path:    path,
		Imports: make(map[string]string),
	}
	p := ParseVisitor{
		v:    0,
		file: &codeFile,
	}
	ast.Walk(p, f)
	return &codeFile
}
