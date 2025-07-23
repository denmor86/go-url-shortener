// Package exitcheck предоставляет функционал статического анализатора с проверкой запрета использования прямого вызова os.Exit в функции main пакета main.
package exitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer -  статический анализатор проверяющий вызов os.Exit() в функции main пакета main
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit() calls in main function of main package",
	Run:  run,
}

// run -  метод запуска анализа наличия вызова os.Exit() в функции main пакета main
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// пропускаем все не main пакеты
		if file.Name.Name != "main" {
			continue
		}

		// проверяем что был импортирован пакет os
		osImported := false
		for _, imp := range file.Imports {
			if imp.Path.Value == `"os"` {
				osImported = true
				break
			}
		}

		if !osImported {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				continue
			}

			ast.Inspect(fn, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit call in main function of main package")
				}
				return true
			})
		}
	}

	return nil, nil
}
