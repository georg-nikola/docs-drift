package output

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// JSONOutput represents the JSON output format
type JSONOutput struct {
	Version   string        `json:"version"`
	Timestamp string        `json:"timestamp"`
	Summary   JSONSummary   `json:"summary"`
	Files     []JSONFile    `json:"files,omitempty"`
	Failures  []JSONFailure `json:"failures,omitempty"`
}

type JSONSummary struct {
	TotalFiles  int  `json:"total_files"`
	TotalBlocks int  `json:"total_blocks"`
	Passed      int  `json:"passed"`
	Failed      int  `json:"failed"`
	Skipped     int  `json:"skipped"`
	HasDrift    bool `json:"has_drift"`
}

type JSONFile struct {
	Path   string `json:"path"`
	Blocks int    `json:"blocks"`
}

type JSONFailure struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Language string `json:"language"`
	Error    string `json:"error"`
	Output   string `json:"output,omitempty"`
}

// FormatJSON formats results as JSON
func FormatJSON(results *drift.Results) (string, error) {
	output := JSONOutput{
		Version:   "1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Summary: JSONSummary{
			TotalFiles:  results.TotalFiles,
			TotalBlocks: results.TotalBlocks,
			Passed:      results.Passed,
			Failed:      len(results.Failed),
			Skipped:     results.Skipped,
			HasDrift:    results.HasDrift(),
		},
	}

	// Group blocks by file
	fileBlocks := make(map[string]int)
	for _, failure := range results.Failed {
		fileBlocks[failure.FilePath]++
	}

	for path, count := range fileBlocks {
		output.Files = append(output.Files, JSONFile{
			Path:   path,
			Blocks: count,
		})
	}

	// Add failures
	for _, failure := range results.Failed {
		output.Failures = append(output.Failures, JSONFailure{
			File:     failure.FilePath,
			Line:     failure.LineNumber,
			Language: failure.Language,
			Error:    failure.Error,
			Output:   failure.Output,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatHTML formats results as an HTML report
func FormatHTML(results *drift.Results) (string, error) {
	var sb strings.Builder

	// HTML header
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>docs-drift Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .header {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            margin: 0 0 10px 0;
            color: #333;
        }
        .timestamp {
            color: #666;
            font-size: 14px;
        }
        .summary {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
        }
        .summary-item {
            text-align: center;
        }
        .summary-value {
            font-size: 32px;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .summary-label {
            color: #666;
            font-size: 14px;
        }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .status-badge {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-weight: bold;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .status-pass {
            background-color: #d4edda;
            color: #155724;
        }
        .status-fail {
            background-color: #f8d7da;
            color: #721c24;
        }
        .failures {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .failure-item {
            border-left: 4px solid #dc3545;
            padding: 15px;
            margin-bottom: 15px;
            background-color: #f8f9fa;
        }
        .failure-header {
            font-weight: bold;
            margin-bottom: 10px;
        }
        .failure-location {
            color: #666;
            font-size: 14px;
            margin-bottom: 5px;
        }
        .failure-error {
            background-color: #fff;
            padding: 10px;
            border-radius: 4px;
            font-family: "Courier New", monospace;
            font-size: 13px;
            color: #dc3545;
            white-space: pre-wrap;
        }
        .no-failures {
            text-align: center;
            padding: 40px;
            color: #28a745;
            font-size: 18px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 docs-drift Report</h1>
        <div class="timestamp">Generated: ` + time.Now().Format("2006-01-02 15:04:05 MST") + `</div>
    </div>
`)

	// Status badge
	if results.HasDrift() {
		sb.WriteString(`    <div class="status-badge status-fail">❌ Drift Detected</div>` + "\n")
	} else {
		sb.WriteString(`    <div class="status-badge status-pass">✅ No Drift Detected</div>` + "\n")
	}

	// Summary
	sb.WriteString(`    <div class="summary">
        <h2>Summary</h2>
        <div class="summary-grid">
            <div class="summary-item">
                <div class="summary-value">` + fmt.Sprintf("%d", results.TotalFiles) + `</div>
                <div class="summary-label">Files Checked</div>
            </div>
            <div class="summary-item">
                <div class="summary-value">` + fmt.Sprintf("%d", results.TotalBlocks) + `</div>
                <div class="summary-label">Code Blocks</div>
            </div>
            <div class="summary-item">
                <div class="summary-value passed">` + fmt.Sprintf("%d", results.Passed) + `</div>
                <div class="summary-label">Passed</div>
            </div>
            <div class="summary-item">
                <div class="summary-value failed">` + fmt.Sprintf("%d", len(results.Failed)) + `</div>
                <div class="summary-label">Failed</div>
            </div>
            <div class="summary-item">
                <div class="summary-value skipped">` + fmt.Sprintf("%d", results.Skipped) + `</div>
                <div class="summary-label">Skipped</div>
            </div>
        </div>
    </div>
`)

	// Failures
	sb.WriteString(`    <div class="failures">
        <h2>Failures</h2>
`)

	if len(results.Failed) == 0 {
		sb.WriteString(`        <div class="no-failures">🎉 All code blocks passed validation!</div>` + "\n")
	} else {
		for _, failure := range results.Failed {
			sb.WriteString(`        <div class="failure-item">
            <div class="failure-header">` + html.EscapeString(failure.FilePath) + `</div>
            <div class="failure-location">Line ` + fmt.Sprintf("%d", failure.LineNumber) + ` • ` + html.EscapeString(failure.Language) + `</div>
            <div class="failure-error">` + html.EscapeString(failure.Error) + `</div>
        </div>
`)
		}
	}

	sb.WriteString(`    </div>
</body>
</html>`)

	return sb.String(), nil
}
