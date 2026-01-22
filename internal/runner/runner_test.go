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

func TestNewRunner_Aliases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"js", "javascript"},
		{"py", "python"},
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
