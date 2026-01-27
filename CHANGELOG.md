# Changelog

All notable changes to docs-drift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1] - 2026-01-27

### Added
- **EULA.md**: Comprehensive End User License Agreement for GitHub Marketplace
  - Explicit code execution notices and user acknowledgments
  - Security measures and limitations documentation
  - Data privacy guarantees (no collection, no transmission)
  - Satisfies GitHub Marketplace Terms requirements
- **SECURITY.md**: Security policy and responsible disclosure process
  - Intentional code execution explanation (by design)
  - Five layers of security protection documentation
  - Vulnerability reporting guidelines with response timeline
  - Best practices for secure usage
  - Known limitations and transparency commitments
- **Security Documentation in README**: Enhanced security section
  - Status badges (release, CI, license, Go Report Card)
  - Execution safety principles and best practices
  - Link to security reporting process

### Documentation
- Prepared repository for official GitHub Marketplace publishing
- Added legal protections against potential misunderstandings about intentional code execution
- Enhanced transparency and security-by-design communication

## [0.3.0] - 2026-01-25

### Added
- **Bash/Shell Runner**: Validate shell scripts with automatic shebang and error handling
  - Supports `bash`, `sh`, and `shell` language tags
  - Uses `set -e` for fail-fast behavior
  - Comprehensive error extraction for shell errors
- **Go Runner**: Validate Go code examples with smart auto-wrapping
  - Auto-wraps code without `package main` declaration
  - Smart `fmt` import detection and injection
  - Creates temporary Go modules for execution
  - Supports `go` and `golang` language tags
- **Ruby Runner**: Validate Ruby code examples
  - Standard `.rb` file execution
  - Enhanced error extraction for Ruby-specific errors
  - Supports `ruby` and `rb` language tags
- **JSON Output Format**: Structured output for tool integration
  - `--format json` flag generates structured JSON
  - Includes version, timestamp, summary statistics, and detailed failures
  - Perfect for CI/CD pipelines and automation
- **HTML Report Generation**: Self-contained HTML reports
  - `--format html` flag generates standalone reports with embedded CSS
  - Color-coded status indicators and collapsible sections
  - Proper HTML escaping for security
  - Suitable for CI artifacts and offline viewing

### Changed
- Updated configuration validation to support bash, go, and ruby languages
- Enhanced language normalization for new aliases (sh, shell, golang, rb)
- Improved error messages in configuration validation

### Documentation
- Updated README.md with all new language runners and output formats
- Added comprehensive examples for JSON and HTML output
- Updated supported languages table
- Updated GitHub Action version references to v0.3

## [0.2.0] - 2026-01-22

### Added
- **Git Integration**: Only check changed files for faster CI validation
  - `--changed-only` flag to check only modified markdown files
  - `--base <ref>` flag to specify base branch/commit (auto-detects main/master)
  - Full git repository detection and branch operations
  - Dramatically faster validation in CI pipelines for large repos
- **Parallel Execution**: Concurrent code block validation
  - `--parallel` flag enables concurrent processing
  - `--workers <n>` flag to configure number of workers (default: 4)
  - Worker pool pattern with thread-safe result collection
  - Benchmarks show 2-4x speedup on multi-core systems
- **Benchmark Suite**: Performance tracking and validation
  - Parser benchmarks for markdown parsing performance
  - Checker benchmarks comparing sequential vs parallel execution
  - Memory allocation profiling support

### Changed
- Optimized code block processing for large documentation repositories
- Improved error handling in concurrent execution mode

### Documentation
- Added performance optimization guide
- Added changed-only mode documentation
- Added GitHub Action optimization examples

## [0.1.0] - 2026-01-20

### Added
- **Core CLI**: Command-line interface with `init`, `check`, and `version` commands
- **Markdown Parser**: CommonMark-compliant code block extraction
  - Supports backtick (```) and tilde (~~~) fenced code blocks
  - Line number tracking for accurate error reporting
  - `docs-drift:skip` directive to skip specific code blocks
- **JavaScript Runner**: Validate JavaScript code with Node.js
  - Isolated process execution with restricted environment
  - Configurable timeout per code block
  - Comprehensive error extraction
  - Supports `javascript` and `js` language tags
- **Python Runner**: Validate Python code with Python 3
  - Isolated mode execution for security
  - Traceback parsing for clear error messages
  - Supports `python` and `py` language tags
- **YAML Configuration**: Flexible configuration system
  - Glob patterns for markdown file discovery
  - Per-language validation control
  - Configurable execution timeouts
- **Colored Output**: Terminal-friendly error reporting
  - Clear file:line:error format
  - Color-coded success/failure indicators
  - Summary statistics
- **GitHub Action**: CI/CD integration
  - Composite action for easy workflow integration
  - Configurable inputs for all CLI flags
  - Action outputs for drift detection status
- **Exit Codes**: Semantic exit codes for CI integration
  - `0`: No drift detected
  - `1`: Drift detected (code blocks failed)
  - `2`: Runtime or configuration error

### Documentation
- Comprehensive README.md with usage examples
- CONTRIBUTING.md with development guidelines
- CLAUDE.md for AI-assisted development context
- Example configuration file

### Testing
- 55+ comprehensive tests across all packages
- 80-95% code coverage
- CI pipeline with automated testing
- Race condition detection

[0.3.1]: https://github.com/georg-nikola/docs-drift/releases/tag/v0.3.1
[0.3.0]: https://github.com/georg-nikola/docs-drift/releases/tag/v0.3.0
[0.2.0]: https://github.com/georg-nikola/docs-drift/releases/tag/v0.2.0
[0.1.0]: https://github.com/georg-nikola/docs-drift/releases/tag/v0.1.0
