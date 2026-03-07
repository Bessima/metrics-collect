package errcheckanalyzer

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// CriticErrorAnalyzer checks for panic calls and critical exit functions
// (log.Fatal, os.Exit) outside main.main
var CriticErrorAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "check for panic and critical exit functions (log.Fatal, os.Exit) outside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		mainFunc := findMainFunc(pass, file)

		// Обходим AST и проверяем вызовы функций
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				checkCriticalCalls(pass, call, mainFunc)
			}
			return true
		})
	}
	return nil, nil
}

// findMainFunc возвращает объявление функции main в пакете main, иначе nil.
func findMainFunc(pass *analysis.Pass, file *ast.File) *ast.FuncDecl {
	if pass.Pkg.Name() != "main" {
		return nil
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == "main" {
			return fn
		}
	}
	return nil
}

// checkCriticalCalls проверяет вызовы критических функций
func checkCriticalCalls(pass *analysis.Pass, call *ast.CallExpr, mainFunc *ast.FuncDecl) {
	// Проверяем, находимся ли мы в функции main пакета main
	isMainFunc := false
	if mainFunc != nil && mainFunc.Body != nil {
		isMainFunc = mainFunc.Body.Pos() <= call.Pos() && call.End() <= mainFunc.Body.End()
	}

	// Проверяем встроенную функцию panic
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "panic" {
			pass.Reportf(call.Pos(), "panic call detected")
			return
		}
	}

	// Проверяем log.Fatal* и os.Exit
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}

	pkgName, ok := pass.TypesInfo.Uses[x].(*types.PkgName)
	if !ok {
		return
	}

	importPath := pkgName.Imported().Path()
	funcName := sel.Sel.Name

	switch {
	case importPath == "log" && strings.HasPrefix(funcName, "Fatal"):
		if !isMainFunc {
			pass.Reportf(call.Pos(), "log.Fatal call outside main.main")
		}
	case importPath == "os" && funcName == "Exit":
		if !isMainFunc {
			pass.Reportf(call.Pos(), "os.Exit call outside main.main")
		}
	}
}
