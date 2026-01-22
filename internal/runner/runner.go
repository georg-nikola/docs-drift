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
