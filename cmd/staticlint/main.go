// Analyzer set of analyzers for static analysis
//
// Contains:
// all analyzers of golang.org/x/tools/go/analysis
// all analyzers of https://pkg.go.dev/honnef.co/go/tools@v0.4.6/staticcheck and https://pkg.go.dev/honnef.co/go/tools@v0.4.6/stylecheck
// OsExitAnalyzer - checks call os.Exit() in main.
// go run main.go ./...
// help
// go run main.go help
package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// OsExitAnalyzer Custom os.Exit analyzer
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "OSExitAnalyzer",
	Doc:  "Провекра os.Exit в пакете main, в функции main()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.File:
				if x.Name.Name != "main" {
					return false
				}

				_, ok := x.Scope.Objects["tests"]
				if ok {
					return false
				}
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.SelectorExpr:

				if fmt.Sprintf("%v", x.X) == "os" && x.Sel.Name == "Exit" {
					pass.Report(analysis.Diagnostic{
						Pos:     x.Pos(),
						Message: "direct call to os.Exit found",
					})
				}
			}

			return true
		})
	}

	return nil, nil
}

func main() {
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	mychecks = append(
		mychecks,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		nilfunc.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		OsExitAnalyzer,
	)

	multichecker.Main(
		mychecks...,
	)
}
