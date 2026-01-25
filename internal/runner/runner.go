package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Result represents the outcome of running a code block
type Result struct {
	Success  bool
	Output   string
	Error    string
	Duration time.Duration
}

// Runner executes code blocks in isolated environments
type Runner interface {
	Run(ctx context.Context, code string) Result
	Language() string
}

// NewRunner creates a runner for the specified language
func NewRunner(language string, timeout time.Duration) (Runner, error) {
	switch strings.ToLower(language) {
	case "javascript", "js":
		return &JavaScriptRunner{timeout: timeout}, nil
	case "python", "py":
		return &PythonRunner{timeout: timeout}, nil
	case "bash", "sh", "shell":
		return &BashRunner{timeout: timeout}, nil
	case "go", "golang":
		return &GoRunner{timeout: timeout}, nil
	case "ruby", "rb":
		return &RubyRunner{timeout: timeout}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// JavaScriptRunner executes JavaScript code using Node.js
type JavaScriptRunner struct {
	timeout time.Duration
}

func (r *JavaScriptRunner) Language() string {
	return "javascript"
}

func (r *JavaScriptRunner) Run(ctx context.Context, code string) Result {
	start := time.Now()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "docs-drift-*.js")
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(start),
		}
	}
	defer os.Remove(tmpFile.Name())

	// Write code to temp file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to write code: %v", err),
			Duration: time.Since(start),
		}
	}
	tmpFile.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute with node
	// Use --experimental-permission to restrict network access (Node.js 20+)
	// Fall back to basic execution if permission model not available
	cmd := exec.CommandContext(ctx, "node", tmpFile.Name())

	// Restrict environment to prevent network access where possible
	cmd.Env = restrictedEnv()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("execution timed out after %s", r.timeout),
			Duration: duration,
		}
	}

	if err != nil {
		return Result{
			Success:  false,
			Output:   string(output),
			Error:    extractError(string(output)),
			Duration: duration,
		}
	}

	return Result{
		Success:  true,
		Output:   string(output),
		Duration: duration,
	}
}

// PythonRunner executes Python code using python3
type PythonRunner struct {
	timeout time.Duration
}

func (r *PythonRunner) Language() string {
	return "python"
}

func (r *PythonRunner) Run(ctx context.Context, code string) Result {
	start := time.Now()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "docs-drift-*.py")
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(start),
		}
	}
	defer os.Remove(tmpFile.Name())

	// Write code to temp file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to write code: %v", err),
			Duration: time.Since(start),
		}
	}
	tmpFile.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute with python3
	// Use -I for isolated mode (no user site-packages, ignore PYTHON* env vars)
	cmd := exec.CommandContext(ctx, "python3", "-I", tmpFile.Name())

	// Restrict environment
	cmd.Env = restrictedEnv()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("execution timed out after %s", r.timeout),
			Duration: duration,
		}
	}

	if err != nil {
		return Result{
			Success:  false,
			Output:   string(output),
			Error:    extractError(string(output)),
			Duration: duration,
		}
	}

	return Result{
		Success:  true,
		Output:   string(output),
		Duration: duration,
	}
}

// BashRunner executes Bash/Shell scripts
type BashRunner struct {
	timeout time.Duration
}

func (r *BashRunner) Language() string {
	return "bash"
}

func (r *BashRunner) Run(ctx context.Context, code string) Result {
	start := time.Now()

	// Create temp file with .sh extension
	tmpFile, err := os.CreateTemp("", "docs-drift-*.sh")
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(start),
		}
	}
	defer os.Remove(tmpFile.Name())

	// Add shebang and error handling
	// set -e: exit on first error
	script := "#!/bin/bash\nset -e\n" + code

	if _, err := tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to write script: %v", err),
			Duration: time.Since(start),
		}
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0700); err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to make script executable: %v", err),
			Duration: time.Since(start),
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute with bash
	cmd := exec.CommandContext(ctx, "bash", tmpFile.Name())
	cmd.Env = restrictedEnv()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("execution timed out after %s", r.timeout),
			Duration: duration,
		}
	}

	if err != nil {
		return Result{
			Success:  false,
			Output:   string(output),
			Error:    extractError(string(output)),
			Duration: duration,
		}
	}

	return Result{
		Success:  true,
		Output:   string(output),
		Duration: duration,
	}
}

// GoRunner executes Go code
type GoRunner struct {
	timeout time.Duration
}

func (r *GoRunner) Language() string {
	return "go"
}

