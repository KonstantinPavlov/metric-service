// Package main представляет собой статический анализатор кода (мультичекер),
// объединяющий стандартные инструменты проверки Go, анализаторы Staticcheck,
// популярные сторонние линтеры и кастомные правила валидации проекта.
//
// # Механизм запуска и использования
//
// Утилита построена на базе пакета [golang.org/x/tools/go/analysis/multichecker], 
// благодаря чему скомпилированный бинарник поддерживает стандартный CLI-интерфейс.
//
// Сборка инструмента:
//
//	go build -o mychecker cmd/staticlint/main.go
//
// Проверка всего проекта (из корневой директории):
//
//	./mychecker ./...
//
// Проверка конкретного пакета с флагом вывода в формате JSON:
//
//	./mychecker -json ./internal/server/...
//
// Получение подробной справки по всем доступным анализаторам и их флагам:
//
//	./mychecker help
//
// # Список и назначение включенных анализаторов
//
// Инструмент автоматически применяет следующие группы проверок:
//
// 1. Стандартные анализаторы (пакет golang.org/x/tools/go/analysis/passes):
//   - asmdecl: Проверяет соответствие между сигнатурами функций Go и файлами ассемблера.
//   - assign: Обнаруживает бесполезные присваивания вида `x = x` или `a[i] = a[i]`.
//   - atomic: Находит некорректные варианты использования функций из пакета sync/atomic (например, неатомарное `x = atomic.AddUint64(&x, 1)`).
//   - bools: Детектирует логические ошибки и избыточные выражения с булевыми операторами.
//   - buildtag: Проверяет правильность форматирования и размещения директив сборки `//go:build` и `// +build`.
//   - cgocall: Ищет нарушения правил передачи указателей в коде Cgo.
//   - composites: Запрещает использование неименованных (unkeyed) полей при инициализации структур из внешних пакетов.
//   - copylock: Находит ошибочную передачу мьютексов (sync.Mutex) или других объектов блокировки по значению.
//   - errorsas: Проверяет, что второй аргумент функции `errors.As` является указателем на интерфейс или тип, реализующий ошибку.
//   - httpresponse: Обнаруживает опасные паттерны работы с HTTP-ответами (например, вызов `defer resp.Body.Close()` до проверки ошибки `err != nil`).
//   - loopclosure: Предотвращает захват переменных цикла внутри вложенных анонимных функций (goroutines или defer).
//   - lostcancel: Требует обязательного вызова функции отмены `cancel()`, возвращаемой контекстами `context.WithCancel` и аналогами.
//   - nilfunc: Предупреждает о бесполезных сравнениях функций с `nil` (забытый вызов функции).
//   - printf: Контролирует строгое соответствие строк форматирования `fmt.Printf` переданным аргументам.
//   - shadow: Подсвечивает непреднамеренное скрытие (shadowing) переменных во вложенных блоках кода.
//   - shift: Находит битовые сдвиги, которые превышают разрядность целевого целочисленного типа.
//   - stdmethods: Проверяет сигнатуры методов, которые по названию похожи на методы стандартных интерфейсов (например, `Write`, `Read`), на предмет опечаток.
//   - structtag: Проверяет синтаксическую валидность тегов структур (JSON, XML и др.).
//   - tests: Находит ошибочные сигнатуры и названия у тестовых функций, бенчмарков и примеров (Examples).
//   - unmarshal: Запрещает передачу значений, не являющихся указателями или интерфейсами, в функции распаковки данных (`json.Unmarshal`).
//   - unreachable: Обнаруживает мертвый и недостижимый код (после `return`, `panic` или бесконечных циклов).
//   - unsafeptr: Ищет некорректные и ломающие логику рантайма преобразования `uintptr` обратно в `unsafe.Pointer`.
//   - unusedresult: Предупреждает о неиспользовании возвращаемых значений чистых функций стандартной библиотеки (например, `strings.Replace`).
//
// 2. Анализаторы статического анализа Staticcheck (staticcheck.io):
//   - Все чекеры класса SA: Проверки на корректность логики, утечки памяти, рантайм-ошибки и проблемы с многопоточностью.
//   - ST1000: Проверяет наличие обязательного документирующего комментария к пакету.
//   - QF1001: Ищет возможности оптимизации аллокаций при выделении слайсов (`append` вместо ручного цикла).
//   - S1000: Находит блоки `select` с одним кейсом и рекомендует заменить их на чистое чтение/запись в канал.
//
// 3. Сторонние публичные анализаторы:
//   - errcheck: Гарантирует, что ни одна возвращаемая функциями ошибка (`error`) не была проигнорирована в коде.
//   - bodyclose: Следит за тем, чтобы тело HTTP-ответа (`resp.Body`) закрывалось методом `Close()` во избежание утечки дескрипторов.
//
// 4. Кастомные анализаторы проекта:
//   - OsExitAnalyzer: Запрещает прямой вызов функции `os.Exit` внутри функции `main` пакета `main`, мотивируя использовать корректный механизм возврата ошибок.
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
