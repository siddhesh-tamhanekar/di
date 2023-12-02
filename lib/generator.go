package lib

import (
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"
)

var Generators []Generator
var GlobalScope *Scope

func init() {
	GlobalScope = NewScope("global")
}
func CreateGenerator(method string) Generator {
	if method == "Build" {
		return &StructGenerator{}
	}
	if method == "Bind" || method == "BindEnv" {
		return &InterfaceGenerator{}
	}
	if method == "Share" {
		return &SharedGenerator{}
	}
	if method == "Singleton" {
		return &SingletonGenerator{}
	}
	return nil
}

func GetGenerator(name string) Generator {
	for _, gen := range Generators {
		if gen.Name() == name {
			return gen
		}
	}
	return nil
}

type Generator interface {
	ParseDeclaration(n ast.Node, pkg string)
	Generate() (*ast.FuncDecl, string)
	Name() string
}

/**
 ****************************
 * Interace generator
 ****************************
 */
type InterfaceGenerator struct {
	pkg        string
	src        string
	env        string
	inter      string
	isEnvBased bool
	elts       []ast.Expr
}

func (ig *InterfaceGenerator) Name() string {
	return strings.Title(ig.env) + ig.inter
}

func (ig *InterfaceGenerator) ParseDeclaration(n ast.Node, pkg string) {
	ig.pkg = pkg

	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		panic("not given call expr to interface builder")
	}
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if sel.Sel.Name == "BindEnv" {
		ig.isEnvBased = true
		env, ok := callExpr.Args[2].(*ast.BasicLit)
		if ok {
			ig.env = strings.Trim(env.Value, "\"")
		}

	}
	if ok {
		// set src
		switch callExpr.Args[1].(type) {
		case *ast.CompositeLit:
			switch callExpr.Args[1].(*ast.CompositeLit).Type.(type) {
			case *ast.Ident:
				ig.src = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				ig.src = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
				ig.pkg = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			case *ast.MapType:
				ig.elts = callExpr.Args[1].(*ast.CompositeLit).Elts
			}

		case *ast.BasicLit:
			ig.src = strings.Trim(callExpr.Args[1].(*ast.BasicLit).Value, "\"")
		case *ast.Ident:
			ig.src = callExpr.Args[1].(*ast.Ident).Name
		}
		// set inter
		switch callExpr.Args[0].(type) {
		case *ast.CompositeLit:
			switch callExpr.Args[0].(*ast.CompositeLit).Type.(type) {
			case *ast.Ident:
				ig.inter = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				ig.inter = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
				ig.pkg = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			}
		case *ast.BasicLit:
			ig.inter = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
		case *ast.Ident:
			ig.inter = callExpr.Args[0].(*ast.Ident).Name
		}

	}
}

