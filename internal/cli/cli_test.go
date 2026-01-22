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
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

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
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create existing file
	os.WriteFile("docs-drift.yml", []byte("existing"), 0644)

	code := Run([]string{"init"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d when file exists, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Init_Force(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create existing file
	os.WriteFile("docs-drift.yml", []byte("existing"), 0644)

	code := Run([]string{"init", "--force"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d with --force, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	code := Run([]string{"check"}, "test")
	if code != ExitRuntimeErr {
		t.Errorf("expected exit code %d when config missing, got %d", ExitRuntimeErr, code)
	}
}

func TestRun_Check_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

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
	os.WriteFile("docs-drift.yml", []byte(config), 0644)

	// Create a markdown file with valid code
	readme := `# Test
` + "```javascript" + `
console.log("hello");
` + "```" + `
`
	os.WriteFile("README.md", []byte(readme), 0644)

	code := Run([]string{"check"}, "test")
	if code != ExitNoDrift {
		t.Errorf("expected exit code %d for valid code, got %d", ExitNoDrift, code)
	}
}

func TestRun_Check_WithDrift(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

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
	os.WriteFile("docs-drift.yml", []byte(config), 0644)

	// Create a markdown file with broken code
	readme := `# Test
` + "```javascript" + `
throw new Error("broken");
` + "```" + `
`
	os.WriteFile("README.md", []byte(readme), 0644)

	code := Run([]string{"check"}, "test")
	if code != ExitDrift {
		t.Errorf("expected exit code %d for drift, got %d", ExitDrift, code)
	}
}

func TestCollectFiles_SimplePattern(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create test files
	os.WriteFile("README.md", []byte("# README"), 0644)
	os.WriteFile("CONTRIBUTING.md", []byte("# Contributing"), 0644)
	os.WriteFile("main.go", []byte("package main"), 0644)

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
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create subdirectory with files
	os.MkdirAll("docs", 0755)
	os.WriteFile("README.md", []byte("# README"), 0644)
	os.WriteFile(filepath.Join("docs", "guide.md"), []byte("# Guide"), 0644)

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
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	os.WriteFile("README.md", []byte("# README"), 0644)

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
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	files, err := collectFiles([]string{"*.md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}
