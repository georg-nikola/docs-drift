package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRun_Help(t *testing.T) {
	tests := []struct {
		args     []string
		exitCode int
	}{
		{[]string{"help"}, ExitNoDrift},
		{[]string{"-h"}, ExitNoDrift},
		{[]string{"--help"}, ExitNoDrift},
	}

	for _, tt := range tests {
		t.Run(tt.args[0], func(t *testing.T) {
			code := Run(tt.args, "test")
			if code != tt.exitCode {
				t.Errorf("expected exit code %d, got %d", tt.exitCode, code)
			}
		})
	}
}

func TestRun_Version(t *testing.T) {
	code := Run([]string{"version"}, "1.2.3")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d, got %d", ExitNoDrift, code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code := Run([]string{}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d for no args, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	code := Run([]string{"unknown"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d for unknown command, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Init(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	code := Run([]string{"init"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d, got %d", ExitNoDrift, code)
	}

	// Check file was created
	if _, err := os.Stat("docs-drift.yml"); err != nil {
		t.Error("docs-drift.yml was not created")
	}
}

func TestRun_Init_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create existing file
	if err := os.WriteFile("docs-drift.yml", []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	code := Run([]string{"init"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d when file exists, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Init_Force(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create existing file
	if err := os.WriteFile("docs-drift.yml", []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	code := Run([]string{"init", "--force"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d with --force, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	code := Run([]string{"check"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d when config missing, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Check_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a markdown file with valid code
	readme := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile("README.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write README file: %v", err)
	}

	code := Run([]string{"check"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d for valid code, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_WithDrift(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create a markdown file with broken code
	readme := `# Test
` + "```javascript" + `
throw new Error("broken");
` + "```" + `
`
	if err := os.WriteFile("README.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write README file: %v", err)
	}

	code := Run([]string{"check"}, "test")
	if code != ExitDrift {
		t.Errorf("expected exit code %d for drift, got %d", ExitDrift, code)
	}
}

func TestCollectFiles_SimplePattern(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create test files
	if err := os.WriteFile("README.md", []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}
	if err := os.WriteFile("CONTRIBUTING.md", []byte("# Contributing"), 0644); err != nil {
		t.Fatalf("failed to write CONTRIBUTING.md: %v", err)
	}
	if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	files, err := collectFiles([]string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestCollectFiles_Subdirectory(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create subdirectory with files
	if err := os.MkdirAll("docs", 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile("README.md", []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join("docs", "guide.md"), []byte("# Guide"), 0644); err != nil {
		t.Fatalf("failed to write guide.md: %v", err)
	}

	files, err := collectFiles([]string{"*.md", "docs/*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestCollectFiles_NoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	if err := os.WriteFile("README.md", []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Same file matched by multiple patterns
	files, err := collectFiles([]string{"*.md", "README.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file (no duplicates), got %d", len(files))
	}
}

func TestCollectFiles_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	files, err := collectFiles([]string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestRun_Check_Parallel(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create markdown files with valid code
	readme := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile("README.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}
	if err := os.WriteFile("GUIDE.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write GUIDE.md: %v", err)
	}

	code := Run([]string{"check", "--parallel"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d for parallel check, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_ParallelWithWorkers(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create markdown file with valid code
	readme := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile("README.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	code := Run([]string{"check", "--parallel", "--workers", "2"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d for parallel check with workers, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_InvalidWorkers(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	code := Run([]string{"check", "--workers", "0"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d for invalid workers, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Check_ChangedOnlyNotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Not a git repo, so --changed-only should fail
	code := Run([]string{"check", "--changed-only"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d for --changed-only outside git repo, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Check_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config
	config := `version: 1
docs:
  paths:
    - "*.md"
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile("docs-drift.yml", []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Create markdown file with valid code
	readme := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	if err := os.WriteFile("README.md", []byte(readme), 0644); err != nil {
		t.Fatalf("failed to write README.md: %v", err)
	}

	code := Run([]string{"check", "--verbose"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d for verbose check, got %d", ExitNoDrift, code)
	}
}
