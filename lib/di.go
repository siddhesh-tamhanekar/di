package lib

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
)

var diassignments map[string]*Di

var pkgs map[string]*Package

type Function struct {
	name string
	args []string
	ret  []string
	code string
}

type Package struct {
	name        string
	path        string
	Structs     []*Struct
	Interfaces  []*Interface
	Methods     []*Method
	Imports     []string
	Fns         map[string]Function
	Vars        []string
	Shared      map[string]string
	ImportSpecs []ast.Spec
	SharedVars  []ast.Spec
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
	fmt.Println("adding improt", imp, "into", pkg.name)
	importStmt := &ast.ImportSpec{Path: &ast.BasicLit{Value: "\"" + imp + "\""}}
	if name != "" {
		importStmt.Name = ast.NewIdent(name)
	}
	pkg.ImportSpecs = append(pkg.ImportSpecs, importStmt)

}

type Di struct {
	method string
	inter  string
	src    string
	code   string
	pkg    string
	call   bool
	env    string
}

type visitor struct {
	v             int
	diassignments map[string]*Di
	pkg           string
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	callExpr, ok := n.(*ast.CallExpr)
	if ok {
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if ok && sel.X.(*ast.Ident).Name == "di" {
			// fmt.Printf("%#v\n\n", callExpr.Args[0])
			var d Di
			d.method = sel.Sel.Name
			d.pkg = v.pkg

			switch callExpr.Args[0].(type) {
			case *ast.CompositeLit:
				switch callExpr.Args[0].(*ast.CompositeLit).Type.(type) {
				case *ast.Ident:
					d.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.Ident).Name
					break
				case *ast.SelectorExpr:
					d.src = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
					d.pkg = callExpr.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
					break
				}
				break
			case *ast.BasicLit:
				d.src = strings.Trim(callExpr.Args[0].(*ast.BasicLit).Value, "\"")
				break
			case *ast.Ident:
				d.src = callExpr.Args[0].(*ast.Ident).Name
				break
			}

			if d.method == "Share" {
				// fmt.Printf("%#v\n", callExpr.Args[1])
				fset := token.NewFileSet()
				buf := new(strings.Builder)
				printer.Fprint(buf, fset, callExpr.Args[1])
				d.code = buf.String()
				_, d.call = callExpr.Args[1].(*ast.CallExpr)
			}

			if d.method == "Bind" || d.method == "BindEnv" {
				d.inter = d.src
				switch callExpr.Args[1].(type) {
				case *ast.CompositeLit:
					switch callExpr.Args[1].(*ast.CompositeLit).Type.(type) {
					case *ast.Ident:
						d.src = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.Ident).Name
						break
					case *ast.SelectorExpr:
						d.src = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
						d.pkg = callExpr.Args[1].(*ast.CompositeLit).Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
						break
					}
				}

				if d.method == "BindEnv" {
					env, ok := callExpr.Args[2].(*ast.BasicLit)
					if ok {
						d.env = strings.Trim(env.Value, "\"")
					}

				}
			}

			// fmt.Printf("%#v\n\n", d)

			if d.inter != "" {
				k := strings.Title(d.env) + d.inter
				v.diassignments[k] = &d
			} else {
				v.diassignments[d.src] = &d
			}
		}
	}

	// fmt.Printf("%s%#v\n", strings.Repeat("*", int(v)), n)
	v.v = v.v + 1
	return v
}

func Run(dir string, mod string) {
	vs := visitor{
		diassignments: make(map[string]*Di),
	}
	diassignments = vs.diassignments
	pkgs = make(map[string]*Package)
	if mod == "" {
		mod = getCurrentModuleName()
	}

	diFiles, otherFiles := traversDir(dir)
	for _, v := range diFiles {
		vs.pkg = v.Name.Name
		ast.Walk(vs, v)
	}
	// spew.Dump(diassignments)
	for _, file := range otherFiles {
		if file == nil {
			continue
		}
		pkg, ok := pkgs[file.Package]

		if ok == false {
			pkg = &Package{
				name:        file.Package,
				path:        file.Path,
				Fns:         make(map[string]Function),
				Shared:      make(map[string]string),
				Vars:        make([]string, 0),
				ImportSpecs: make([]ast.Spec, 0),
			}
		}
		pkg.Structs = append(pkg.Structs, file.Structs...)
		pkg.Methods = append(pkg.Methods, file.Methods...)
		pkg.Interfaces = append(pkg.Interfaces, file.Interfaces...)
		pkgs[file.Package] = pkg
	}

	for _, v := range vs.diassignments {
		generateCode(v)
	}

	for _, pkg := range pkgs {
		cleanup(pkg)
		generateDiGenFile(pkg, mod)
		generateFile(pkg)
	}

}
func cleanup(pkg *Package) {
	fp := pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_gen.go"
	os.Remove(fp)
	fp = pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_generated.go"
	os.Remove(fp)

}
func generateDiGenFile(pkg *Package, mod string) {
	pkgbytes := []byte(fmt.Sprintf("package %s\n", pkg.name))

	// fmt.Println("pkgs", pkg.Fns)
	b := pkgbytes
	b = append(b, getImports(pkg, mod)...)
	b = append(b, getSharedVars(pkg.Vars)...)
	for _, v := range pkg.Fns {
		b = append(b, []byte(v.code)...)
	}

	if string(pkgbytes) == string(b) {
		// fmt.Println("NO NEED TO GENERATE CODE FOR" + pkg.path)
		return
	}

	// fmt.Println(string(b))
	b = append([]byte(`
	// Code generated by DI library. DO NOT EDIT.
	// To generate file use <path_to_di>/di --path= --module=
	`), b...)
	b, err := format.Source(b)
	if err != nil {
		fmt.Println("format error", err)
		return
	}

	fp := pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_gen.go"
	os.Remove(fp)
	WriteFile(fp, b)
}

