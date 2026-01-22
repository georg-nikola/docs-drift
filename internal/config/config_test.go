package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse_ValidConfig(t *testing.T) {
	yaml := `
version: 1
docs:
  paths:
    - README.md
    - docs/*.md
checks:
  code_blocks:
    enabled: true
    languages:
      - javascript
      - python
    timeout: 60s
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}

	if len(cfg.Docs.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(cfg.Docs.Paths))
	}

	if !cfg.Checks.CodeBlocks.Enabled {
		t.Error("expected code_blocks.enabled to be true")
	}

	if len(cfg.Checks.CodeBlocks.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(cfg.Checks.CodeBlocks.Languages))
	}

	if cfg.Checks.CodeBlocks.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Checks.CodeBlocks.Timeout)
	}
}

func TestParse_MinimalConfig(t *testing.T) {
	yaml := `
version: 1
docs:
  paths:
    - README.md
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use default timeout
	if cfg.Checks.CodeBlocks.Timeout != DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultTimeout, cfg.Checks.CodeBlocks.Timeout)
	}
}

func TestParse_InvalidVersion(t *testing.T) {
	yaml := `
version: 2
docs:
  paths:
    - README.md
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}

func TestParse_NoPaths(t *testing.T) {
	yaml := `
version: 1
docs:
  paths: []
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for empty paths")
	}
}

func TestParse_NoLanguages(t *testing.T) {
	yaml := `
version: 1
docs:
  paths:
    - README.md
checks:
  code_blocks:
    enabled: true
    languages: []
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for empty languages")
	}
}

func TestParse_UnsupportedLanguage(t *testing.T) {
	yaml := `
version: 1
docs:
  paths:
    - README.md
checks:
  code_blocks:
    enabled: true
    languages: [rust]
`
	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_ValidFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "docs-drift.yml")

	yaml := `
version: 1
docs:
  paths:
    - README.md
checks:
  code_blocks:
    enabled: true
    languages: [javascript]
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
}

func TestIsLanguageEnabled(t *testing.T) {
	cfg := &Config{
		Checks: ChecksConfig{
			CodeBlocks: CodeBlocksConfig{
				Languages: []string{"javascript", "python"},
			},
		},
	}

	tests := []struct {
		lang     string
		expected bool
	}{
		{"javascript", true},
		{"js", true},
		{"python", true},
		{"py", true},
		{"rust", false},
		{"go", false},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := cfg.IsLanguageEnabled(tt.lang)
			if result != tt.expected {
				t.Errorf("IsLanguageEnabled(%q) = %v, expected %v", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"javascript", "javascript"},
		{"js", "javascript"},
		{"python", "python"},
		{"py", "python"},
		{"rust", "rust"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeLanguage(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
