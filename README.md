# docs-drift

CLI and GitHub Action that detects documentation drift by validating code examples in Markdown files.

## Why docs-drift?

Documentation gets out of sync with code. Code examples in README files break silently. Users copy-paste examples that no longer work.

**docs-drift** validates that code examples in your documentation actually run without errors.

## Features

- Validates JavaScript, Python, Bash, Go, and Ruby code blocks in Markdown files
- **Multiple output formats**: Text, JSON, and HTML reports
- Runs in CI/CD pipelines via GitHub Action
- **Git integration**: Only check changed files in CI for faster validation
- **Parallel execution**: Concurrent code block checking for improved performance
- Fast, local-first, deterministic
- No SaaS, no telemetry
- Skip specific code blocks with `docs-drift:skip` directive

## Installation

### Using Go

```bash
go install github.com/georg-nikola/docs-drift/cmd/docs-drift@latest
```

### From Source

```bash
git clone https://github.com/georg-nikola/docs-drift.git
cd docs-drift
go build -o docs-drift ./cmd/docs-drift
```

## Quick Start

1. Initialize a config file:

```bash
docs-drift init
```

2. Run the check:

```bash
docs-drift check
```

## Configuration

Create a `docs-drift.yml` file in your project root:

```yaml
version: 1

docs:
  paths:
    - README.md
    - docs/**/*.md

checks:
  code_blocks:
    enabled: true
    languages:
      - javascript  # Node.js required
      - python      # Python 3 required
      - bash        # Bash required
      - go          # Go required
      - ruby        # Ruby required
    timeout: 30s
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `version` | Config file version (must be 1) | Required |
| `docs.paths` | Glob patterns for Markdown files | Required |
| `checks.code_blocks.enabled` | Enable code block validation | `true` |
| `checks.code_blocks.languages` | Languages to validate | Required |
| `checks.code_blocks.timeout` | Execution timeout per block | `30s` |

## Usage

### CLI Commands

```bash
# Create config file
docs-drift init

# Validate code blocks
docs-drift check

# Use custom config file
docs-drift check --config ./custom-config.yml

# Verbose output
docs-drift check --verbose

# Only check changed files (requires git repository)
docs-drift check --changed-only

# Check changes compared to specific branch/commit
docs-drift check --changed-only --base origin/main

# Enable parallel execution for faster validation
docs-drift check --parallel

# Specify number of concurrent workers (default: 4)
docs-drift check --parallel --workers 8

# Combine flags for optimal CI performance
docs-drift check --changed-only --parallel --workers 4

# Generate JSON output
docs-drift check --format json

# Generate HTML report
docs-drift check --format html

# Show version
docs-drift version
```

### Output Formats

docs-drift supports multiple output formats to integrate with different workflows:

#### Text (Default)

Human-readable colored output for terminal use:

```
Docs Drift Detected

README.md
  Line 42: javascript code block
    -> ReferenceError: undefinedVar is not defined

Summary: 1 failed, 5 passed
```

#### JSON

Structured output for tool integration and automation:

```bash
docs-drift check --format json
```

```json
{
  "version": "1",
  "timestamp": "2026-01-25T10:30:00Z",
  "summary": {
    "total": 6,
    "passed": 5,
    "failed": 1,
    "skipped": 0
  },
  "failures": [
    {
      "file": "README.md",
      "line": 42,
      "language": "javascript",
      "error": "ReferenceError: undefinedVar is not defined"
    }
  ]
}
```

#### HTML

Self-contained HTML report with embedded CSS for CI artifacts:

```bash
docs-drift check --format html > report.html
```

The HTML report includes:
- Summary statistics with color-coded status
- Detailed failure information with file paths and line numbers
- Collapsible sections for easy navigation
- Embedded CSS for offline viewing

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No drift detected |
| 1 | Drift detected (code blocks failed) |
| 2 | Runtime or configuration error |

### Skipping Code Blocks

Add `docs-drift:skip` to the code fence to skip validation:

~~~markdown
```javascript docs-drift:skip
// This code won't be validated
const example = "skipped";
```
~~~

### Changed-only Mode

The `--changed-only` flag uses git to only validate code blocks in modified Markdown files, significantly improving performance in CI pipelines for large documentation repositories.

```bash
# Check only files changed since default branch (auto-detects main/master)
docs-drift check --changed-only