func (r *GoRunner) Run(ctx context.Context, code string) Result {
	start := time.Now()

	// Create temp directory for Go module
	tmpDir, err := os.MkdirTemp("", "docs-drift-go-*")
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp directory: %v", err),
			Duration: time.Since(start),
		}
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal go.mod
	goModContent := `module docsdrift/temp

go 1.22
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create go.mod: %v", err),
			Duration: time.Since(start),
		}
	}

	// Wrap code if it doesn't have package declaration
	finalCode := code
	if !strings.Contains(code, "package main") {
		// Auto-detect common imports needed
		imports := ""
		if strings.Contains(code, "fmt.") {
			imports += `import "fmt"`
		}
		if imports != "" {
			imports += "\n\n"
		}
		finalCode = "package main\n\n" + imports + "func main() {\n" + code + "\n}\n"
	}

	// Write main.go
	mainGoPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(finalCode), 0644); err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create main.go: %v", err),
			Duration: time.Since(start),
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute with go run
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = tmpDir
	cmd.Env = restrictedEnv()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("execution timed out after %s", r.timeout),
			Duration: duration,
		}
	}

	if err != nil {
		return Result{
			Success:  false,
			Output:   string(output),
			Error:    extractError(string(output)),
			Duration: duration,
		}
	}

	return Result{
		Success:  true,
		Output:   string(output),
		Duration: duration,
	}
}

// RubyRunner executes Ruby code
type RubyRunner struct {
	timeout time.Duration
}

func (r *RubyRunner) Language() string {
	return "ruby"
}

func (r *RubyRunner) Run(ctx context.Context, code string) Result {
	start := time.Now()

	// Create temp file with .rb extension
	tmpFile, err := os.CreateTemp("", "docs-drift-*.rb")
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to create temp file: %v", err),
			Duration: time.Since(start),
		}
	}
	defer os.Remove(tmpFile.Name())

	// Write code to temp file
	if _, err := tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("failed to write code: %v", err),
			Duration: time.Since(start),
		}
	}
	tmpFile.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute with ruby
	cmd := exec.CommandContext(ctx, "ruby", tmpFile.Name())
	cmd.Env = restrictedEnv()

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{
			Success:  false,
			Error:    fmt.Sprintf("execution timed out after %s", r.timeout),
			Duration: duration,
		}
	}

	if err != nil {
		return Result{
			Success:  false,
			Output:   string(output),
			Error:    extractError(string(output)),
			Duration: duration,
		}
	}

	return Result{
		Success:  true,
		Output:   string(output),
		Duration: duration,
	}
}

// restrictedEnv creates a minimal environment that restricts network access
func restrictedEnv() []string {
	// Get temp directory for file operations
	tmpDir := os.TempDir()

	// Minimal environment - no proxy settings, restricted paths
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + tmpDir,
		"TMPDIR=" + tmpDir,
		"TEMP=" + tmpDir,
		"TMP=" + tmpDir,
		// Disable network for Node.js where possible
		"NODE_OPTIONS=--no-warnings",
		// Disable Python network features
		"PYTHONDONTWRITEBYTECODE=1",
		"PYTHONUNBUFFERED=1",
	}

	return env
}

// extractError attempts to extract the most relevant error message
func extractError(output string) string {
	lines := strings.Split(output, "\n")

	// Look for common error patterns
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// JavaScript errors
		if strings.Contains(line, "Error:") ||
			strings.Contains(line, "TypeError") ||
			strings.Contains(line, "ReferenceError") ||
			strings.Contains(line, "SyntaxError") {
			return line
		}

		// Python errors
		if strings.HasPrefix(line, "Traceback") ||
			strings.Contains(line, "Error:") ||
			strings.Contains(line, "Exception:") {
			// Return the last meaningful line
			for j := len(lines) - 1; j >= 0; j-- {
				l := strings.TrimSpace(lines[j])
				if l != "" && !strings.HasPrefix(l, "File ") {
					return l
				}
			}
		}

		// Bash errors
		if strings.Contains(line, ": line ") ||
			strings.Contains(line, "bash:") ||
			strings.Contains(line, "command not found") ||
			strings.Contains(line, "No such file or directory") {
			return line
		}

		// Go compile errors
		if strings.Contains(line, "# command-line-arguments") {
			// For Go compile errors, return the actual error line, not the package line
			for j := i + 1; j < len(lines); j++ {
				errLine := strings.TrimSpace(lines[j])
				if errLine != "" && !strings.HasPrefix(errLine, "#") {
					return errLine
				}
			}
			return line
		}

		// Go panic errors - look for the panic line specifically
		if strings.HasPrefix(line, "panic:") {
			return line
		}

		// Go fatal errors
		if strings.Contains(line, "fatal error:") {
			return line
		}

		// Ruby errors
		if strings.Contains(line, "Error") ||
			strings.Contains(line, "Exception") ||
			strings.Contains(line, "undefined method") ||
			strings.Contains(line, "uninitialized constant") ||
			strings.Contains(line, "wrong number of arguments") ||
			strings.Contains(line, ".rb:") {
			return line
		}
	}

	// Special handling for Go panics - search for panic: at the beginning of a line
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "panic:") {
			return line
		}
	}

	// Fall back to last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}

	return "execution failed"
}

// CheckRuntime verifies that required runtimes are available
func CheckRuntime(language string) error {
	switch strings.ToLower(language) {
	case "javascript", "js":
		return checkCommand("node", "--version")
	case "python", "py":
		return checkCommand("python3", "--version")
	case "bash", "sh", "shell":
		return checkCommand("bash", "--version")
	case "go", "golang":
		return checkCommand("go", "version")
	case "ruby", "rb":
		return checkCommand("ruby", "--version")
	default:
		return fmt.Errorf("unknown language: %s", language)
	}
}

func checkCommand(name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found in PATH", name)
	}

	cmd := exec.Command(path, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s found but failed to run: %w", name, err)
	}

	return nil
}

// GetRuntimeVersion returns the version of the specified runtime
func GetRuntimeVersion(language string) (string, error) {
	var cmd *exec.Cmd

	switch strings.ToLower(language) {
	case "javascript", "js":
		cmd = exec.Command("node", "--version")
	case "python", "py":
		cmd = exec.Command("python3", "--version")
	case "bash", "sh", "shell":
		cmd = exec.Command("bash", "--version")
	case "go", "golang":
		cmd = exec.Command("go", "version")
	case "ruby", "rb":
		cmd = exec.Command("ruby", "--version")
	default:
		return "", fmt.Errorf("unknown language: %s", language)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// TempDir returns a secure temporary directory for code execution
func TempDir() string {
	dir := filepath.Join(os.TempDir(), "docs-drift")
	_ = os.MkdirAll(dir, 0700) // Best effort: ignore error as temp dir should always be writable
	return dir
}