func generateFile(pkg *Package) {
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
	fp := pkg.path[:strings.LastIndex(pkg.path, "/")] + "/di_generated.go"
	// if len(file.Decls) == 0 {
	// 	fmt.Println("No Need to generate file", fp)
	// 	return
	// }

	addGeneratedComments(&file)
	buf := new(strings.Builder)
	printer.Fprint(buf, fset, &file)

	fstr := strings.ReplaceAll(buf.String(), "EOF\n", "")
	b := []byte(fstr)
	emptyfile := fmt.Sprintf("package %s\n\n// Code generated by DI library. DO NOT EDIT.\n// To generate file use <path_to_di>/di --path= --module=\n", pkg.name)

	for _, v := range pkg.Fns {
		b = append(b, []byte(v.code)...)
	}
	if emptyfile == string(b) {
		fmt.Println("No Need to generate file v2 ", fp)
		return
	}
	b, err := format.Source(b)
	if err != nil {
		fmt.Println("format error", err)
		return
	}

	os.Remove(fp)
	WriteFile(fp, b)

}

func addGeneratedComments(file *ast.File) {

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

func getSharedVars(vars []string) (b []byte) {
	if len(vars) > 0 {
		varstr := "var("
		for _, v := range vars {
			varstr += v + "\n"
		}
		varstr += ")"
		b = append(b, []byte(varstr)...)
	}
	return
}

func getImports(pkg *Package, mod string) []byte {
	var imports []string
	for _, im := range removeDuplicateStr(pkg.Imports) {
		if im == "" {
			continue
		}
		// fmt.Println("Import", im)
		imp := im
		if strings.Contains(im, ".go") {
			imp = im[:strings.LastIndex(im, "/")]
			imp = strings.ReplaceAll(imp, "./", mod+"/")

		}
		imports = append(imports, "import \""+imp+"\"\n")
	}
	return []byte(strings.Join(removeDuplicateStr(imports), ""))

}

func generateCode(v *Di) {
	var fn Function
	var args, codes string
	if (v.method == "BindEnv") && os.Getenv("ENV") != v.env {
		// fmt.Println("returning", v)
		return
	}
	fn.name = "New" + v.src
	pk, ok := pkgs[v.pkg]

	if ok == false {
		pk = &Package{
			name:        v.pkg,
			Fns:         make(map[string]Function),
			Shared:      make(map[string]string),
			Vars:        make([]string, 0),
			ImportSpecs: make([]ast.Spec, 0),
		}
	}

	if v.method == "Share" {
		if v.call {
			code := v.code
			if strings.Contains(code, ".") {
				code = strings.Split(code, ".")[1]
			}
			pk.Vars = append(pk.Vars, getVarName(v.src)+"="+code)
			pk.Shared[v.src] = getVarName(v.src)
			pk.SharedVars = append(pk.SharedVars, &ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(getVarName(v.src))},
				Values: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent(strings.ReplaceAll(code, "()", "")),
					},
				},
			})

		} else {
			pk.Shared[v.src] = v.code
		}
	}

	if v.method == "Build" || v.method == "Bind" || v.method == "BindEnv" {
		s, _ := getStructOrInterface(v.src, v.pkg)
		rootPointer := ""
		ret := v.src
		if v.method == "Bind" || v.method == "BindEnv" {
			ret = v.inter
			fn.name = "New" + v.inter
			_, i := getStructOrInterface(v.inter, v.pkg)
			if len(i.Methods) > 0 {
				n := i.Methods[0].Name
				m := getStructMethod(s.Name, n, v.pkg)
				if m == nil {
					log.Panic("struct " + s.Name + " doesnt have " + n + " method")
				}
				if m.Reciever.T[0:1] == "*" {
					rootPointer = "&"
				}
			}
		}
		co, ar, im, returns := generateFunctionBody(s, v.pkg, true, rootPointer)
		// fmt.Printf("returned %#v\n", returns)
		sort.Strings(ar)
		sort.Strings(returns)
		fn.args = ar
		args = strings.Join(ar, ", ")
		codes = strings.Join(co, "\n")
		// pkgs[v.pkg] = append(pkgs[v.pkg].Imports, im...)
		pk.Imports = append(pk.Imports, im...)
		rets := append([]string{getVarName(v.src) + " " + ret}, returns...)

		// variable is introduced for special case where share variable is wrapped in function and returned.
		_, ok := pk.Shared[v.src]
		retvar := ""
		if ok {
			rets = append([]string{ret}, returns...)
			retvar = getVarName(s.Name)
		}
		fn.ret = append(fn.ret, rets...)
		fn.code = generateFunction(fn.name, codes, args, "("+strings.Join(fn.ret, ", ")+")", retvar)

		pk.Fns[fn.name] = fn
	}
	pkgs[v.pkg] = pk

}