func (ig *InterfaceGenerator) Generate() (*ast.FuncDecl, string) {
	if ig.isEnvBased && os.Getenv("ENV") != ig.env {
		// fmt.Println("returning", v)
		return nil, ""
	}
	fn := &ast.FuncDecl{
		Name: ast.NewIdent("New" + ig.inter),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	if ig.elts == nil {
		s, _ := GetStructOrInterface(ig.src, ig.pkg)

		_, i := GetStructOrInterface(ig.inter, ig.pkg)

		if len(i.Methods) > 0 {
			n := i.Methods[0].Name
			m := GetStructMethod(s.Name, n, ig.pkg)
			if m == nil {
				log.Panic("struct " + s.Name + " doesnt have " + n + " method")
			}
		}
		CreateStruct(s, fn, pkgs[ig.pkg], nil)
		ret := fn.Type.Results.List[0]
		ret.Type = ast.NewIdent(ig.inter)
		return fn, ig.pkg

	} else {
		fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{Type: ast.NewIdent("string"), Names: []*ast.Ident{{Name: "name"}}})
		fn.Type.Results.List = append(fn.Type.Results.List, &ast.Field{Type: ast.NewIdent(ig.inter), Names: []*ast.Ident{{Name: "obj"}}})

		swcase := ast.SwitchStmt{
			Tag: &ast.ParenExpr{
				X: ast.NewIdent("name"),
			},
			Body: &ast.BlockStmt{},
		}
		for _, i := range ig.elts {
			kvp := i.(*ast.KeyValueExpr)

			// spew.Dump("kvp", kvp.Key, kvp.Value)
			caseClause := ast.CaseClause{}
			caseClause.List = append(caseClause.List, kvp.Key)
			implementationType, ok := kvp.Value.(*ast.CompositeLit).Type.(*ast.Ident)
			pkgname := ig.pkg
			var stName string
			if !ok {
				pkgname = kvp.Value.(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
				stName = kvp.Value.(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
			} else {
				stName = implementationType.Name
			}

			// handle package logic later
			s, _ := GetStructOrInterface(stName, pkgname)

			_, i := GetStructOrInterface(ig.inter, ig.pkg)
			if s == nil {
				log.Panic("struct pkg " + pkgname + " " + stName + " doesnt found for bidning with " + ig.inter + " method")
			}
			if len(i.Methods) > 0 {
				n := i.Methods[0].Name
				m := GetStructMethod(s.Name, n, s.File.Package)
				if m == nil {
					log.Panic("struct " + s.Name + " doesnt have " + n + " method")
				}
			}
			specificFunction := &ast.FuncDecl{
				Name: ast.NewIdent("New" + stName),
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: &ast.FieldList{},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			}
			CreateStruct(s, specificFunction, pkgs[ig.pkg], nil)
			ret := specificFunction.Type.Results.List[0]
			ret.Type = ast.NewIdent(ig.inter)

			// assignStmt := &ast.AssignStmt{
			// 	Tok: token.ASSIGN,
			// 	Rhs: []ast.Expr{
			// 		&ast.CallExpr{
			// 			Fun: ast.NewIdent("New" + stName),
			// 		},
			// 	},
			// 	Lhs: []ast.Expr{
			// 		ast.NewIdent("obj"),
			// 	},
			// }
			// if len(specificFunction.Type.Results.List) > 1 {
			// 	assignStmt.Lhs = append(assignStmt.Lhs, ast.NewIdent("err"))
			// 	if len(fn.Type.Results.List) == 1 {
			// 		fn.Type.Results.List = append(fn.Type.Results.List, &ast.Field{Type: ast.NewIdent("error"), Names: []*ast.Ident{{Name: "err"}}})
			// 	}
			// }
			// caseClause.Body = append(caseClause.Body, assignStmt)
			caseClause.Body = specificFunction.Body.List[0 : len(specificFunction.Body.List)-1]
			caseClause.Body[len(caseClause.Body)-1].(*ast.AssignStmt).Lhs[0].(*ast.Ident).Name = "obj"
			swcase.Body.List = append(swcase.Body.List, &caseClause)
		}
		fn.Body.List = append(fn.Body.List, &swcase)
		fn.Body.List = append(fn.Body.List, &ast.ReturnStmt{})
		// need to return here new parent function which resolves all this.
		return fn, ig.pkg
	}
}

/**
 ****************************
 * Struct generator
 ****************************
 */

type StructGenerator struct {
	pkg string
	src string
}

func (sg *StructGenerator) Name() string {
	return sg.src
}

func (sg *StructGenerator) ParseDeclaration(n ast.Node, pkg string) {
	sg.pkg = pkg
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		panic("not given call expr to struct builder")
	}
	if ok {
		switch callExpr.Args[0].(type) {
		case *ast.CompositeLit:
			switch callExpr.Args[0].(*ast.CompositeLit).Type.(type) {
			case *ast.Ident:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
				sg.pkg = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			}
		case *ast.BasicLit:
			sg.src = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
		case *ast.Ident:
			sg.src = callExpr.Args[0].(*ast.Ident).Name
		}
	}
}

func (sg *StructGenerator) Generate() (*ast.FuncDecl, string) {
	fn := &ast.FuncDecl{
		Name: ast.NewIdent("New" + sg.src),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	s, _ := GetStructOrInterface(sg.src, sg.pkg)
	CreateStruct(s, fn, pkgs[sg.pkg], nil)
	// os.Exit(0)
	return fn, sg.pkg
}

func CreateStruct(st *Struct, fn *ast.FuncDecl, pkg *Package, prevStruct *Struct) {
	for _, fnstmt := range fn.Body.List {
		s, ok := fnstmt.(*ast.AssignStmt)
		if ok && getName(s.Lhs[0]) == getVarName(st.Name) {
			return
		}
	}
	//add return paramter
	if prevStruct == nil {
		fn.Type.Results.List = append(fn.Type.Results.List, &ast.Field{
			Type: &ast.StarExpr{
				X: st.Ast.Name,
			},
			Names: []*ast.Ident{
				ast.NewIdent(getVarName(st.Name)),
			},
		})

	}

	var fields []ast.Expr
	for _, field := range st.Ast.Type.(*ast.StructType).Fields.List {
		var value ast.Expr
		n := getName(field.Type)
		if len(field.Names) > 0 {
			n = field.Names[0].String()
		}
		value = ast.NewIdent(getVarName(n))
		co, sharedVariableExists := pkg.Shared[getName(field.Type)]
		pkgname := getPackageName(field.Type)
		if pkgname != "" {
			pkg.addImport(st.File.Imports[pkgname], "")
		}

		if pkgname == "" {
			pkgname = st.File.Package
		}

		method := GetMethod("Get"+getName(field.Type), pkgname)
		if method == nil {
			method = GetMethod("New"+getName(field.Type), pkgname)
			if method != nil {
				for _, sm := range pkg.singleton.SingletonMethods {
					if sm.Name == "Get"+getName(field.Type) {
						method = sm
						break
					}
				}
			}
		}
		// check if method exists, singleton method,or shared method
		if sharedVariableExists {
			value = ast.NewIdent(co)
		} else if method != nil {

			fncall := functionCall(method, true, method.Name[0:3], false)
			if pkgname != pkg.name {
				c := fncall.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Fun
				fncall.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Fun = &ast.SelectorExpr{
					X:   ast.NewIdent(pkgname),
					Sel: c.(*ast.Ident),
				}
			}
			for i, v := range fncall.(*ast.AssignStmt).Rhs[0].(*ast.CallExpr).Args {
				param := &ast.Field{
					Names: []*ast.Ident{v.(*ast.Ident)},
					Type:  method.Ast.Type.Params.List[i].Type,
				}
				fn.Type.Params.List = append(fn.Type.Params.List, param)
			}
			addStatement(fn, fncall.(*ast.AssignStmt))

			for _, v := range method.Ast.Type.Results.List {

				if strings.EqualFold(getName(v.Type), getName(field.Type)) {
					_, variableStoresPointer := v.Type.(*ast.StarExpr)
					_, fieldAsksPointer := field.Type.(*ast.StarExpr)

					if variableStoresPointer && !fieldAsksPointer {
						value = &ast.StarExpr{
							X: value,
						}
					}
					if !variableStoresPointer && fieldAsksPointer {
						value = &ast.UnaryExpr{
							Op: token.AND,
							X:  value,
						}
					}
				}
				if getName(v.Type) == "error" {
					fn.Type.Results.List = append(fn.Type.Results.List, &ast.Field{
						Type: ast.NewIdent("error"),
						Names: []*ast.Ident{
							ast.NewIdent("err"),
						},
					})

					smt := &ast.IfStmt{
						Cond: &ast.BinaryExpr{
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
							X:  ast.NewIdent("err"),
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{},
							},
						},
					}
					addStatement(fn, smt)
				}
			}
		} else {
			s, i := GetStructOrInterface(getName(field.Type), pkgname)
			if s != nil {
				value = ast.NewIdent(getVarName(getName(field.Type)))
				CreateStruct(s, fn, pkg, st)

				_, fieldAsksPointer := field.Type.(*ast.StarExpr)
				_, valueIsPointer := value.(*ast.StarExpr)
				if fieldAsksPointer && !valueIsPointer {
					value = &ast.UnaryExpr{
						X:  value,
						Op: token.AND,
					}
				}

			} else if i != nil {
				value = ast.NewIdent(getVarName(getName(field.Type)))
				// if interface is type of one of field resolve it to given src.
				generator := GetGenerator(i.Name)
				if generator == nil {
					panic("Interface to Implemenation not found forr " + i.Name)
				}
				is, _ := GetStructOrInterface(generator.(*InterfaceGenerator).src, pkgname)
				if is != nil {
					CreateStruct(is, fn, pkg, st)
				}
				value = &ast.UnaryExpr{
					X:  ast.NewIdent(getVarName(is.Name)),
					Op: token.AND,
				}

			} else {
				// unknown types needs to be accept as parameter.
				for _, v := range fn.Type.Params.List {
					if v.Names[0].Name == getName(value) {
						value = ast.NewIdent(getVarName(st.Ast.Name.Name) + strings.Title(v.Names[0].Name))
					}
				}
				param := &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(getName(value))},
					Type:  field.Type,
				}
				fn.Type.Params.List = append(fn.Type.Params.List, param)
			}

		}

		name := getName(field.Type)
		if len(field.Names) > 0 {
			name = field.Names[0].String()
		}

		field := &ast.KeyValueExpr{
			Key:   ast.NewIdent(name),
			Value: value,
		}

		fields = append(fields, field)

	}
	tok := token.DEFINE
	if prevStruct == nil {
		tok = token.ASSIGN
	}
	var t ast.Expr
	t = st.Ast.Name
	if st.File.Package != pkg.name {
		if prevStruct != nil {
			imp := prevStruct.File.Imports[st.File.Package]
			if imp != "" {
				pkg.addImport(imp, "")

			}
		} else {
			pkg.addImport(st.File.Path, "")
		}
		t = &ast.SelectorExpr{
			X:   ast.NewIdent(st.File.Package),
			Sel: st.Ast.Name,
		}

	}
	stmt := &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(getVarName(st.Ast.Name.String())),
		},
		Tok: tok,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Incomplete: false,
				Type:       t,
				Elts:       fields,
			},
		},
	}
	if prevStruct == nil {
		stmt.Rhs[0] = &ast.UnaryExpr{
			Op: token.AND,
			X:  stmt.Rhs[0],
		}
	}
	// add body statements
	addStatement(fn, stmt)
	if prevStruct == nil {
		addStatement(fn, &ast.ReturnStmt{})

	}
}

