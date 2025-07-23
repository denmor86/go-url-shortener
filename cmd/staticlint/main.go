package main

import (
	"strings"

	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/denmor86/go-url-shortener/pkg/exitcheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,          // Проверяет соответствие строк формата и аргументов в Printf-подобных функциях
		shadow.Analyzer,          // Обнаруживает затенение (переопределение) переменных во внутренних scope
		structtag.Analyzer,       // Проверяет корректность тегов структур (json, yaml и др.)
		appends.Analyzer,         // Выявляет некорректные вызовы append (например, без присваивания)
		assign.Analyzer,          // Находит бесполезные присваивания (x = x)
		atomic.Analyzer,          // Проверяет корректность использования atomic-операций
		bools.Analyzer,           // Обнаруживает избыточные логические выражения (x == true)
		buildtag.Analyzer,        // Проверяет корректность build-тегов в файлах
		copylock.Analyzer,        // Выявляет копирование мьютексов и других блокирующих структур
		deepequalerrors.Analyzer, // Предупреждает о некорректном DeepEqual для ошибок
		defers.Analyzer,          // Находит подозрительные defer-вызовы (например, в цикле)
		directive.Analyzer,       // Анализирует Go-директивы (//go:generate и др.)
		findcall.Analyzer,        // Ищет вызовы функций по заданному шаблону
		httpmux.Analyzer,         // Проверяет корректность использования http.ServeMux
		httpresponse.Analyzer,    // Выявляет необработанные HTTP-ответы
		loopclosure.Analyzer,     // Обнаруживает проблемы с замыканиями в циклах
		lostcancel.Analyzer,      // Находит утечку контекстов (невызов cancel())
		tests.Analyzer,           // Проверяет тесты на распространённые ошибки
		nilness.Analyzer,         // Анализирует возможные nil-паники
		unmarshal.Analyzer,       // Проверяет корректность unmarshal в структуры
		unreachable.Analyzer,     // Выявляет недостижимый код
		bodyclose.Analyzer,       // Проверяет закрытие HTTP-ответов (Body.Close())
		exitcheck.Analyzer,       // Проверяет отсутствие os.Exit() в функции main пакета main
	}

	staticcheckIncludes := map[string]bool{
		"S1001":   true, // Замена цикла for-range с игнорированием значений на for {}
		"SA1019":  true, // Использование устаревших функций или пакетов
		"ST1005":  true, // Проверка стиля строк ошибок (должны быть без заглавных букв и точек)
		"QF1008 ": true, // Удаление лишних скобок в if
		"QF1011":  true, // Замена &T{...} на new(T) для простых конструкторов
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") ||
			staticcheckIncludes[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
