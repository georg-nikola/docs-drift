package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the docs-drift.yml configuration
type Config struct {
	Version int          `yaml:"version"`
	Docs    DocsConfig   `yaml:"docs"`
	Checks  ChecksConfig `yaml:"checks"`
}

// DocsConfig specifies which documentation files to check
type DocsConfig struct {
	Paths []string `yaml:"paths"`
}

// ChecksConfig specifies what checks to run
type ChecksConfig struct {
	CodeBlocks CodeBlocksConfig `yaml:"code_blocks"`
}

// CodeBlocksConfig specifies code block validation settings
type CodeBlocksConfig struct {
	Enabled   bool          `yaml:"enabled"`
	Languages []string      `yaml:"languages"`
	Timeout   time.Duration `yaml:"timeout"`
}

// Default values
const (
	DefaultTimeout = 30 * time.Second
)

// Load reads and parses a config file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return Parse(data)
}

// Parse parses config from YAML data
func Parse(data []byte) (*Config, error) {
	cfg := &Config{
		// Set defaults
		Version: 1,
		Checks: ChecksConfig{
			CodeBlocks: CodeBlocksConfig{
				Enabled: true,
				Timeout: DefaultTimeout,
			},
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks the config for errors
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d (expected 1)", c.Version)
	}

	if len(c.Docs.Paths) == 0 {
		return fmt.Errorf("no documentation paths specified")
	}

	if c.Checks.CodeBlocks.Enabled && len(c.Checks.CodeBlocks.Languages) == 0 {
		return fmt.Errorf("code_blocks enabled but no languages specified")
	}

	// Validate supported languages
	supportedLangs := map[string]bool{
		"javascript": true,
		"js":         true,
		"python":     true,
		"py":         true,
	}

	for _, lang := range c.Checks.CodeBlocks.Languages {
		if !supportedLangs[lang] {
			return fmt.Errorf("unsupported language: %s (supported: javascript, python)", lang)
		}
	}

	if c.Checks.CodeBlocks.Timeout <= 0 {
		c.Checks.CodeBlocks.Timeout = DefaultTimeout
	}

	return nil
}

// IsLanguageEnabled checks if a language should be validated
func (c *Config) IsLanguageEnabled(lang string) bool {
	// Normalize language aliases
	normalized := NormalizeLanguage(lang)

	for _, l := range c.Checks.CodeBlocks.Languages {
		if NormalizeLanguage(l) == normalized {
			return true
		}
	}
	return false
}

// NormalizeLanguage converts language aliases to canonical names
func NormalizeLanguage(lang string) string {
	switch lang {
	case "js", "javascript":
		return "javascript"
	case "py", "python":
		return "python"
	default:
		return lang
	}
}