func generateFunction(name, body, args string, ret string, retvar string) string {
	fnTemplate := `
	func {NAME}({ARGS}) {RETURNS} {
		{CODE}
		return {RET}
	}
	`

	code := strings.ReplaceAll(fnTemplate, "{NAME}", name)
	code = strings.ReplaceAll(code, "{ARGS}", args)
	code = strings.ReplaceAll(code, "{CODE}", body)
	code = strings.ReplaceAll(code, "{RETURNS}", ret)
	code = strings.ReplaceAll(code, "{RET}", retvar)
	return code
}

func generateFunctionBody(s *Struct, pkg string, root bool, returnPointer string) (args []string, code []string, imports []string, returns []string) {
	c := ""
	if s == nil {
		return
	}
	pack := pkgs[pkg]
	co, sharedVariableExists := pkgs[pkg].Shared[s.Name]
	met := getMethod("New"+s.Name, s.File.Package)

	if sharedVariableExists {
		if getVarName(s.Name) != co {
			if root {
				c = getVarName(s.Name) + " = " + returnPointer + co
			} else {
				c = getVarName(s.Name) + " := " + co
			}
		}
	} else if met != nil {
		var vars, ar []string
		for _, v := range met.Results {
			if v.T == "error" {
				vars = append(vars, "err")
				continue
			}
			varName := getVarName(v.T)
			if v.T[0:1] == "*" {
				varName = "_" + varName
			}
			vars = append(vars, varName)
		}
		for _, v := range met.Params {
			args = append(args, v.Name+" "+v.T)

		}
		for _, arg := range args {
			ar = append(ar, strings.Split(arg, " ")[0])
		}
		// if vars[0][0:1] != "_" {
		if root {
			c = strings.Join(vars, ", ") + " = " + met.Name + "(" + strings.Join(ar, ", ") + ")"
		} else {
			c = strings.Join(vars, ", ") + " := " + met.Name + "(" + strings.Join(ar, ", ") + ")"
		}
		// }
		retPrefix := ""
		if returnPointer != "" {
			// fmt.Printf("%#v, %s | %s\n", met.Name, c, returnPointer)
			if root {
				c = strings.ReplaceAll(c, "=", ":=")
			}
			c = "ret" + c
			retPrefix = "ret"
		}
		if vars[0][0:1] == "_" {
			ptrPrefix := ""
			if returnPointer != "" {
				ptrPrefix = ""
			}
			if root {
				c += "\n " + vars[0][1:] + "=" + ptrPrefix + retPrefix + vars[0]
			} else {
				c += "\n " + vars[0][1:] + ":=*" + ptrPrefix + retPrefix + vars[0]
			}
		} else {
			if returnPointer != "" {
				if root {
					c += "\n " + vars[0] + "=" + returnPointer + retPrefix + vars[0]
				} else {
					c += "\n " + vars[0] + ":=" + returnPointer + retPrefix + vars[0]
				}

			}

		}
		if ContainsStr(vars, "err") {
			c += "\n if err!=nil{ return }"
			returns = append(returns, "err error")
		}
	} else if s.File.Package != pkg {
		imports = append(imports, s.File.Path)
		pack.addImport(s.File.Path, "")
		generateCode(&Di{
			method: "Build",
			src:    s.Name,
			pkg:    s.File.Package,
		})
		args = append(args, pkgs[s.File.Package].Fns["New"+s.Name].args...)
		var ar []string
		for _, arg := range args {
			ar = append(ar, strings.Split(arg, " ")[0])
		}
		if root {
			c = getVarName(s.Name) + " = " + returnPointer + s.File.Package + ".New" + s.Name + "(" + strings.Join(ar, ", ") + ")"

		} else {
			c = getVarName(s.Name) + " := " + s.File.Package + ".New" + s.Name + "(" + strings.Join(ar, ", ") + ")"
		}
	} else {
		if root {
			c = getVarName(s.Name) + "=" + returnPointer + s.Name + "{\n"

		} else {
			c = getVarName(s.Name) + ":=" + s.Name + "{\n"
		}
		for _, f := range s.Fields {
			p := s.File.Package
			t := f.Type
			isPointer := false
			if strings.Contains(f.Type, ".") {
				p = strings.Split(f.Type, ".")[0]
				if p[0:1] == "*" {
					isPointer = true
					p = p[1:]
				}
				t = strings.Split(f.Type, ".")[1]
			}
			types := []string{
				"bool",
				"string",
				"int", "int8", "int16", "int32", "int64",
				"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
				"byte",
				"float32", "float64",
				"complex64", "complex128",
				"any",
				"interface{}",
			}
			for _, typ := range types {
				types = append(types, "*"+typ)
			}
			if ContainsStr(types, f.Type) {
				args = append(args, getVarName(f.Name)+" "+f.Type)
				c += f.Name + ":" + getVarName(f.Name) + ","
			} else {
				if t[0:1] == "*" {
					isPointer = true
					t = t[1:]
				}
				// fmt.Println(t, p)
				s1, i := getStructOrInterface(t, p)
				if s1 == nil && i == nil {

					args = append(args, getVarName(f.Name)+" "+f.Type)
					c += f.Name + ":" + getVarName(f.Name) + ","
					// spew.Dump(s.File)
					imports = append(imports, s.File.Imports[p])
					pack.addImport(s.File.Imports[p], p)
					// fmt.Println(p, " not found")
				} else {
					if s1 == nil && i != nil {
						d := strings.Title(os.Getenv("ENV")) + i.Name
						dia, ok := diassignments[d]
						if ok == false {
							dia = diassignments[i.Name]
						}
						if dia == nil {
							fmt.Println("Interface to Implementation not found for", i.Name)
							os.Exit(1)
						}
						s1, _ = getStructOrInterface(dia.src, dia.pkg)
						if s1 == nil && i != nil {
							fmt.Println(dia.src + " not found")
							os.Exit(1)
						}
					}
					co, ar, im, ret := generateFunctionBody(s1, pkg, false, "")
					if len(ret) > 0 {
						returns = append(returns, ret...)
					}

					prefix := ""
					if i != nil {
						t = s1.Name
						if len(i.Methods) > 0 {

							n := i.Methods[0].Name
							m := getStructMethod(s1.Name, n, pkg)
							if m == nil {
								fmt.Println("struct " + s.Name + " doesnt have " + n + " method")
								return
							}
							if m.Reciever.T[0:1] == "*" {
								prefix = "&"
							}
						}

					}
					if isPointer {
						prefix = "&"
					}
					code = append(code, co...)
					args = append(args, ar...)
					imports = append(imports, im...)
					c += f.Name + ":" + prefix + getVarName(t) + ",\n"
				}

			}
		}
		c += "\n}"
	}

	code = append(code, c)
	return removeDuplicateStr(code), removeDuplicateStr(args), imports, returns
}

