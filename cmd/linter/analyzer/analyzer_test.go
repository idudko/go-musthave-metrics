// Package analyzer_test provides tests for the static analyzer
package analyzer_test

import (
	"testing"

	"github.com/idudko/go-musthave-metrics/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "a", "b", "c", "d", "e")
}