/**
 ****************************
 * Shared Variable generator
 ****************************
 */
type SharedGenerator struct {
	pkg  string
	src  string
	code string
	call bool
}

func (shg *SharedGenerator) Name() string {
	return shg.src
}

func (sg *SharedGenerator) ParseDeclaration(n ast.Node, pkg string) {
	sg.pkg = pkg
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		panic("not given call expr to shared builder")
	}
	if ok {
		switch callExpr.Args[0].(type) {
		case *ast.CompositeLit:
			switch callExpr.Args[0].(*ast.CompositeLit).Type.(type) {
			case *ast.Ident:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
				sg.pkg = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			}
		case *ast.BasicLit:
			sg.src = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
		case *ast.Ident:
			sg.src = callExpr.Args[0].(*ast.Ident).Name
		}

		fset := token.NewFileSet()
		buf := new(strings.Builder)
		printer.Fprint(buf, fset, callExpr.Args[1])
		sg.code = buf.String()
		_, sg.call = callExpr.Args[1].(*ast.CallExpr)
	}
}

func (sg *SharedGenerator) Generate() (*ast.FuncDecl, string) {
	pk := pkgs[sg.pkg]

	if sg.call {
		code := sg.code
		if strings.Contains(code, ".") {
			code = strings.Split(code, ".")[1]
		}
		pk.SharedVars = append(pk.SharedVars, &ast.ValueSpec{
			Names: []*ast.Ident{ast.NewIdent(getVarName(sg.src))},
			Values: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent(strings.ReplaceAll(code, "()", "")),
				},
			},
		})
		pk.Shared[sg.src] = getVarName(sg.src)

	} else {
		pk.Shared[sg.src] = sg.code

	}
	return nil, sg.pkg
}

