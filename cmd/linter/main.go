// Package linter provides a static analyzer for Go code that detects:
// - Usage of panic function
// - Calls to log.Fatal or os.Exit outside of main function in main package
package main

import (
	"github.com/idudko/go-musthave-metrics/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