func getStructOrInterface(s string, p string) (*Struct, *Interface) {
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

func getStructMethod(s string, method string, packageName string) *Method {
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

func getMethod(method string, packageName string) *Method {
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

// traverse dir traverses given directory recursively and parse go files
func traversDir(dir string) (diFiles []*ast.File, otherFiles []*CodeFile) {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {

			if ContainsStr([]string{"vendor", ".git"}, file.Name()) == false {
				dif, otf := traversDir(dir + "/" + file.Name())
				diFiles = append(diFiles, dif...)
				otherFiles = append(otherFiles, otf...)
			} else {
				logMsg("[Parser] Skipped " + dir + "/" + file.Name())

			}
		} else {
			switch file.Name() {
			case "di_gen.go":
			case "di_generated.go":
			case "di.go":
				diFiles = append(diFiles, parseFile(dir+"/"+file.Name()))
				logMsg("[Parser] Parsed " + dir + "/" + file.Name())
			default:
				if strings.LastIndex(file.Name(), ".go") > 0 {
					otherFiles = append(otherFiles, parseGoFile(dir+"/"+file.Name()))
					logMsg("[Parser] Parsed " + dir + "/" + file.Name())
				}
			}

		}
	}
	return
}

func getModule() string {
	mod := ""
	if mod == "" {
		mod = getCurrentModuleName()
	}
	return mod
}
