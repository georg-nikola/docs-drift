package runner

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewRunner_JavaScript(t *testing.T) {
	r, err := NewRunner("javascript", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Language() != "javascript" {
		t.Errorf("expected javascript, got %s", r.Language())
	}
}

func TestNewRunner_Python(t *testing.T) {
	r, err := NewRunner("python", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Language() != "python" {
		t.Errorf("expected python, got %s", r.Language())
	}
}

func TestNewRunner_Bash(t *testing.T) {
	r, err := NewRunner("bash", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Language() != "bash" {
		t.Errorf("expected bash, got %s", r.Language())
	}
}

func TestNewRunner_Aliases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"js", "javascript"},
		{"py", "python"},
		{"bash", "bash"},
		{"sh", "bash"},
		{"shell", "bash"},
		{"go", "go"},
		{"golang", "go"},
		{"ruby", "ruby"},
		{"rb", "ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r, err := NewRunner(tt.input, 30*time.Second)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.Language() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, r.Language())
			}
		})
	}
}

func TestNewRunner_UnsupportedLanguage(t *testing.T) {
	_, err := NewRunner("rust", 30*time.Second)
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestJavaScriptRunner_Success(t *testing.T) {
	if err := CheckRuntime("javascript"); err != nil {
		t.Skip("Node.js not available:", err)
	}

	r := &JavaScriptRunner{timeout: 30 * time.Second}
	code := `console.log("Hello, World!");`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", result.Output)
	}
}

func TestJavaScriptRunner_SyntaxError(t *testing.T) {
	if err := CheckRuntime("javascript"); err != nil {
		t.Skip("Node.js not available:", err)
	}

	r := &JavaScriptRunner{timeout: 30 * time.Second}
	code := `console.log("unclosed string`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for syntax error")
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestJavaScriptRunner_RuntimeError(t *testing.T) {
	if err := CheckRuntime("javascript"); err != nil {
		t.Skip("Node.js not available:", err)
	}

	r := &JavaScriptRunner{timeout: 30 * time.Second}
	code := `throw new Error("test error");`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for runtime error")
	}

	if !strings.Contains(result.Error, "test error") {
		t.Errorf("expected error to contain 'test error', got %q", result.Error)
	}
}

func TestJavaScriptRunner_Timeout(t *testing.T) {
	if err := CheckRuntime("javascript"); err != nil {
		t.Skip("Node.js not available:", err)
	}

	r := &JavaScriptRunner{timeout: 100 * time.Millisecond}
	code := `while(true) {}`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for timeout")
	}

	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
}

func TestPythonRunner_Success(t *testing.T) {
	if err := CheckRuntime("python"); err != nil {
		t.Skip("Python not available:", err)
	}

	r := &PythonRunner{timeout: 30 * time.Second}
	code := `print("Hello, World!")`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", result.Output)
	}
}

func TestPythonRunner_SyntaxError(t *testing.T) {
	if err := CheckRuntime("python"); err != nil {
		t.Skip("Python not available:", err)
	}

	r := &PythonRunner{timeout: 30 * time.Second}
	code := `print("unclosed string`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for syntax error")
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestPythonRunner_RuntimeError(t *testing.T) {
	if err := CheckRuntime("python"); err != nil {
		t.Skip("Python not available:", err)
	}

	r := &PythonRunner{timeout: 30 * time.Second}
	code := `raise Exception("test error")`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for runtime error")
	}

	if !strings.Contains(result.Error, "test error") {
		t.Errorf("expected error to contain 'test error', got %q", result.Error)
	}
}

func TestPythonRunner_Timeout(t *testing.T) {
	if err := CheckRuntime("python"); err != nil {
		t.Skip("Python not available:", err)
	}

	r := &PythonRunner{timeout: 100 * time.Millisecond}
	code := `while True: pass`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for timeout")
	}

	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
}

func TestCheckRuntime_Node(t *testing.T) {
	err := CheckRuntime("javascript")
	if err != nil {
		t.Logf("Node.js not available (expected in some environments): %v", err)
	}
}

func TestCheckRuntime_Python(t *testing.T) {
	err := CheckRuntime("python")
	if err != nil {
		t.Logf("Python not available (expected in some environments): %v", err)
	}
}

