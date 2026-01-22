package drift

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/georg-nikola/docs-drift/internal/config"
)

// BenchmarkCheck_SingleFile benchmarks checking a single file with multiple code blocks
func BenchmarkCheck_SingleFile(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	// Create a file with multiple code blocks
	content := `# Test Documentation

This is a test file with multiple code blocks.

` + "```javascript" + `
console.log("Hello, World!");
` + "```" + `

Some text between blocks.

` + "```python" + `
print("Hello from Python")
` + "```" + `

More documentation here.

` + "```javascript" + `
const x = 1 + 2;
console.log(x);
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
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
	files := []string{mdFile}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.Check(files)
		if err != nil {
			b.Fatalf("Check failed: %v", err)
		}
	}
}

// BenchmarkCheckConcurrent_SingleFile benchmarks concurrent checking of a single file
func BenchmarkCheckConcurrent_SingleFile(b *testing.B) {
	tmpDir := b.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Test Documentation

` + "```javascript" + `
console.log("Hello, World!");
` + "```" + `

` + "```python" + `
print("Hello from Python")
` + "```" + `

` + "```javascript" + `
const x = 1 + 2;
console.log(x);
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write test file: %v", err)
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
	files := []string{mdFile}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckConcurrent(files, 4)
		if err != nil {
			b.Fatalf("CheckConcurrent failed: %v", err)
		}
	}
}

// BenchmarkCheck_MultipleFiles benchmarks checking multiple files
func BenchmarkCheck_MultipleFiles(b *testing.B) {
	tmpDir := b.TempDir()

	// Create 5 files with 2 code blocks each
	var files []string
	for i := 0; i < 5; i++ {
		mdFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.md", i))
		content := fmt.Sprintf(`# Test %d

`+"```javascript"+`
console.log("file %d");
`+"```"+`

`+"```python"+`
print("file %d")
`+"```"+`
`, i, i, i)
		if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write test file: %v", err)
		}
		files = append(files, mdFile)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.Check(files)
		if err != nil {
			b.Fatalf("Check failed: %v", err)
		}
	}
}

// BenchmarkCheckConcurrent_MultipleFiles benchmarks concurrent checking of multiple files
func BenchmarkCheckConcurrent_MultipleFiles(b *testing.B) {
	tmpDir := b.TempDir()

	// Create 5 files with 2 code blocks each
	var files []string
	for i := 0; i < 5; i++ {
		mdFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.md", i))
		content := fmt.Sprintf(`# Test %d

`+"```javascript"+`
console.log("file %d");
`+"```"+`

`+"```python"+`
print("file %d")
`+"```"+`
`, i, i, i)
		if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write test file: %v", err)
		}
		files = append(files, mdFile)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckConcurrent(files, 4)
		if err != nil {
			b.Fatalf("CheckConcurrent failed: %v", err)
		}
	}
}

// BenchmarkCheckConcurrent_Workers compares different worker counts
func BenchmarkCheckConcurrent_Workers(b *testing.B) {
	tmpDir := b.TempDir()

	// Create 10 files with 3 code blocks each
	var files []string
	for i := 0; i < 10; i++ {
		mdFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.md", i))
		content := fmt.Sprintf(`# Test %d

`+"```javascript"+`
console.log("block 1 file %d");
`+"```"+`

`+"```javascript"+`
console.log("block 2 file %d");
`+"```"+`

`+"```javascript"+`
console.log("block 3 file %d");
`+"```"+`
`, i, i, i, i)
		if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write test file: %v", err)
		}
		files = append(files, mdFile)
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

	workerCounts := []int{1, 2, 4, 8}
	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("workers_%d", workers), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := checker.CheckConcurrent(files, workers)
				if err != nil {
					b.Fatalf("CheckConcurrent failed: %v", err)
				}
			}
		})
	}
}
