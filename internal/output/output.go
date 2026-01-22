package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/georg-nikola/docs-drift/pkg/drift"
)

// Colors for terminal output (ANSI escape codes)
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// PrintResults outputs the check results to stdout/stderr
func PrintResults(results *drift.Results, verbose bool) {
	if results.TotalBlocks == 0 {
		fmt.Println("No code blocks found to check")
		return
	}

	if verbose {
		fmt.Printf("\nChecked %d code block(s) across %d file(s)\n",
			results.TotalBlocks, results.TotalFiles)
	}

	if !results.HasDrift() {
		printSuccess(results, verbose)
		return
	}

	printDrift(results)
}

func printSuccess(results *drift.Results, verbose bool) {
	if isTerminal() {
		fmt.Printf("%s%sNo drift detected%s\n", colorBold, colorGreen, colorReset)
	} else {
		fmt.Println("No drift detected")
	}

	if verbose {
		fmt.Printf("  %d code block(s) passed\n", results.Passed)
		if results.Skipped > 0 {
			fmt.Printf("  %d code block(s) skipped\n", results.Skipped)
		}
	}
}

func printDrift(results *drift.Results) {
	// Print header
	if isTerminal() {
		fmt.Fprintf(os.Stderr, "\n%s%sDocs Drift Detected%s\n\n", colorBold, colorRed, colorReset)
	} else {
		fmt.Fprint(os.Stderr, "\nDocs Drift Detected\n\n")
	}

	// Group errors by file
	errorsByFile := make(map[string][]drift.BlockResult)
	for _, err := range results.Failed {
		errorsByFile[err.FilePath] = append(errorsByFile[err.FilePath], err)
	}

	// Print errors grouped by file
	for filePath, errors := range errorsByFile {
		// Make path relative for cleaner output
		relPath := filePath
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, filePath); err == nil {
				relPath = rel
			}
		}

		if isTerminal() {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", colorCyan, relPath, colorReset)
		} else {
			fmt.Fprintln(os.Stderr, relPath)
		}

		for _, blockErr := range errors {
			printBlockError(blockErr)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Print summary
	printSummary(results)
}

func printBlockError(err drift.BlockResult) {
	description := fmt.Sprintf("%s code block", err.Language)

	if isTerminal() {
		fmt.Fprintf(os.Stderr, "  Line %d: %s\n", err.LineNumber, description)
		fmt.Fprintf(os.Stderr, "    %s%s%s\n", colorRed, formatError(err.Error), colorReset)
	} else {
		fmt.Fprintf(os.Stderr, "  Line %d: %s\n", err.LineNumber, description)
		fmt.Fprintf(os.Stderr, "    -> %s\n", formatError(err.Error))
	}
}

func printSummary(results *drift.Results) {
	if isTerminal() {
		fmt.Fprintf(os.Stderr, "%sSummary:%s %d failed, %d passed",
			colorBold, colorReset, len(results.Failed), results.Passed)
	} else {
		fmt.Fprintf(os.Stderr, "Summary: %d failed, %d passed", len(results.Failed), results.Passed)
	}

	if results.Skipped > 0 {
		fmt.Fprintf(os.Stderr, ", %d skipped", results.Skipped)
	}
	fmt.Fprintln(os.Stderr)
}

// formatError cleans up error messages for display
func formatError(err string) string {
	// Truncate very long errors
	const maxLen = 200
	if len(err) > maxLen {
		err = err[:maxLen] + "..."
	}

	// Remove file paths from stack traces for cleaner output
	lines := strings.Split(err, "\n")
	var cleaned []string
	for _, line := range lines {
		// Skip lines that are just file paths
		if strings.HasPrefix(strings.TrimSpace(line), "at ") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, " ")
}

// isTerminal checks if stdout is a terminal (for color output)
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// FormatJSON returns results in JSON format (for CI integration)
func FormatJSON(results *drift.Results) string {
	// Simple JSON formatting without external dependencies
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf(`  "hasDrift": %t,`, results.HasDrift()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  "totalFiles": %d,`, results.TotalFiles))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  "totalBlocks": %d,`, results.TotalBlocks))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  "passed": %d,`, results.Passed))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  "failed": %d,`, len(results.Failed)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  "skipped": %d,`, results.Skipped))
	sb.WriteString("\n")
	sb.WriteString(`  "errors": [`)

	for i, err := range results.Failed {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n    {")
		sb.WriteString(fmt.Sprintf(`"file": %q, `, err.FilePath))
		sb.WriteString(fmt.Sprintf(`"line": %d, `, err.LineNumber))
		sb.WriteString(fmt.Sprintf(`"language": %q, `, err.Language))
		sb.WriteString(fmt.Sprintf(`"error": %q`, escapeJSON(err.Error)))
		sb.WriteString("}")
	}

	if len(results.Failed) > 0 {
		sb.WriteString("\n  ")
	}
	sb.WriteString("]\n}")

	return sb.String()
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
