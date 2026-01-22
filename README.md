# docs-drift

CLI and GitHub Action that detects documentation drift by validating code examples in Markdown files.

## Why docs-drift?

Documentation gets out of sync with code. Code examples in README files break silently. Users copy-paste examples that no longer work.

**docs-drift** validates that code examples in your documentation actually run without errors.

## Features

- Validates JavaScript and Python code blocks in Markdown files
- Runs in CI/CD pipelines via GitHub Action
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
      - javascript
      - python
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

# Show version
docs-drift version
```

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

      - uses: georg-nikola/docs-drift@v0.1
        with:
          config: docs-drift.yml
```

### Action Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `config` | Path to config file | `docs-drift.yml` |
| `version` | docs-drift version | `latest` |
| `verbose` | Enable verbose output | `false` |

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

## Supported Languages

| Language | Runtime | Aliases |
|----------|---------|---------|
| JavaScript | Node.js | `javascript`, `js` |
| Python | Python 3 | `python`, `py` |

## Security

Code blocks are executed in isolated processes with:
- Restricted environment variables
- Execution timeouts
- Temporary file cleanup

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
