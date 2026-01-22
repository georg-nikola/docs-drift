# Contributing to docs-drift

Thank you for your interest in contributing to docs-drift!

## Development Setup

### Prerequisites

- Go 1.21 or later
- Node.js (for running JavaScript tests)
- Python 3 (for running Python tests)

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/docs-drift.git
   cd docs-drift
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## Project Structure

```
docs-drift/
├── cmd/docs-drift/     # CLI entry point
├── internal/           # Private packages
│   ├── cli/           # CLI command handling
│   ├── config/        # Configuration loading
│   ├── output/        # Output formatting
│   ├── parser/        # Markdown parsing
│   └── runner/        # Language runners
├── pkg/drift/         # Public API
├── testdata/          # Test fixtures
├── action.yml         # GitHub Action definition
└── docs-drift.yml     # Example config
```

## Making Changes

### Code Style

- Follow standard Go conventions
- Run `go fmt ./...` before committing
- Run `go vet ./...` to check for issues

### Testing

All changes should include tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Building

```bash
# Build the binary
go build -o docs-drift ./cmd/docs-drift

# Build with version info
go build -ldflags "-X main.Version=dev" -o docs-drift ./cmd/docs-drift
```

## Pull Request Process

1. Create a feature branch from `main`
2. Make your changes
3. Add or update tests as needed
4. Ensure all tests pass
5. Update documentation if applicable
6. Submit a pull request

### PR Guidelines

- Keep PRs focused on a single change
- Write clear commit messages
- Reference any related issues
- Update the CHANGELOG if applicable

## Reporting Issues

When reporting bugs, please include:

- docs-drift version (`docs-drift version`)
- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior

## Feature Requests

Feature requests are welcome! Please:

- Check existing issues first
- Describe the use case
- Explain why it would be valuable

## Code of Conduct

Be respectful and constructive in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
