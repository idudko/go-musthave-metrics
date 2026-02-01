// Package analyzer provides a static analyzer for Go code that detects:
// - Usage of panic function
// - Calls to log.Fatal or os.Exit outside of main function in main package
package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for usage of panic, log.Fatal and os.Exit outside of main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Объект для встроенной функции panic
	panicObj := types.Universe.Lookup("panic")

	// Кэшируем объекты для log.Fatal и os.Exit
	var logFatalObj, osExitObj types.Object

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			fun := callExpr.Fun

			// Получаем информацию о типе вызываемой функции
			var obj types.Object
			switch v := fun.(type) {
			case *ast.Ident:
				obj = pass.TypesInfo.ObjectOf(v)
			case *ast.SelectorExpr:
				obj = pass.TypesInfo.ObjectOf(v.Sel)
			}

			if obj == nil {
				return true
			}

			// Проверяем на использование panic
			if obj == panicObj {
				pass.Reportf(callExpr.Pos(), "panic should not be used in production code")
				return true
			}

			// Проверяем на использование log.Fatal
			if isLogFatal(obj) {
				if logFatalObj == nil {
					logFatalObj = obj
				}
				// Проверяем, что мы находимся в пакете main
				if pass.Pkg.Name() != "main" {
					pass.Reportf(callExpr.Pos(), "log.Fatal should only be used in main.main function")
					return true
				}

				// Проверяем, что мы находимся в функции main
				if !isInMainFunc(file, pass.Fset, n) {
					pass.Reportf(callExpr.Pos(), "log.Fatal should only be used in main.main function")
					return true
				}
			}

			// Проверяем на использование os.Exit
			if isOsExit(obj) {
				if osExitObj == nil {
					osExitObj = obj
				}
				// Проверяем, что мы находимся в пакете main
				if pass.Pkg.Name() != "main" {
					pass.Reportf(callExpr.Pos(), "os.Exit should only be used in main.main function")
					return true
				}

				// Проверяем, что мы находимся в функции main
				if !isInMainFunc(file, pass.Fset, n) {
					pass.Reportf(callExpr.Pos(), "os.Exit should only be used in main.main function")
					return true
				}
			}

			return true
		})
	}

	return nil, nil
}

// isLogFatal проверяет, является ли объект функцией log.Fatal или log.Fatalf
func isLogFatal(obj types.Object) bool {
	if obj == nil {
		return false
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	pkg := fn.Pkg()
	if pkg == nil || pkg.Path() != "log" {
		return false
	}
	return fn.Name() == "Fatal" || fn.Name() == "Fatalf" || fn.Name() == "Fatalln"
}

// isOsExit проверяет, является ли объект функцией os.Exit
func isOsExit(obj types.Object) bool {
	if obj == nil {
		return false
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	pkg := fn.Pkg()
	if pkg == nil || pkg.Path() != "os" {
		return false
	}
	return fn.Name() == "Exit"
}

// isInMainFunc проверяется, находится ли узел в функции main пакета main
func isInMainFunc(file *ast.File, fset *token.FileSet, node ast.Node) bool {
	path, _ := astutil.PathEnclosingInterval(file, node.Pos(), node.End())
	if path == nil {
		return false
	}
	for _, n := range path {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			return true
		}
	}
	return false
}
