package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Bessima/metrics-collect/cmd/linter/errcheckanalyzer"
)

func main() {
	singlechecker.Main(errcheckanalyzer.CriticErrorAnalyzer)
}
