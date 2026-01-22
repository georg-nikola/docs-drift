package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkParseFile_Small benchmarks parsing a small markdown file
func BenchmarkParseFile_Small(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Small Test File

` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	p := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(mdFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

// BenchmarkParseFile_Medium benchmarks parsing a medium markdown file
func BenchmarkParseFile_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create a file with 10 code blocks
	var sb strings.Builder
	sb.WriteString("# Medium Test File\n\n")
	for i := 0; i < 10; i++ {
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("Some documentation text here.\n\n")
		sb.WriteString("```javascript\n")
		sb.WriteString(fmt.Sprintf("console.log(\"section %d\");\n", i))
		sb.WriteString("```\n\n")
	}

	if err := os.WriteFile(mdFile, []byte(sb.String()), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	p := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(mdFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

// BenchmarkParseFile_Large benchmarks parsing a large markdown file
func BenchmarkParseFile_Large(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create a file with 50 code blocks of various languages
	var sb strings.Builder
	sb.WriteString("# Large Test File\n\n")
	languages := []string{"javascript", "python", "go", "rust", "bash"}
	for i := 0; i < 50; i++ {
		lang := languages[i%len(languages)]
		sb.WriteString(fmt.Sprintf("## Section %d\n\n", i))
		sb.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
		sb.WriteString("Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n")
		sb.WriteString(fmt.Sprintf("```%s\n", lang))
		sb.WriteString(fmt.Sprintf("// Code block %d\n", i))
		sb.WriteString("// Some code here\n")
		sb.WriteString("```\n\n")
	}

	if err := os.WriteFile(mdFile, []byte(sb.String()), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	p := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(mdFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

// BenchmarkParseFile_WithSkipDirectives benchmarks parsing with skip directives
func BenchmarkParseFile_WithSkipDirectives(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	var sb strings.Builder
	sb.WriteString("# Test File with Skip Directives\n\n")
	for i := 0; i < 20; i++ {
		directive := ""
		if i%3 == 0 {
			directive = " docs-drift:skip"
		}
		sb.WriteString(fmt.Sprintf("```javascript%s\n", directive))
		sb.WriteString(fmt.Sprintf("console.log(\"block %d\");\n", i))
		sb.WriteString("```\n\n")
	}

	if err := os.WriteFile(mdFile, []byte(sb.String()), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
	}

	p := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.ParseFile(mdFile)
		if err != nil {
			b.Fatalf("ParseFile failed: %v", err)
		}
	}
}

// BenchmarkFilterByLanguages benchmarks filtering code blocks by language
func BenchmarkFilterByLanguages(b *testing.B) {
	// Create a slice of code blocks
	blocks := make([]CodeBlock, 100)
	languages := []string{"javascript", "python", "go", "rust", "bash", "typescript"}
	for i := range blocks {
		blocks[i] = CodeBlock{
			Language:   languages[i%len(languages)],
			Code:       fmt.Sprintf("code block %d", i),
			LineNumber: i * 10,
			FilePath:   "/test/file.md",
		}
	}

	filter := []string{"javascript", "python"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterByLanguages(blocks, filter)
	}
}
