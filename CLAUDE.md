# CLAUDE.md - Complete Project Context

This file provides comprehensive guidance to Claude Code when working with docs-drift. It contains all context needed to continue development, including references to original specifications and implementation details.

## Project Overview

**docs-drift** is an open-source CLI tool and GitHub Action that detects documentation drift by validating code examples in Markdown files. It fails CI when documentation examples break, ensuring docs stay in sync with code.

**Repository**: https://github.com/georg-nikola/docs-drift
**Current Version**: v0.2
**Status**: Production Ready ✅

### Core Value Proposition
- **Fast**: Local execution, no network calls
- **Deterministic**: Same input = same output
- **Simple**: Single binary, no runtime dependencies
- **Private**: No SaaS, no telemetry

### How It Works
docs-drift extracts code blocks from markdown files, executes them in isolated environments (Node.js for JavaScript, Python 3 for Python), and reports failures with clear file:line:error messages.

## Original Documentation Reference

**IMPORTANT**: Complete project specifications are located at:
```
~/Downloads/docs-drift-implementation-docs/
```

### Documentation Files (Always Reference These)
1. `01_PROJECT_OVERVIEW.md` - High-level goals, motivation, and principles
2. `02_ARCHITECTURE.md` - System design, components, and data flow
3. `03_CLI_SPEC.md` - Command-line interface specification
4. `04_CONFIG_SPEC.md` - Configuration file format and validation
5. `05_MARKDOWN_PARSING.md` - Code block extraction rules and edge cases
6. `06_LANGUAGE_RUNNERS.md` - Execution environment specifications
7. `07_GITHUB_ACTION.md` - GitHub Action implementation details
8. `08_ERROR_OUTPUT.md` - Error formatting and presentation standards
9. `09_ROADMAP.md` - Feature roadmap, versions, and priorities
10. `10_GO_NO_GO_METRICS.md` - Success metrics and evaluation criteria

**When implementing new features or making architectural decisions, always consult these specification files first.**

## Build and Development Commands

### Build
```bash
# Build the main binary
go build -o docs-drift ./cmd/docs-drift

# Build all packages
go build -v ./...

# Build with version info (ldflags)
go build -ldflags "-X main.Version=v0.1.0" -o docs-drift ./cmd/docs-drift
```

### Testing
```bash
# Run all tests
go test -v ./...

# Run tests with race detector and coverage
go test -v -race -coverprofile=coverage.out ./...

# Run tests for a specific package
go test -v ./internal/parser
go test -v ./pkg/drift
```

### Running the Tool
```bash
# Initialize config file
./docs-drift init

# Run drift check
./docs-drift check

# Run with custom config
./docs-drift check --config ./custom-config.yml

# Verbose output
./docs-drift check --verbose

# Check only changed files (Git mode)
./docs-drift check --changed-only
./docs-drift check --changed-only --base origin/main

# Parallel execution
./docs-drift check --parallel
./docs-drift check --parallel --workers 8
```

### Linting
```bash
# The project uses golangci-lint
golangci-lint run
```

### Benchmarking
```bash
# Run all benchmarks
go test -bench=. ./...

# Benchmark specific packages
go test -bench=. ./internal/parser
go test -bench=. ./pkg/drift

# Benchmark with memory allocation stats
go test -bench=. -benchmem ./...

# Compare parallel performance with different worker counts
go test -bench=BenchmarkChecker ./pkg/drift
```

## Architecture

### Package Structure

The codebase follows Go's standard project layout:

