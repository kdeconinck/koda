// =====================================================================================================================
// = LICENSE:       Copyright (c) 2026 Kevin De Coninck
// =
// =                Permission is hereby granted, free of charge, to any person
// =                obtaining a copy of this software and associated documentation
// =                files (the "Software"), to deal in the Software without
// =                restriction, including without limitation the rights to use,
// =                copy, modify, merge, publish, distribute, sublicense, and/or sell
// =                copies of the Software, and to permit persons to whom the
// =                Software is furnished to do so, subject to the following
// =                conditions:
// =
// =                The above copyright notice and this permission notice shall be
// =                included in all copies or substantial portions of the Software.
// =
// =                THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// =                EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// =                OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// =                NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// =                HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// =                WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// =                FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// =                OTHER DEALINGS IN THE SOFTWARE.
// =====================================================================================================================

// Koda CLI
//
// This is the binary entry point for the Koda universal linter and formatter.
// It handles command-line argument parsing, file I/O, and coordinates between the configuration parser and the
// analysis engine.
//
// A Koda configuration is defined in a `.core` file. Each file contains one or more language profiles, where every
// profile describes:
//
// - The language name and supported file extensions.
// - Reusable character sets.
// - Token definitions.
// - Structural analysis rules.
//
// The Koda configuration language is written using explicit sections and declarative definitions.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/engine"
	"github.com/kdeconinck/koda/internal/parser"
	"github.com/kdeconinck/koda/internal/scanner"
	"github.com/kdeconinck/koda/internal/validator"
)

const version = "0.1.0"

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// FileReport contains all diagnostics produced for a single target file.
type FileReport struct {
	// FileName is the display name of the analyzed file.
	FileName string

	// Diagnostics contains all issues reported for this file.
	Diagnostics []engine.Diagnostic
}

// The main entry point of the application.
//
// This function should remain minimal. Its only responsibility is to delegate execution to run and translate returned
// errors into proper process exit codes.
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

// Contains the high-level orchestration logic of the CLI.
//
// This function exists so that:
//   - The application logic can be tested independently.
//   - The function "main" remains minimal and free of business logic.
//   - The program avoids calling os.Exit outside of main.
//
// It should return an error instead of terminating the process.
func run() error {
	args := os.Args[1:]

	if len(args) != 2 {
		printHeader()
		printUsage()

		return fmt.Errorf("%sinvalid arguments%s", colorRed, colorReset)
	}

	configPath, targetPath := args[0], args[1]

	printHeader()
	fmt.Printf("%sConfig:%s %s\n", colorBlue, colorReset, configPath)
	fmt.Printf("%sTarget:%s %s\n\n", colorBlue, colorReset, targetPath)

	configSource, err := os.ReadFile(configPath)

	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	targetSource, err := os.ReadFile(targetPath)

	if err != nil {
		return fmt.Errorf("failed to read target file %q: %w", targetPath, err)
	}

	tokens, err := scanner.Scan(string(configSource))

	if err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	profile, err := parser.Parse(tokens)

	if err != nil {
		return fmt.Errorf("parser error: %w", err)
	}

	if err := validator.Validate(profile); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	compiledProfile := compiler.Compile(profile)

	result := engine.Analyze(compiledProfile, string(targetSource))

	reports := []FileReport{
		{
			FileName:    targetPath,
			Diagnostics: result.Diagnostics,
		},
	}

	// Print diagnostics.
	printReports(reports)

	return nil
}

// Prints the CLI header with product name and version information.
func printHeader() {
	fmt.Println("")
	fmt.Printf(
		"%s%sKoda%s - %sThe universal linter & code formatter.%s %s(v%s)%s\n",
		colorBold,
		colorCyan,
		colorReset,
		colorDim,
		colorReset,
		colorYellow,
		version,
		colorReset,
	)

	fmt.Printf("%s%s%s\n\n", colorDim, "------------------------------------------------------------", colorReset)
}

// Prints the command usage in a readable format.
func printUsage() {
	fmt.Printf("%sUsage:%s koda <config.core> <target-file>\n\n", colorBold, colorReset)
}

func printReports(reports []FileReport) {
	totalIssues := countDiagnostics(reports)

	if totalIssues == 0 {
		fmt.Printf("%s%sNo issues detected.%s\n", colorGreen, colorBold, colorReset)
		return
	}

	sortedReports := sortReports(reports)

	for _, report := range sortedReports {
		if len(report.Diagnostics) == 0 {
			continue
		}

		printFileReport(report)
	}

	fmt.Printf(
		"%s%sTotal issues detected:%s %s%d%s\n",
		colorBold,
		colorYellow,
		colorReset,
		colorRed,
		totalIssues,
		colorReset,
	)
}

// Prints the diagnostics of a single file as one visual group.
func printFileReport(report FileReport) {
	fmt.Printf(
		"%s%sFile:%s %s%s%s\n",
		colorBold,
		colorBlue,
		colorReset,
		colorCyan,
		report.FileName,
		colorReset,
	)

	fmt.Printf("%s%s%s\n", colorDim, "", colorReset)

	for _, diagnostic := range report.Diagnostics {
		fmt.Printf(
			"  %s[%d:%d]%s %s→%s %s\n",
			colorYellow,
			diagnostic.Span.Start.Line,
			diagnostic.Span.Start.Column,
			colorReset,
			colorRed,
			colorReset,
			diagnostic.Message,
		)
	}

	fmt.Println()
}

// Returns the total number of diagnostics across all file reports.
func countDiagnostics(reports []FileReport) int {
	total := 0

	for _, report := range reports {
		total += len(report.Diagnostics)
	}

	return total
}

// Returns file reports sorted by file name for stable output.
func sortReports(reports []FileReport) []FileReport {
	sorted := append([]FileReport(nil), reports...)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FileName < sorted[j].FileName
	})

	return sorted
}
