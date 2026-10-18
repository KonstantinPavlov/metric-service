package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

func main() {

	// Карта для конфигурации статических чекеров
	staticCheckers := make(map[string]bool)

	// Включаем ВСЕ анализаторы класса SA (Correctness)
	// А также по одному из классов ST (Style), QF (Quickfixes), S (Code Simplification)
	staticCheckers["ST1000"] = true // Необходимость документации пакета
	staticCheckers["QF1001"] = true // Оптимизация выделения срезов
	staticCheckers["S1000"] = true  // Использование чистых каналов вместо select

	var checkers []*analysis.Analyzer

	// 1. Добавляем стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	checkers = append(checkers,
		asmdecl.Analyzer, assign.Analyzer, atomic.Analyzer, bools.Analyzer,
		buildtag.Analyzer, cgocall.Analyzer, composite.Analyzer, copylock.Analyzer,
		errorsas.Analyzer, httpresponse.Analyzer, loopclosure.Analyzer, lostcancel.Analyzer,
		nilfunc.Analyzer, printf.Analyzer, shadow.Analyzer, shift.Analyzer,
		stdmethods.Analyzer, structtag.Analyzer, tests.Analyzer, unmarshal.Analyzer,
		unreachable.Analyzer, unsafeptr.Analyzer, unusedresult.Analyzer,
	)

	// 2. Добавляем анализаторы из staticcheck.io по фильтру
	for _, v := range staticcheck.Analyzers {
		// Класс SA или явно указанные другие классы
		if (v.Analyzer.Name[:2] == "SA") || staticCheckers[v.Analyzer.Name] {
			checkers = append(checkers, v.Analyzer)
		}
	}

	// 3. Добавляем публичные сторонние анализаторы
	checkers = append(checkers,
		errcheck.Analyzer,  // Проверка необработанных ошибок
		bodyclose.Analyzer, // Проверка закрытия body у HTTP-ответов
	)

	// 4. Добавляем собственный анализатор запрета os.Exit
	checkers = append(checkers, OsExitAnalyzer)

	multichecker.Main(checkers...)
}
