// Package main содержит реализацию кастомного статического анализатора OsExitAnalyzer.
//
// # Назначение анализатора
//
// Анализатор предназначен для контроля чистоты завершения приложения. Он запрещает
// совершать прямой вызов функции [os.Exit] внутри функции main() пакета main.
//
// Использование os.Exit в главной функции обходит механизмы отложенного вызова
// функций (defer), что приводит к утечке ресурсов, незакрытым сетевым соединениям,
// дескрипторам файлов и незавершенной записи буферов логирования. Вместо этого
// рекомендуется возвращать ошибку из бизнес-логики и обрабатывать её централизованно.
//
// # Исключения из проверки
//
// Анализатор игнорирует:
//   - Любые другие функции, кроме main() (например, инициализационные или вспомогательные).
//   - Любые другие пакеты, кроме пакета main.
//   - Тестовые файлы Go (файлы с суффиксом _test.go), так как использование os.Exit
//     в функциях вида TestMain(m *testing.M) является стандартной практикой для Go.
//
// # Надежность и анализ типов
//
// Для минимизации ложных срабатываний анализатор не полагается на сырое текстовое
// сравнение имен (string matching). Вместо этого он использует информацию о типах 
// из [golang.org/x/tools/go/analysis.Pass.TypesInfo]:
//   - Защита от алиасов: код корректно определит вызов os.Exit, даже если пакет
//     был импортирован под другим именем (например, `import o "os"` и вызов `o.Exit(1)`).
//   - Защита от маскирования переменных: если в коде объявлена локальная переменная
//     или структура с именем `os` (например, `os := &myStruct{}`), анализатор поймет,
//     что это не системный пакет "os", и пропустит её.
package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer запрещает использовать прямой вызов os.Exit в функции main пакета main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks for direct os.Exit calls in main function of main package",
	Run:  runOsExit,
}

func runOsExit(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что анализируется пакет main (входная точка приложения)
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Пропускаем тестовые файлы пакета main, так как в тестах вызов os.Exit может быть легитимен
		if strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			// Ищем объявления функций
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Проверяем, что функция называется строго main
			if fn.Name.Name != "main" {
				return true
			}

			// Рекурсивно проверяем тело функции main на наличие вызовов os.Exit
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Извлекаем селектор выражения вызова (например, <X>.<Sel> -> os.Exit)
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Проверяем, равен ли вызываемый метод "Exit"
				if sel.Sel.Name != "Exit" {
					return true
				}

				// Используем TypesInfo для надежного определения того, что находится слева от точки
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				// Извлекаем объект типа из карты Uses на основе идентификатора
				obj := pass.TypesInfo.Uses[ident]
				if obj == nil {
					return true
				}

				// Проверяем, является ли объект именем импортированного пакета (types.PkgName)
				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				// Окончательно убеждаемся, что путь импорта этого пакета равен строго "os"
				if pkgName.Imported().Path() == "os" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function is prohibited")
				}

				return true
			})

			// Возвращаем false, чтобы не обходить тело этой функции повторно на верхнем уровне Inspect
			return false
		})
	}

	return nil, nil
}
