package drift

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/georg-nikola/docs-drift/internal/config"
)

func TestResults_HasDrift(t *testing.T) {
	tests := []struct {
		name     string
		results  Results
		expected bool
	}{
		{
			name:     "no drift",
			results:  Results{Failed: nil},
			expected: false,
		},
		{
			name:     "empty failed",
			results:  Results{Failed: []BlockResult{}},
			expected: false,
		},
		{
			name:     "has drift",
			results:  Results{Failed: []BlockResult{{Error: "test"}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.results.HasDrift() != tt.expected {
				t.Errorf("HasDrift() = %v, expected %v", tt.results.HasDrift(), tt.expected)
			}
		})
	}
}

func TestChecker_Check_NoFiles(t *testing.T) {
	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", results.TotalFiles)
	}

	if results.TotalBlocks != 0 {
		t.Errorf("expected 0 blocks, got %d", results.TotalBlocks)
	}
}

func TestChecker_Check_PassingCode(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.HasDrift() {
		t.Error("expected no drift for valid code")
	}

	if results.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", results.Passed)
	}
}

func TestChecker_Check_FailingCode(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test
` + "```javascript" + `
throw new Error("intentional error");
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !results.HasDrift() {
		t.Error("expected drift for failing code")
	}

	if len(results.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(results.Failed))
	}
}

func TestChecker_Check_SkippedBlock(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test
` + "```javascript docs-drift:skip" + `
throw new Error("should be skipped");
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.HasDrift() {
		t.Error("expected no drift for skipped code")
	}

	if results.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", results.Skipped)
	}
}

func TestChecker_Check_MixedResults(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test

` + "```javascript" + `
console.log("pass");
` + "```" + `

` + "```javascript" + `
throw new Error("fail");
` + "```" + `

` + "```javascript docs-drift:skip" + `
throw new Error("skipped");
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.TotalBlocks != 3 {
		t.Errorf("expected 3 total blocks, got %d", results.TotalBlocks)
	}

	if results.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", results.Passed)
	}

	if len(results.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(results.Failed))
	}

	if results.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", results.Skipped)
	}
}

func TestChecker_Check_MultipleLanguages(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test

` + "```javascript" + `
console.log("js pass");
` + "```" + `

` + "```python" + `
print("py pass")
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript", "python"},
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.HasDrift() {
		t.Error("expected no drift")
	}

	if results.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", results.Passed)
	}
}

func TestChecker_Check_FilteredLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test

` + "```javascript" + `
console.log("js pass");
` + "```" + `

` + "```python" + `
raise Exception("should not run - python not enabled")
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg := &config.Config{
		Checks: config.ChecksConfig{
			CodeBlocks: config.CodeBlocksConfig{
				Enabled:   true,
				Languages: []string{"javascript"}, // Python not enabled
				Timeout:   30 * time.Second,
			},
		},
	}

	checker := NewChecker(cfg, false)
	results, err := checker.Check([]string{mdFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.HasDrift() {
		t.Error("expected no drift (python should be ignored)")
	}

	if results.TotalBlocks != 1 {
		t.Errorf("expected 1 block (only JS), got %d", results.TotalBlocks)
	}
}