# Check files changed compared to a specific branch
docs-drift check --changed-only --base origin/develop

# Check files changed compared to a specific commit
docs-drift check --changed-only --base abc123f
```

**Requirements**:
- Must be run inside a git repository
- Base branch/commit must exist in the repository
- If `--base` is not specified, automatically detects the default branch (main or master)

**Use cases**:
- Pull request validation (only check docs modified in the PR)
- Incremental validation in large documentation repositories
- Faster feedback loops during development

### Performance Optimization

The `--parallel` flag enables concurrent execution of code block validation, providing significant speedup on multi-core systems.

```bash
# Enable parallel execution with default workers (4)
docs-drift check --parallel

# Specify custom number of workers
docs-drift check --parallel --workers 8
```

**Performance characteristics**:
- **Sequential (default)**: Processes code blocks one at a time
- **Parallel**: Benchmarks show ~2-4x speedup depending on worker count
- **Optimal worker count**: Typically matches CPU core count
- **Minimum workers**: Must be at least 1

**Best practices**:
- Use `--parallel` for repositories with many code blocks (10+)
- Combine with `--changed-only` for maximum CI performance
- Start with default workers (4) and adjust based on benchmarks
- For very large repositories, consider using both flags together:
  ```bash
  docs-drift check --changed-only --parallel --workers $(nproc)
  ```

## GitHub Action

Add to your workflow:

```yaml
name: Documentation Check

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  docs-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for --changed-only mode

      - uses: georg-nikola/docs-drift@v0.3
        with:
          config: docs-drift.yml
```

### Optimized CI Configuration

For pull requests, use `--changed-only` to validate only modified documentation:

```yaml
name: Documentation Check

on:
  pull_request:
    branches: [main]

jobs:
  docs-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch all history for git comparison

      - uses: georg-nikola/docs-drift@v0.3
        with:
          config: docs-drift.yml
          changed-only: true
          base: origin/main
          parallel: true
          workers: 4
```

### Action Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `config` | Path to config file | `docs-drift.yml` |
| `version` | docs-drift version | `latest` |
| `verbose` | Enable verbose output | `false` |
| `changed-only` | Only check changed files (requires git) | `false` |
| `base` | Base branch/commit for comparison | Auto-detect |
| `parallel` | Enable parallel execution | `false` |
| `workers` | Number of concurrent workers | `4` |
| `format` | Output format (text, json, html) | `text` |

### Action Outputs

| Output | Description |
|--------|-------------|
| `drift-detected` | `true` if drift was found |
| `total-blocks` | Number of code blocks checked |
| `failed-blocks` | Number of failed code blocks |

## Error Output

When drift is detected, docs-drift shows clear error messages:

```
Docs Drift Detected

README.md
  Line 42: javascript code block
    -> ReferenceError: undefinedVar is not defined

  Line 78: python code block
    -> NameError: name 'undefined_var' is not defined

Summary: 2 failed, 5 passed
```

## Requirements

- Go 1.21+ (for building)
- Node.js (for JavaScript validation)
- Python 3 (for Python validation)
- Bash (for Bash/Shell validation)
- Ruby (for Ruby validation)

**Note**: Only the runtimes for languages specified in your configuration are required. For example, if you only validate JavaScript and Python, you don't need Bash, Go, or Ruby installed.

## Supported Languages

| Language | Runtime | Aliases |
|----------|---------|---------|
| JavaScript | Node.js | `javascript`, `js` |
| Python | Python 3 | `python`, `py` |
| Bash/Shell | bash | `bash`, `sh`, `shell` |
| Go | go | `go`, `golang` |
| Ruby | ruby | `ruby`, `rb` |

## Security

Code blocks are executed in isolated processes with:
- Restricted environment variables
- Execution timeouts
- Temporary file cleanup

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
