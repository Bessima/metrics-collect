package errcheckanalyzer

import (
	"go/ast"
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
		// Сначала находим функцию main, если она есть в пакете main
		var mainFunc *ast.FuncDecl
		if pass.Pkg.Name() == "main" {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Name.Name == "main" {
						mainFunc = fn
						break
					}
				}
			}
		}

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
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			// Проверяем log.Fatal, log.Fatalf, log.Fatalln
			if x.Name == "log" && strings.HasPrefix(sel.Sel.Name, "Fatal") {
				if !isMainFunc {
					pass.Reportf(call.Pos(), "log.Fatal call outside main.main")
				}
				return
			}

			// Проверяем os.Exit
			if x.Name == "os" && sel.Sel.Name == "Exit" {
				if !isMainFunc {
					pass.Reportf(call.Pos(), "os.Exit call outside main.main")
				}
				return
			}
		}
	}
}