func TestCheckRuntime_Unknown(t *testing.T) {
	err := CheckRuntime("unknown")
	if err == nil {
		t.Error("expected error for unknown language")
	}
}

func TestGetRuntimeVersion_Node(t *testing.T) {
	if err := CheckRuntime("javascript"); err != nil {
		t.Skip("Node.js not available:", err)
	}

	version, err := GetRuntimeVersion("javascript")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(version, "v") {
		t.Errorf("expected version to start with 'v', got %s", version)
	}
}

func TestGetRuntimeVersion_Python(t *testing.T) {
	if err := CheckRuntime("python"); err != nil {
		t.Skip("Python not available:", err)
	}

	version, err := GetRuntimeVersion("python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(version, "Python") {
		t.Errorf("expected version to contain 'Python', got %s", version)
	}
}

func TestExtractError_JavaScript(t *testing.T) {
	output := `file.js:1
throw new Error("something went wrong");
^

Error: something went wrong
    at Object.<anonymous> (file.js:1:7)
    at Module._compile (node:internal/modules/cjs/loader:1254:14)`

	err := extractError(output)

	if !strings.Contains(err, "something went wrong") {
		t.Errorf("expected extracted error to contain 'something went wrong', got %q", err)
	}
}

func TestExtractError_Python(t *testing.T) {
	output := `Traceback (most recent call last):
  File "test.py", line 1, in <module>
    raise Exception("test error")
Exception: test error`

	err := extractError(output)

	if !strings.Contains(err, "test error") {
		t.Errorf("expected extracted error to contain 'test error', got %q", err)
	}
}

func TestBashRunner_Success(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `echo "Hello, World!"`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", result.Output)
	}
}

func TestBashRunner_MultipleCommands(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `echo "First"
echo "Second"
echo "Third"`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "First") ||
		!strings.Contains(result.Output, "Second") ||
		!strings.Contains(result.Output, "Third") {
		t.Errorf("expected output to contain all three lines, got %q", result.Output)
	}
}

func TestBashRunner_SyntaxError(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `echo "unclosed string`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for syntax error")
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestBashRunner_RuntimeError(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `nonexistent_command`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for runtime error")
	}

	if !strings.Contains(result.Error, "command not found") && !strings.Contains(result.Error, "not found") {
		t.Errorf("expected error to contain 'command not found', got %q", result.Error)
	}
}

func TestBashRunner_ExitOnError(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `false
echo "This should not run"`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure due to set -e")
	}

	// The echo should not have run due to set -e
	if strings.Contains(result.Output, "This should not run") {
		t.Error("expected script to exit on first error, but it continued")
	}
}

func TestBashRunner_Timeout(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 100 * time.Millisecond}
	code := `sleep 10`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for timeout")
	}

	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
}

func TestBashRunner_Variables(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	r := &BashRunner{timeout: 30 * time.Second}
	code := `NAME="World"
echo "Hello, $NAME!"`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", result.Output)
	}
}

func TestCheckRuntime_Bash(t *testing.T) {
	err := CheckRuntime("bash")
	if err != nil {
		t.Logf("Bash not available (expected in some environments): %v", err)
	}
}

func TestGetRuntimeVersion_Bash(t *testing.T) {
	if err := CheckRuntime("bash"); err != nil {
		t.Skip("Bash not available:", err)
	}

	version, err := GetRuntimeVersion("bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(strings.ToLower(version), "bash") {
		t.Errorf("expected version to contain 'bash', got %s", version)
	}
}

func TestExtractError_Bash(t *testing.T) {
	output := `/tmp/docs-drift-12345.sh: line 3: nonexistent_command: command not found`

	err := extractError(output)

	if !strings.Contains(err, "command not found") {
		t.Errorf("expected extracted error to contain 'command not found', got %q", err)
	}
}

func TestGoRunner_Success(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	// Use longer timeout for first Go run (toolchain setup can be slow on some platforms)
	r := &GoRunner{timeout: 15 * time.Second}
	result := r.Run(context.Background(), `fmt.Println("Hello from Go")`)

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if !strings.Contains(result.Output, "Hello from Go") {
		t.Errorf("expected output to contain 'Hello from Go', got: %s", result.Output)
	}
}

