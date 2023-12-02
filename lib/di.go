package lib

import (
	"go/ast"
	"os"
	"strings"
)

type visitor struct {
	v   int
	pkg string
}

// this function extracts data from di.go and tracks list of functions to generate at declare level.
func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	callExpr, ok := n.(*ast.CallExpr)
	if ok {
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if ok && sel.X.(*ast.Ident).Name == "di" {
			// fmt.Printf("%#v\n\n", callExpr.Args[0])
			g := CreateGenerator(sel.Sel.Name)
			if g != nil {
				g.ParseDeclaration(n, v.pkg)
				Generators = append(Generators, g)
			}

		}
	}
	v.v = v.v + 1
	return v
}

func Run(dir string, mod string) {

	pkgs = make(map[string]*Package)

	// we need to understand why we need mod as parameter
	// if mod == "" {
	// 	mod = getCurrentModuleName()
	// }

	diFiles, otherFiles := traversDir(dir)

	for _, file := range otherFiles {
		if file == nil {
			continue
		}
		pkg, ok := pkgs[file.Package]

		if !ok {
			pkg = NewPackage(file)
		}
		pkg.AddParsedFile(file)
		pkgs[file.Package] = pkg
	}

	// parse di.go and extract dependancy injection information.
	vs := visitor{}
	for _, v := range diFiles {
		vs.pkg = v.Name.Name
		ast.Walk(vs, v)
	}

	for _, v := range Generators {
		fn, pkg := v.Generate()

		if fn != nil {
			if os.Getenv("ENV") != "" && fn != nil {
				envSpecificGeneratorName := strings.Title(os.Getenv("ENV")) + v.Name()
				envGenerator := GetGenerator(envSpecificGeneratorName)
				if envGenerator != nil && envGenerator.Name() != v.Name() {
					continue
				}
			}
			pkgs[pkg].Funcs = append(pkgs[pkg].Funcs, fn)
		}
	}

	for _, pkg := range pkgs {
		pkg.cleanup()
		pkg.generateFile()

	}

}

// traverse dir traverses given directory recursively and parse go files
func traversDir(dir string) (diFiles []*ast.File, otherFiles []*CodeFile) {
	files, _ := os.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {

			if !ContainsStr([]string{"vendor", ".git"}, file.Name()) {
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
