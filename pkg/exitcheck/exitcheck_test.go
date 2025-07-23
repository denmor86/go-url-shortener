package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestExitCheck(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "./...") // тест без ошибкой
}
