package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/georg-nikola/docs-drift/internal/config"
	"github.com/georg-nikola/docs-drift/internal/git"
	"github.com/georg-nikola/docs-drift/internal/output"
	"github.com/georg-nikola/docs-drift/pkg/drift"
)

// Exit codes as per spec
const (
	ExitNoDrift    = 0
	ExitDrift      = 1
	ExitRuntimeErr = 2
)

// Default values for parallel execution
const (
	DefaultWorkers = 4
)

// Run executes the CLI with the given arguments and returns an exit code
func Run(args []string, version string) int {
	if len(args) == 0 {
		printUsage()
		return ExitRuntimeErr
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "check":
		return runCheck(args[1:])
	case "version":
		return runVersion(version)
	case "help", "-h", "--help":
		printUsage()
		return ExitNoDrift
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		printUsage()
		return ExitRuntimeErr
	}
}

func printUsage() {
	fmt.Println(`docs-drift - Detect documentation drift in code examples

Usage:
  docs-drift <command> [options]

Commands:
  init      Create a docs-drift.yml config file
  check     Validate code blocks in documentation
  version   Show version information
  help      Show this help message

Check Options:
  --config <path>      Path to config file (default: docs-drift.yml)
  --verbose            Show verbose output
  --changed-only       Only check changed markdown files (requires git)
  --base <ref>         Base reference for --changed-only (default: auto-detect)
  --parallel           Run code block checks in parallel
  --workers <n>        Number of parallel workers (default: 4)

Examples:
  docs-drift init
  docs-drift check
  docs-drift check --config ./custom-config.yml
  docs-drift check --changed-only
  docs-drift check --changed-only --base main
  docs-drift check --parallel --workers 8`)
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "Overwrite existing config file")

	if err := fs.Parse(args); err != nil {
		return ExitRuntimeErr
	}

	configPath := "docs-drift.yml"

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "Config file %s already exists. Use --force to overwrite.\n", configPath)
		return ExitRuntimeErr
	}

	defaultConfig := `# docs-drift configuration
# https://github.com/georg-nikola/docs-drift

version: 1

docs:
  paths:
    - README.md
    - docs/**/*.md

checks:
  code_blocks:
    enabled: true
    languages:
      - javascript
      - python
    timeout: 30s
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config file: %v\n", err)
		return ExitRuntimeErr
	}

	fmt.Printf("Created %s\n", configPath)
	return ExitNoDrift
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "docs-drift.yml", "Path to config file")
	verbose := fs.Bool("verbose", false, "Show verbose output")
	changedOnly := fs.Bool("changed-only", false, "Only check changed markdown files")
	base := fs.String("base", "", "Base reference for changed-only mode (branch or commit)")
	parallel := fs.Bool("parallel", false, "Run checks in parallel")
	workers := fs.Int("workers", DefaultWorkers, "Number of parallel workers")

	if err := fs.Parse(args); err != nil {
		return ExitRuntimeErr
	}

	// Validate worker count
	if *workers <= 0 {
		fmt.Fprintf(os.Stderr, "Invalid worker count: %d (must be positive)\n", *workers)
		return ExitRuntimeErr
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		return ExitRuntimeErr
	}

	// Create drift checker
	checker := drift.NewChecker(cfg, *verbose)

	// Collect files to check
	var files []string
	if *changedOnly {
		files, err = collectChangedFiles(cfg.Docs.Paths, *base, *verbose)
	} else {
		files, err = collectFiles(cfg.Docs.Paths)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to collect files: %v\n", err)
		return ExitRuntimeErr
	}

	if len(files) == 0 {
		if *changedOnly {
			fmt.Println("No changed markdown files found to check")
		} else {
			fmt.Println("No markdown files found to check")
		}
		return ExitNoDrift
	}

	if *verbose {
		mode := ""
		if *changedOnly {
			mode = " (changed only)"
		}
		if *parallel {
			mode += fmt.Sprintf(" (parallel, %d workers)", *workers)
		}
		fmt.Printf("Checking %d file(s)%s...\n", len(files), mode)
	}

	// Run checks
	var results *drift.Results
	if *parallel {
		results, err = checker.CheckConcurrent(files, *workers)
	} else {
		results, err = checker.Check(files)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Check failed: %v\n", err)
		return ExitRuntimeErr
	}

	// Output results
	output.PrintResults(results, *verbose)

	if results.HasDrift() {
		return ExitDrift
	}

	return ExitNoDrift
}

func runVersion(version string) int {
	fmt.Printf("docs-drift version %s\n", version)
	return ExitNoDrift
}

// collectFiles expands glob patterns and returns unique file paths
func collectFiles(patterns []string) ([]string, error) {
	seen := make(map[string]bool)
	var files []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			absPath, err := filepath.Abs(match)
			if err != nil {
				return nil, err
			}

			if !seen[absPath] {
				seen[absPath] = true

				// Check if it's a markdown file
				if filepath.Ext(absPath) == ".md" {
					files = append(files, absPath)
				}
			}
		}
	}

	return files, nil
}

// collectChangedFiles returns only changed markdown files using git
func collectChangedFiles(patterns []string, base string, verbose bool) ([]string, error) {
	// Check if we're in a git repository
	if !git.IsGitRepository() {
		return nil, fmt.Errorf("--changed-only requires a git repository")
	}

	// Auto-detect base if not specified
	if base == "" {
		base = git.GetDefaultBranch()
		if verbose {
			fmt.Printf("Auto-detected base branch: %s\n", base)
		}
	}

	// Get changed files
	files, err := git.ChangedFiles(base, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	if verbose && len(files) > 0 {
		fmt.Println("Changed files:")
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}
	}

	return files, nil
}
