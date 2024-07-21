package checkers

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// NoOsExitInMainAnalyzer кастомный чекер
var NoOsExitInMainAnalyzer = &analysis.Analyzer{
	Name: "noOsExitInMain",
	Doc:  "Checks if os.Exit is called directly in main function",
	Run:  Run,
}

// Run проверяет, вызывается ли os.Exit напрямую в функции main
func Run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if ident.Sel.Name == "Exit" && ident.X.(*ast.Ident).Name == "os" {
				pass.Reportf(callExpr.Pos(), "direct call to os.Exit found")
			}

			return true
		})
	}

	return nil, nil
}
