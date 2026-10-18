package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer запрещает использовать прямой вызов os.Exit в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks for direct os.Exit calls in main function of main package",
	Run:  runOsExit,
}

func runOsExit(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что анализируется пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			// Ищем объявления функций
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Проверяем, что функция называется main
			if fn.Name.Name != "main" {
				return true
			}

			// Рекурсивно проверяем тело функции main на наличие os.Exit
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				// Проверяем, что вызывается именно os.Exit
				if ident.Name == "map" || ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function is prohibited")
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
