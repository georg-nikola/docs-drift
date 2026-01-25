package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/georg-nikola/docs-drift/pkg/drift"
)

func TestFormatJSON_NoFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 2,
		Passed:      2,
		Failed:      []drift.BlockResult{},
		Skipped:     0,
	}

	output, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed JSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify summary
	if parsed.Summary.TotalFiles != 1 {
		t.Errorf("expected TotalFiles=1, got %d", parsed.Summary.TotalFiles)
	}
	if parsed.Summary.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", parsed.Summary.Passed)
	}
	if parsed.Summary.HasDrift {
		t.Error("expected HasDrift=false")
	}
	if parsed.Version != "1.0" {
		t.Errorf("expected Version=1.0, got %s", parsed.Version)
	}
}

func TestFormatJSON_WithFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 2,
		Passed:      1,
		Failed: []drift.BlockResult{
			{
				FilePath:   "README.md",
				LineNumber: 10,
				Language:   "javascript",
				Success:    false,
				Error:      "ReferenceError: x is not defined",
				Output:     "some output",
			},
		},
		Skipped: 0,
	}

	output, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var parsed JSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if !parsed.Summary.HasDrift {
		t.Error("expected HasDrift=true")
	}
	if len(parsed.Failures) != 1 {
		t.Errorf("expected 1 failure, got %d", len(parsed.Failures))
	}
	if parsed.Failures[0].File != "README.md" {
		t.Errorf("expected failure in README.md, got %s", parsed.Failures[0].File)
	}
	if parsed.Failures[0].Line != 10 {
		t.Errorf("expected line 10, got %d", parsed.Failures[0].Line)
	}
	if parsed.Failures[0].Language != "javascript" {
		t.Errorf("expected language javascript, got %s", parsed.Failures[0].Language)
	}
	if parsed.Failures[0].Error != "ReferenceError: x is not defined" {
		t.Errorf("expected error message, got %s", parsed.Failures[0].Error)
	}
	if parsed.Failures[0].Output != "some output" {
		t.Errorf("expected output, got %s", parsed.Failures[0].Output)
	}
}

func TestFormatJSON_MultipleFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  2,
		TotalBlocks: 4,
		Passed:      2,
		Failed: []drift.BlockResult{
			{
				FilePath:   "README.md",
				LineNumber: 10,
				Language:   "javascript",
				Success:    false,
				Error:      "Error 1",
			},
			{
				FilePath:   "GUIDE.md",
				LineNumber: 20,
				Language:   "python",
				Success:    false,
				Error:      "Error 2",
			},
		},
		Skipped: 0,
	}

	output, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var parsed JSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if len(parsed.Failures) != 2 {
		t.Errorf("expected 2 failures, got %d", len(parsed.Failures))
	}
	if len(parsed.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(parsed.Files))
	}
}

func TestFormatJSON_WithSkipped(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 3,
		Passed:      2,
		Failed:      []drift.BlockResult{},
		Skipped:     1,
	}

	output, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	var parsed JSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if parsed.Summary.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d", parsed.Summary.Skipped)
	}
}

func TestFormatHTML_NoFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 2,
		Passed:      2,
		Failed:      []drift.BlockResult{},
		Skipped:     0,
	}

	output, err := FormatHTML(results)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	// Verify it's valid HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("expected HTML doctype")
	}
	if !strings.Contains(output, "No Drift Detected") {
		t.Error("expected 'No Drift Detected' message")
	}
	if !strings.Contains(output, "All code blocks passed") {
		t.Error("expected success message")
	}
	if !strings.Contains(output, "docs-drift Report") {
		t.Error("expected page title")
	}
	if !strings.Contains(output, "status-pass") {
		t.Error("expected pass status badge")
	}
}

func TestFormatHTML_WithFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 2,
		Passed:      1,
		Failed: []drift.BlockResult{
			{
				FilePath:   "README.md",
				LineNumber: 10,
				Language:   "javascript",
				Success:    false,
				Error:      "ReferenceError: x is not defined",
			},
		},
		Skipped: 0,
	}

	output, err := FormatHTML(results)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	if !strings.Contains(output, "Drift Detected") {
		t.Error("expected 'Drift Detected' message")
	}
	if !strings.Contains(output, "README.md") {
		t.Error("expected file name in output")
	}
	if !strings.Contains(output, "ReferenceError") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(output, "Line 10") {
		t.Error("expected line number in output")
	}
	if !strings.Contains(output, "javascript") {
		t.Error("expected language in output")
	}
	if !strings.Contains(output, "status-fail") {
		t.Error("expected fail status badge")
	}
}

func TestFormatHTML_MultipleFailures(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  2,
		TotalBlocks: 4,
		Passed:      2,
		Failed: []drift.BlockResult{
			{
				FilePath:   "README.md",
				LineNumber: 10,
				Language:   "javascript",
				Success:    false,
				Error:      "Error 1",
			},
			{
				FilePath:   "GUIDE.md",
				LineNumber: 20,
				Language:   "python",
				Success:    false,
				Error:      "Error 2",
			},
		},
		Skipped: 0,
	}

	output, err := FormatHTML(results)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	if !strings.Contains(output, "README.md") {
		t.Error("expected README.md in output")
	}
	if !strings.Contains(output, "GUIDE.md") {
		t.Error("expected GUIDE.md in output")
	}
	if !strings.Contains(output, "Error 1") {
		t.Error("expected Error 1 in output")
	}
	if !strings.Contains(output, "Error 2") {
		t.Error("expected Error 2 in output")
	}
}

func TestFormatHTML_HTMLEscaping(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  1,
		TotalBlocks: 1,
		Passed:      0,
		Failed: []drift.BlockResult{
			{
				FilePath:   "<script>alert('xss')</script>",
				LineNumber: 10,
				Language:   "javascript",
				Success:    false,
				Error:      "<img src=x onerror=alert(1)>",
			},
		},
		Skipped: 0,
	}

	output, err := FormatHTML(results)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	// Verify HTML is escaped
	if strings.Contains(output, "<script>alert('xss')</script>") {
		t.Error("HTML should be escaped in file path")
	}
	if strings.Contains(output, "<img src=x onerror=alert(1)>") {
		t.Error("HTML should be escaped in error message")
	}
	if !strings.Contains(output, "&lt;script&gt;") {
		t.Error("expected escaped HTML in output")
	}
}

func TestFormatHTML_Summary(t *testing.T) {
	results := &drift.Results{
		TotalFiles:  5,
		TotalBlocks: 10,
		Passed:      7,
		Failed:      []drift.BlockResult{{FilePath: "test.md", LineNumber: 1, Language: "js"}},
		Skipped:     2,
	}

	output, err := FormatHTML(results)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	// Check all summary values are present
	if !strings.Contains(output, "5") { // Total files
		t.Error("expected total files count in output")
	}
	if !strings.Contains(output, "10") { // Total blocks
		t.Error("expected total blocks count in output")
	}
	if !strings.Contains(output, "7") { // Passed
		t.Error("expected passed count in output")
	}
	if !strings.Contains(output, "2") { // Skipped
		t.Error("expected skipped count in output")
	}
}