func TestGoRunner_WithPackage(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	code := `package main

import "fmt"

func main() {
	fmt.Println("Complete Go program")
}`

	r := &GoRunner{timeout: 15 * time.Second}
	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if !strings.Contains(result.Output, "Complete Go program") {
		t.Errorf("expected output to contain 'Complete Go program', got: %s", result.Output)
	}
}

func TestGoRunner_CompileError(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	r := &GoRunner{timeout: 15 * time.Second}
	// Use undeclared variable instead of missing import
	result := r.Run(context.Background(), `x := undeclaredVariable`)

	if result.Success {
		t.Error("expected failure for compile error")
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestGoRunner_RuntimeError(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	code := `package main

func main() {
	panic("runtime error")
}`

	r := &GoRunner{timeout: 15 * time.Second}
	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for runtime panic")
	}
	if !strings.Contains(result.Error, "panic") {
		t.Errorf("expected panic in error, got: %s", result.Error)
	}
}

func TestGoRunner_Timeout(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	code := `package main

import "time"

func main() {
	time.Sleep(1 * time.Minute)
}`

	r := &GoRunner{timeout: 2 * time.Second}
	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected timeout failure")
	}
	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got: %s", result.Error)
	}
}

func TestCheckRuntime_Go(t *testing.T) {
	err := CheckRuntime("go")
	if err != nil {
		t.Logf("Go not available (expected in some environments): %v", err)
	}
}

func TestGetRuntimeVersion_Go(t *testing.T) {
	if err := CheckRuntime("go"); err != nil {
		t.Skip("Go not available:", err)
	}

	version, err := GetRuntimeVersion("go")
	if err != nil {
		t.Errorf("GetRuntimeVersion(go) failed: %v", err)
	}
	if version == "" {
		t.Error("expected non-empty version")
	}
	if !strings.Contains(version, "go") {
		t.Errorf("expected version to contain 'go', got: %s", version)
	}
}

func TestRubyRunner_Success(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	r := &RubyRunner{timeout: 30 * time.Second}
	code := `puts "Hello, World!"`

	result := r.Run(context.Background(), code)

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Error)
	}

	if !strings.Contains(result.Output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got %q", result.Output)
	}
}

func TestRubyRunner_SyntaxError(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	r := &RubyRunner{timeout: 30 * time.Second}
	code := `puts "unclosed string`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for syntax error")
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestRubyRunner_RuntimeError(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	r := &RubyRunner{timeout: 30 * time.Second}
	code := `raise "test error"`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for runtime error")
	}

	if !strings.Contains(result.Error, "test error") {
		t.Errorf("expected error to contain 'test error', got %q", result.Error)
	}
}

func TestRubyRunner_UndefinedMethod(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	r := &RubyRunner{timeout: 30 * time.Second}
	code := `nonexistent_method()`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for undefined method")
	}

	if !strings.Contains(result.Error, "undefined") {
		t.Errorf("expected error to contain 'undefined', got %q", result.Error)
	}
}

func TestRubyRunner_Timeout(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	r := &RubyRunner{timeout: 100 * time.Millisecond}
	code := `loop { }`

	result := r.Run(context.Background(), code)

	if result.Success {
		t.Error("expected failure for timeout")
	}

	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
}

func TestCheckRuntime_Ruby(t *testing.T) {
	err := CheckRuntime("ruby")
	if err != nil {
		t.Logf("Ruby not available (expected in some environments): %v", err)
	}
}

func TestGetRuntimeVersion_Ruby(t *testing.T) {
	if err := CheckRuntime("ruby"); err != nil {
		t.Skip("Ruby not available:", err)
	}

	version, err := GetRuntimeVersion("ruby")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(strings.ToLower(version), "ruby") {
		t.Errorf("expected version to contain 'ruby', got %s", version)
	}
}

func TestExtractError_Ruby(t *testing.T) {
	output := `test.rb:1:in '<main>': undefined method 'nonexistent_method' for main:Object (NoMethodError)`

	err := extractError(output)

	if !strings.Contains(err, "undefined method") {
		t.Errorf("expected extracted error to contain 'undefined method', got %q", err)
	}
}
