package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_ParseFile_BasicBlocks(t *testing.T) {
	content := `# Test Document

Some text here.

` + "```javascript" + `
console.log("Hello, World!");
` + "```" + `

More text.

` + "```python" + `
print("Hello, World!")
` + "```" + `
`
	tmpFile := createTempMarkdown(t, content)
	defer os.Remove(tmpFile)

	p := New()
	blocks, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	// Check first block
	if blocks[0].Language != "javascript" {
		t.Errorf("expected javascript, got %s", blocks[0].Language)
	}
	if blocks[0].LineNumber != 5 {
		t.Errorf("expected line 5, got %d", blocks[0].LineNumber)
	}
	expected := `console.log("Hello, World!");`
	if blocks[0].Code != expected {
		t.Errorf("expected %q, got %q", expected, blocks[0].Code)
	}

	// Check second block
	if blocks[1].Language != "python" {
		t.Errorf("expected python, got %s", blocks[1].Language)
	}
	if blocks[1].LineNumber != 11 {
		t.Errorf("expected line 11, got %d", blocks[1].LineNumber)
	}
}

func TestParser_ParseFile_NoLanguage(t *testing.T) {
	content := `# Test

` + "```" + `
some code without language
` + "```" + `
`
	tmpFile := createTempMarkdown(t, content)
	defer os.Remove(tmpFile)

	p := New()
	blocks, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Language != "" {
		t.Errorf("expected empty language, got %s", blocks[0].Language)
	}
}

func TestParser_ParseFile_SkipDirective(t *testing.T) {
	content := `# Test

` + "```javascript docs-drift:skip" + `
// This should be skipped
throw new Error("Should not run");
` + "```" + `

` + "```javascript" + `
console.log("This should run");
` + "```" + `
`
	tmpFile := createTempMarkdown(t, content)
	defer os.Remove(tmpFile)

	p := New()
	blocks, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	if !blocks[0].Skip {
		t.Error("expected first block to be marked as skip")
	}

	if blocks[1].Skip {
		t.Error("expected second block to NOT be marked as skip")
	}
}

func TestParser_ParseFile_TildeFence(t *testing.T) {
	content := `# Test

~~~python
print("tilde fence")
~~~
`
	tmpFile := createTempMarkdown(t, content)
	defer os.Remove(tmpFile)

	p := New()
	blocks, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Language != "python" {
		t.Errorf("expected python, got %s", blocks[0].Language)
	}
}

func TestParser_ParseFile_MultilineCode(t *testing.T) {
	content := `# Test

` + "```javascript" + `
function hello() {
  console.log("Hello");
  console.log("World");
}
hello();
` + "```" + `
`
	tmpFile := createTempMarkdown(t, content)
	defer os.Remove(tmpFile)

	p := New()
	blocks, err := p.ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	expectedLines := 5
	actualLines := len(splitLines(blocks[0].Code))
	if actualLines != expectedLines {
		t.Errorf("expected %d lines in code, got %d", expectedLines, actualLines)
	}
}

func TestParser_ParseFile_FileNotFound(t *testing.T) {
	p := New()
	_, err := p.ParseFile("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFilterByLanguages(t *testing.T) {
	blocks := []CodeBlock{
		{Language: "javascript"},
		{Language: "python"},
		{Language: "rust"},
		{Language: ""},
	}

	filtered := FilterByLanguages(blocks, []string{"javascript", "python"})

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered blocks, got %d", len(filtered))
	}
}

func TestFilterByLanguages_EmptyFilter(t *testing.T) {
	blocks := []CodeBlock{
		{Language: "javascript"},
		{Language: "python"},
	}

	filtered := FilterByLanguages(blocks, []string{})

	if len(filtered) != 2 {
		t.Errorf("expected all blocks when filter is empty, got %d", len(filtered))
	}
}

func TestFilterByLanguages_Aliases(t *testing.T) {
	blocks := []CodeBlock{
		{Language: "js"},
		{Language: "py"},
	}

	filtered := FilterByLanguages(blocks, []string{"javascript", "python"})

	if len(filtered) != 2 {
		t.Errorf("expected 2 blocks (aliases should match), got %d", len(filtered))
	}
}

func TestFilterSkipped(t *testing.T) {
	blocks := []CodeBlock{
		{Language: "javascript", Skip: false},
		{Language: "javascript", Skip: true},
		{Language: "python", Skip: false},
	}

	filtered := FilterSkipped(blocks)

	if len(filtered) != 2 {
		t.Errorf("expected 2 non-skipped blocks, got %d", len(filtered))
	}

	for _, block := range filtered {
		if block.Skip {
			t.Error("found skipped block in filtered result")
		}
	}
}

// Helper functions

func createTempMarkdown(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