- **cmd/docs-drift** - Entry point that calls `cli.Run()`
- **internal/** - Private application code
  - **cli/** - Command-line interface, argument parsing, and exit code handling
  - **config/** - YAML config file parsing and validation
  - **parser/** - Markdown parsing to extract code blocks with regex-based fence detection
  - **runner/** - Code execution in isolated processes (creates temp files, runs node/python3)
  - **output/** - Result formatting and display
  - **git/** - Git integration for changed-only mode (repository detection, branch operations, changed file detection)
- **pkg/drift** - Public API exposing `Checker` type that orchestrates parsing and running

### Key Data Flow

1. **CLI** (`internal/cli`) parses arguments and loads config file via `config.Load()`
2. **Config** (`internal/config`) validates YAML structure and ensures required runtimes are available
3. **CLI** collects markdown files:
   - In normal mode: uses `filepath.Glob()` based on `docs.paths` patterns
   - In `--changed-only` mode: uses `git.ChangedFiles()` to get only modified markdown files since base ref
4. **Checker** (`pkg/drift`) creates a `Parser` and iterates through each file:
   - **Parser** (`internal/parser`) uses regex to find code fences (``` or ~~~), extracts language and code content
   - Blocks with `docs-drift:skip` directive are marked but not executed
   - Languages are normalized (js→javascript, py→python)
5. **Checker** creates a **Runner** (`internal/runner`) for each block's language:
   - Runner writes code to temp file in `os.TempDir()`
   - Executes via `exec.CommandContext` with timeout
   - Uses restricted environment (limited env vars, isolated mode flags)
   - Captures combined stdout/stderr
6. **Output** (`internal/output`) formats results showing file path, line numbers, and error messages
7. **CLI** returns appropriate exit code: 0 (no drift), 1 (drift detected), 2 (runtime error)

### Important Implementation Details

- **Exit Codes**: The tool has strict exit code semantics defined in `internal/cli/cli.go`
- **Timeout Handling**: Each code block execution has a configurable timeout (default 30s)
- **Security**: Code runs in isolated processes with restricted environments to limit network access
- **Language Support**: Currently supports JavaScript (via node) and Python (via python3). Adding new languages requires implementing the `Runner` interface in `internal/runner/runner.go`
- **Concurrent Execution**: `Checker.CheckConcurrent()` processes code blocks in parallel using worker goroutines. Enabled via `--parallel` flag with configurable worker count (default: 4)
- **Git Integration**: `internal/git` package provides repository detection, branch operations, and changed file detection for `--changed-only` mode
- **Error Extraction**: `runner.extractError()` attempts to parse error messages from runtime output to show users the most relevant line

### Testing Strategy

Each package has corresponding `_test.go` files:
- Config tests validate YAML parsing and error cases
- Parser tests verify fence detection, language extraction, and skip directive handling
- Runner tests check code execution, timeout behavior, and error extraction
- Checker tests ensure end-to-end validation works correctly
- CLI tests verify command routing and exit codes
- Git tests validate repository detection, branch operations, and changed file detection

Benchmark files (`*_bench_test.go`) track performance:
- Parser benchmarks measure markdown parsing performance
- Checker benchmarks compare sequential vs parallel execution with different worker counts

### Benchmark Suite

The project includes comprehensive benchmarks to track and validate performance:

**Parser Benchmarks** (`internal/parser/parser_bench_test.go`):
- Measures markdown parsing performance
- Tests various file sizes and complexity levels
- Validates regex-based fence detection efficiency

**Checker Benchmarks** (`pkg/drift/checker_bench_test.go`):
- Compares sequential vs parallel execution
- Tests different worker counts (1, 2, 4, 8 workers)
- Demonstrates 2-4x speedup with parallel mode on multi-core systems
- Includes memory allocation profiling

**Running Benchmarks:**
```bash
# Run all benchmarks
go test -bench=. ./...

# Run with memory stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkChecker ./pkg/drift

# Compare results (save baseline first)
go test -bench=. ./... > old.txt
# Make changes
go test -bench=. ./... > new.txt
benchcmp old.txt new.txt  # if benchcmp is installed
```

**Performance Insights from Benchmarks:**
- Parallel execution scales well up to CPU core count
- Diminishing returns beyond 8 workers on typical systems
- Optimal worker count: 4-8 for most documentation repositories
- Changed-only mode provides 10-100x speedup for incremental checks (depending on change size)

## Configuration

The tool requires a `docs-drift.yml` file:
- `version`: must be 1
- `docs.paths`: glob patterns for markdown files
- `checks.code_blocks.enabled`: whether to validate code blocks
- `checks.code_blocks.languages`: list of languages to validate (javascript/js, python/py)
- `checks.code_blocks.timeout`: execution timeout per block (e.g., "30s")

## Requirements

- Go 1.21+ for building
- Node.js for JavaScript validation
- Python 3 for Python validation

## Implementation Status

### v0.1 - COMPLETE ✅

All v0.1 features have been successfully implemented and tested:

#### Core Features Delivered
- ✅ CLI with `init`, `check`, and `version` commands
- ✅ Markdown parser (CommonMark-compliant)
- ✅ JavaScript runner (Node.js with sandboxing)
- ✅ Python runner (python3 with isolated mode)
- ✅ YAML configuration system with validation
- ✅ Colored error output with file:line tracking
- ✅ GitHub Action for CI/CD integration
- ✅ Comprehensive test suite (55 tests, 80-95% coverage)
- ✅ CI pipeline with automated testing
- ✅ Complete documentation (README, CONTRIBUTING)

#### Exit Codes
- `0` - No drift detected (all code blocks passed)
- `1` - Drift detected (one or more code blocks failed)
- `2` - Runtime or configuration error

### v0.2 - COMPLETE ✅

All v0.2 features have been successfully implemented and tested:

#### Performance and Git Integration Features Delivered
- ✅ **Git Integration** - Changed-only mode for faster CI
  - `internal/git` package with repository detection (`IsGitRepository`)
  - Branch operations (`GetCurrentBranch`, `GetDefaultBranch`)
  - Changed file detection (`ChangedFiles`) using `git diff`
  - CLI flags: `--changed-only` and `--base <ref>` (defaults to main/master)
  - Full test coverage for all git operations

- ✅ **Parallel Execution** - Significant performance improvements
  - Enabled `CheckConcurrent()` method in CLI
  - CLI flags: `--parallel` and `--workers <n>` (default: 4)
  - Worker pool pattern with configurable concurrency
  - 2-4x speedup on multi-core systems for large documentation sets
  - Thread-safe result collection and error handling

- ✅ **Benchmark Suite** - Performance tracking and validation
  - Parser benchmarks (`internal/parser/parser_bench_test.go`)
  - Checker benchmarks (`pkg/drift/checker_bench_test.go`)
  - Worker count comparison benchmarks (1, 2, 4, 8 workers)
  - Memory allocation profiling support
  - Demonstrates measurable performance improvements with parallel execution

## Roadmap & Next Steps

Reference `~/Downloads/docs-drift-implementation-docs/09_ROADMAP.md` for the complete roadmap.

### v0.3 - Planned Features (Next)

1. **Additional language runners**
   - Shell/Bash (`#!/bin/bash` or `bash` language tag)
   - Ruby (`ruby` command)
   - PHP (`php` command)
   - Go itself (`go run` with temp module)

2. **Result caching**
   - Cache parse results to avoid re-parsing unchanged files
   - Track file hashes for incremental validation
   - Persistent cache across runs for faster local development

3. **Enhanced error reporting**
   - HTML report generation for CI artifacts
   - JSON output format for tool integration
   - Detailed statistics and summaries

### v0.4+ - Future Ideas
- Watch mode for local development (`--watch` flag)
- Custom runner commands in config (specify arbitrary executables)
- IDE integrations (VS Code extension for live feedback)
- Docker support for complex multi-language environments
- Smart retry logic for flaky tests
- Code coverage tracking for documentation examples

## Success Metrics

From `~/Downloads/docs-drift-implementation-docs/10_GO_NO_GO_METRICS.md`:

**30 days target:**
- 50 GitHub stars
- 10 repositories using docs-drift

**60 days target:**
- 150 GitHub stars
- 25 repositories using docs-drift

**Decision criteria**: If metrics are unmet, archive the project or pivot strategy.

## Development Guidelines

### Code Style
- Follow standard Go conventions (`gofmt`, `golint`)
- Keep functions small and focused (single responsibility)
- Write tests for all new functionality
- Use meaningful variable and function names
- Add comments only for non-obvious logic
- Prefer explicit error handling over panics

### Testing Strategy
- Unit tests for each component (`*_test.go` files)
- Integration tests in `pkg/drift/checker_test.go`
- Test both success and error cases
- Test edge conditions (empty files, malformed markdown, etc.)
- Aim for >80% code coverage on core packages
- Use table-driven tests where appropriate

### Git Workflow
- Create feature branches for new work (`git checkout -b feature/name`)
- Write descriptive commit messages (conventional commits preferred)
- Push to GitHub regularly to trigger CI
- CI pipeline automatically runs tests on all commits
- Keep commits focused and atomic

## Common Development Tasks

### Adding a New Language Runner

To add support for a new programming language:

1. **Update Runner Interface** (`internal/runner/runner.go`):
   ```go
   func (r *Runner) runNewLanguage(code string) error {
       // Create temp file with appropriate extension
       // Execute with language-specific command
       // Apply timeout and environment restrictions
       // Extract and return errors
   }
   ```

2. **Add to Switch Statement** in `Run()` method:
   ```go
   case "newlang":
       return r.runNewLanguage(block.Code)
   ```

3. **Add Language Normalization** in `normalizeLanguage()`:
   ```go
   case "nl": // add alias
       return "newlang"
   ```

4. **Write Tests** (`internal/runner/runner_test.go`):
   ```go
   func TestRunNewLanguage(t *testing.T) {
       // Test successful execution
       // Test error cases
       // Test timeout behavior
   }
   ```

5. **Update Documentation**:
   - Add to README.md supported languages list
   - Update example config in `docs-drift.yml`
   - Add examples to test suite

6. **Check Requirements**:
   - Update `internal/config/config.go` validation to check for runtime
   - Add installation instructions to README

### Modifying the Parser

The parser (`internal/parser/parser.go`) follows CommonMark specification:
- Fenced code blocks start with 3+ backticks (```) or tildes (~~~)
- Language tag immediately follows opening fence (no spaces)
- Closing fence must use same character and be at least as long
- Line numbers track original file positions (critical for error reporting)

**When modifying the parser:**
- Maintain line number accuracy (used in error messages)
- Test with various markdown edge cases (nested blocks, info strings, etc.)
- Ensure `docs-drift:skip` directive continues working
- Verify behavior matches CommonMark spec
- Reference `~/Downloads/docs-drift-implementation-docs/05_MARKDOWN_PARSING.md`

### Configuration Changes

When adding new configuration options:

1. **Update Struct** (`internal/config/config.go`):
   ```go
   type Config struct {
       NewOption string `yaml:"new_option"`
   }
   ```

2. **Add Validation** in `Validate()` method:
   ```go
   if c.NewOption == "" {
       return fmt.Errorf("new_option is required")
   }
   ```

3. **Update Example Config** (`docs-drift.yml`)

4. **Document in README**

5. **Add Tests** for new validation logic

6. **Reference Spec**: Check `~/Downloads/docs-drift-implementation-docs/04_CONFIG_SPEC.md`

### Error Output Formatting

When adding new error types or modifying output:

1. Follow the standard format: `file:line: description`
2. Use colors appropriately (red for errors, yellow for warnings, green for success)
3. Include actionable information (what failed, why, how to fix)
4. Reference `~/Downloads/docs-drift-implementation-docs/08_ERROR_OUTPUT.md`
5. Test output in both color and no-color modes

### Using Git Integration

The `internal/git` package provides Git operations for changed-only mode:

**Key Functions:**
- `IsGitRepository(dir)` - Checks if a directory is a Git repository
- `GetCurrentBranch(dir)` - Returns the current branch name
- `GetDefaultBranch(dir)` - Auto-detects main or master branch
- `ChangedFiles(dir, base, patterns)` - Returns changed markdown files since base ref

**Example Usage in CLI:**
```go
// Check if we're in a git repo
isRepo, err := git.IsGitRepository(".")
if err != nil {
    return err
}

// Get changed files
changedFiles, err := git.ChangedFiles(".", baseRef, patterns)
if err != nil {
    return fmt.Errorf("failed to get changed files: %w", err)
}
```

**Testing Git Integration:**
```bash
# Create test repo
mkdir test-repo && cd test-repo
git init
echo '```javascript\nconsole.log("test");\n```' > README.md
git add . && git commit -m "Initial"

# Modify file
echo '```javascript\nconsole.log("changed");\n```' >> README.md

# Test changed-only mode
docs-drift check --changed-only --verbose
```

**Important Notes:**
- `--changed-only` requires being in a Git repository (exits with code 2 if not)
- Base ref defaults to the default branch (main or master)
- Only markdown files matching config patterns are included
- Uses `git diff --name-only` to detect changes efficiently
- Handles renamed files correctly (tracks as modified)

## Debugging Tips

### Enable Verbose Output
```bash
./docs-drift check --verbose
# Shows detailed information about:
# - Files being scanned
# - Code blocks found
# - Execution results
# - Timing information
```

### View Raw Parse Results
Add temporary debug logging in `internal/parser/parser.go`:
```go
fmt.Printf("DEBUG: Found code block: %+v\n", block)
```

### Test Individual Components
```bash
# Test just the parser
go test -v ./internal/parser -run TestParseFencedCodeBlock

# Test just the runner
go test -v ./internal/runner -run TestRunJavaScript

# Test git integration
go test -v ./internal/git

# Test with race detector (important for parallel execution)
go test -race -v ./...
```

### Check File Glob Matching
```bash
# Create test files and verify pattern matching
cd /path/to/test/docs
docs-drift check --verbose
# Verbose output shows which files matched glob patterns
```

### Test Timeout Behavior
```bash
# Create a markdown file with infinite loop
echo '```javascript
while(true) {}
```' > test.md

# Should timeout after configured duration
./docs-drift check
```

### Test Parallel Execution
```bash
# Create multiple markdown files with code blocks
for i in {1..10}; do
  echo '```javascript
console.log("test");
```' > "test${i}.md"
done

# Compare sequential vs parallel performance
time ./docs-drift check
time ./docs-drift check --parallel --workers 4
```

### Test Changed-Only Mode
```bash
# Must be in a git repository
git init
git add .
git commit -m "Initial commit"

# Modify a file
echo '```javascript
console.log("modified");
```' >> README.md

# Check only changed files
./docs-drift check --changed-only --verbose
# Should only process README.md

# Check against specific base ref
./docs-drift check --changed-only --base HEAD~1
```

## Important Implementation Notes

### Sandboxing and Security
- Runners execute code in separate OS processes (not in-memory)
- Environment variables are restricted to minimal set
- No network access enforced via process environment settings
- Timeouts prevent runaway processes
- Temp files are cleaned up after execution
- **Future**: Consider Docker containers for stronger isolation

### Performance Characteristics
- **Sequential mode** (default): processes files and code blocks one at a time
- **Parallel mode** (`--parallel` flag): processes code blocks concurrently with worker pool
  - Default: 4 workers (configurable via `--workers` flag)
  - Benchmarks show 2-4x speedup on multi-core systems
  - Best for large repos with many code blocks
  - Worker count should typically match CPU core count
- **Changed-only mode** (`--changed-only` flag): processes only modified files
  - Uses Git to detect changes since base ref
  - Dramatically faster for incremental CI builds
  - Ideal for PR checks in large documentation repositories
- Markdown parsing is fast (regex-based, minimal overhead)
- Bottleneck is usually code execution time (subprocess overhead and runtime duration)
- Combining `--parallel` and `--changed-only` provides maximum performance

### GitHub Action Integration
- Binary is downloaded from GitHub releases
- Action caches binary for faster subsequent runs
- Uses composite action pattern (shell + actions toolkit)
- Exit codes properly propagate to workflow status
- Supports both public and private repositories

### Parallel Execution
- `pkg/drift/Checker.CheckConcurrent()` is used when `--parallel` flag is set
- Uses worker pool pattern with configurable concurrency via `--workers` flag
- Default worker count: 4 (optimal for most systems)
- Collects results from multiple goroutines safely using channels
- Each worker processes code blocks independently
- Thread-safe error collection and result aggregation
- Benchmarks available in `pkg/drift/checker_bench_test.go`
- Performance scales well up to CPU core count, diminishing returns beyond

## Getting Started (For New Claude Sessions)

When starting a new session to work on docs-drift:

1. **Read this file completely** to understand project context

2. **Review relevant original docs** from `~/Downloads/docs-drift-implementation-docs/`:
   - For architecture changes → `02_ARCHITECTURE.md`
   - For CLI changes → `03_CLI_SPEC.md`
   - For config changes → `04_CONFIG_SPEC.md`
   - For new runners → `06_LANGUAGE_RUNNERS.md`
   - For GitHub Action → `07_GITHUB_ACTION.md`

3. **Check current state**:
   ```bash
   cd /Users/geoko/repos/docs-drift
   git status
   git log --oneline -5
   go test ./...
   ```

4. **Identify the task** - What feature or fix are we working on?

5. **Plan the implementation** - Use TodoWrite tool to break down work into steps

6. **Build, test, review continuously** - Validate changes frequently

## Questions to Ask When Unclear

- "What does the original spec say about X?" → Check files in `~/Downloads/docs-drift-implementation-docs/`
- "How should errors be formatted?" → See `08_ERROR_OUTPUT.md`
- "What's the priority order of features?" → See `09_ROADMAP.md`
- "How should the config be structured?" → See `04_CONFIG_SPEC.md`
- "What's the expected behavior for edge case Y?" → See `05_MARKDOWN_PARSING.md`
- "How should the GitHub Action work?" → See `07_GITHUB_ACTION.md`

## Useful Commands for Claude

### Read Original Specifications
```bash
# List all spec files
ls -la ~/Downloads/docs-drift-implementation-docs/

# Read a specific spec
cat ~/Downloads/docs-drift-implementation-docs/02_ARCHITECTURE.md

# Search for specific content
grep -r "timeout" ~/Downloads/docs-drift-implementation-docs/
```

### Check Implementation Status
```bash
cd /Users/geoko/repos/docs-drift

# Run all tests with verbose output
go test -v ./...

# Check test coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check for race conditions (critical for parallel execution)
go test -race ./...

# Run benchmarks
go test -bench=. ./...
go test -bench=. -benchmem ./pkg/drift

# Run linter (if golangci-lint installed)
golangci-lint run
```

### Working with Git
```bash
cd /Users/geoko/repos/docs-drift

# Check current status
git status

# View recent commits
git log --oneline -10

# View changes since last commit
git diff

# Create feature branch
git checkout -b feature/new-runner

# Push to remote
git push origin feature/new-runner
```

### Testing Local Changes
```bash
cd /Users/geoko/repos/docs-drift

# Build fresh binary
go build -o docs-drift ./cmd/docs-drift

# Test on sample markdown
./docs-drift init
./docs-drift check --verbose

# Test specific markdown file
echo '```javascript
console.log("test");
```' > test.md
./docs-drift check

# Test parallel execution
./docs-drift check --parallel --workers 4

# Test changed-only mode (must be in git repo)
git init
git add . && git commit -m "test"
echo '```javascript\ntest\n```' >> test.md
./docs-drift check --changed-only
```

## Project Metadata

- **Repository**: https://github.com/georg-nikola/docs-drift
- **License**: MIT (see LICENSE file)
- **Language**: Go 1.21+
- **CI**: GitHub Actions (`.github/workflows/ci.yml`)
- **Test Framework**: Go's built-in testing package
- **Dependencies**: Minimal (see `go.mod`)

---

**Last Updated**: 2026-01-22
**Current Version**: v0.2
**Next Version**: v0.3 (in planning)

*Keep this file updated as the project evolves. It serves as the single source of truth for Claude Code sessions.*