/**
 ****************************
 * Singleton Variable generator
 ****************************
 */
type SingletonGenerator struct {
	pkg  string
	src  string
	code string
	call bool
}

func (shg *SingletonGenerator) Name() string {
	return shg.src
}

func (sg *SingletonGenerator) ParseDeclaration(n ast.Node, pkg string) {
	sg.pkg = pkg
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		panic("not given call expr to singleton builder")
	}
	if ok {
		switch callExpr.Args[0].(type) {
		case *ast.CompositeLit:
			switch callExpr.Args[0].(*ast.CompositeLit).Type.(type) {
			case *ast.Ident:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				sg.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
				sg.pkg = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
			}
		case *ast.BasicLit:
			sg.src = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
		case *ast.Ident:
			sg.src = callExpr.Args[0].(*ast.Ident).Name
		}

		fset := token.NewFileSet()
		buf := new(strings.Builder)
		printer.Fprint(buf, fset, callExpr.Args[1])
		sg.code = buf.String()
		_, sg.call = callExpr.Args[1].(*ast.CallExpr)

	}
}

func (sg *SingletonGenerator) Generate() (*ast.FuncDecl, string) {
	pkgs[sg.pkg].singleton.Add(sg.src, pkgs[sg.pkg])

	return nil, sg.pkg
}

func addStatement(fn *ast.FuncDecl, stmt ast.Stmt) {
	fn.Body.List = append(fn.Body.List, stmt)
}
