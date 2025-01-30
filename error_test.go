package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func createNestedServiceError() error {
	return Wrapf(createDatabaseError(), "service layer: failed to process request")
}

func createDatabaseError() error {
	return Wrapf(createValidationError(), "database layer: failed to insert record")
}

func createValidationError() error {
	return Wrapf(baseError, "validation layer: invalid input")
}

var baseError = fmt.Errorf("value out of range")
var unrelatedError = New("unrelated error")

func TestErrorWrapping(t *testing.T) {
	err := createNestedServiceError()

	// Test error wrapping and unwrapping
	if !Is(err, baseError) {
		t.Error("error chain should contain baseError")
	}

	if Is(err, unrelatedError) {
		t.Error("error chain should not contain unrelatedError")
	}

	// Verify the complete error stack
	errStr := err.Error()
	expectedLayers := []string{
		"value out of range",
		"validation layer: invalid input",
		"database layer: failed to insert record",
		"service layer: failed to process request",
	}

	for _, expected := range expectedLayers {
		if !strings.Contains(errStr, expected) {
			t.Errorf("error string should contain %q", expected)
		}
	}

	// Test error message extraction
	rootMsg := err.(*Error).ErrorMessage()
	if rootMsg != "value out of range" {
		t.Errorf("expected root message %q, got %q", "value out of range", rootMsg)
	}
}

func TestRecursiveErrorWrapping(t *testing.T) {
	const maxDepth = 3

	var createRecursiveError func(depth int) error
	createRecursiveError = func(depth int) error {
		if depth >= maxDepth {
			return New("reached max depth")
		}
		return Wrapf(createRecursiveError(depth+1), "recursion depth %d", depth)
	}

	err := createRecursiveError(0)
	errStr := err.Error()

	// Verify root cause
	if !strings.Contains(errStr, "reached max depth") {
		t.Errorf("expected root cause 'reached max depth' in error string")
	}

	// Verify recursion messages
	for i := 0; i < maxDepth; i++ {
		expected := fmt.Sprintf("recursion depth %d", i)
		if !strings.Contains(errStr, expected) {
			t.Errorf("expected to contain %q in error string", expected)
		}
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		message error
		inner   error
	}{
		{
			name:    "basic error",
			source:  "test.go:42",
			message: fmt.Errorf("test error"),
		},
		{
			name:    "with inner error",
			source:  "test.go:42",
			message: fmt.Errorf("outer error"),
			inner:   fmt.Errorf("inner error"),
		},
		{
			name:    "nil message",
			source:  "test.go:42",
			message: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.source, tt.message, tt.inner)

			if err == nil {
				t.Fatal("NewError() returned nil")
			}

			if err.FilePath != tt.source {
				t.Errorf("NewError() source = %v, want %v", err.FilePath, tt.source)
			}

			if tt.message != nil && err.Message.Error() != tt.message.Error() {
				t.Errorf("NewError() message = %v, want %v", err.Message, tt.message)
			}

			if tt.inner != nil && err.Inner.Error() != tt.inner.Error() {
				t.Errorf("NewError() inner = %v, want %v", err.Inner, tt.inner)
			}
		})
	}
}

func TestErrorMessageOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "simple error",
			err:  fmt.Errorf("test error"),
			want: "test error",
		},
		{
			name: "nested Error type",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			want: "inner error",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMessageOf(tt.err); got != tt.want {
				t.Errorf("ErrorMessageOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Each(t *testing.T) {
	// Disable function names for this test
	oldConfig := GetErrorConfig()
	SetErrorConfig(WithShowFuncName(false))
	defer SetErrorConfig(WithShowFuncName(oldConfig.ShowFuncName))

	tests := []struct {
		name     string
		err      *Error
		expected []error
		callback func(error) bool
	}{
		{
			name: "traverse all",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			expected: []error{
				NewError("outer.go:1", fmt.Errorf("outer error"),
					NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil),
			},
			callback: func(error) bool { return true },
		},
		{
			name: "stop after first",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			expected: []error{
				NewError("outer.go:1", fmt.Errorf("outer error"),
					NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			},
			callback: func(error) bool { return false },
		},
		{
			name: "nil callback",
			err: NewError("test.go:1",
				fmt.Errorf("test error"), nil),
			expected: nil,
			callback: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []error
			tt.err.Each(tt.callback)

			if tt.callback != nil {
				tt.err.Each(func(err error) bool {
					got = append(got, err)
					return tt.callback(err)
				})
			}

			if tt.expected == nil && got != nil {
				t.Errorf("Each() got %v, want nil", got)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("Each() got %d errors, want %d", len(got), len(tt.expected))
				return
			}

			for i := range got {
				if got[i].Error() != tt.expected[i].Error() {
					t.Errorf("Each() error at index %d = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "simple error",
			err:  NewError("test.go:1", fmt.Errorf("test error"), nil),
			want: "test error",
		},
		{
			name: "nested error",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			want: "inner error",
		},
		{
			name: "with standard error at bottom",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				fmt.Errorf("standard error")),
			want: "standard error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ErrorMessage(); got != tt.want {
				t.Errorf("ErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		contains []string
	}{
		{
			name:     "single error",
			err:      NewError("test.go:1", fmt.Errorf("test error"), nil),
			contains: []string{"test.go:1: test error"},
		},
		{
			name: "nested errors",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			contains: []string{
				"outer.go:1: outer error",
				"inner.go:2: inner error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Error() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	err := NewError("test.go:1", fmt.Errorf("test error"), inner)

	if got := err.Unwrap(); got != inner {
		t.Errorf("Unwrap() = %v, want %v", got, inner)
	}
}

func TestError_Is(t *testing.T) {
	target := fmt.Errorf("target error")

	tests := []struct {
		name   string
		err    *Error
		target error
		want   bool
	}{
		{
			name:   "nil error",
			err:    nil,
			target: nil,
			want:   true,
		},
		{
			name:   "matching message",
			err:    NewError("test.go:1", target, nil),
			target: target,
			want:   true,
		},
		{
			name:   "matching inner",
			err:    NewError("test.go:1", fmt.Errorf("other"), target),
			target: target,
			want:   true,
		},
		{
			name:   "no match",
			err:    NewError("test.go:1", fmt.Errorf("other"), fmt.Errorf("inner")),
			target: target,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Is(tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

func TestError_As(t *testing.T) {
	target := &customError{msg: "test"}

	tests := []struct {
		name string
		err  *Error
		want bool
	}{
		{
			name: "matching message type",
			err:  NewError("test.go:1", target, nil),
			want: true,
		},
		{
			name: "matching inner type",
			err:  NewError("test.go:1", fmt.Errorf("other"), target),
			want: true,
		},
		{
			name: "no match",
			err:  NewError("test.go:1", fmt.Errorf("other"), fmt.Errorf("inner")),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dest *customError
			if got := tt.err.As(&dest); got != tt.want {
				t.Errorf("As() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_JSON(t *testing.T) {
	e := NewError("test.go:42", fmt.Errorf("test error"), NewError("inner.go:24", fmt.Errorf("inner error"), nil))

	got := e.JSON()
	t.Logf("JSON output: %q", got)

	// Verify JSON structure
	var stack []struct {
		FilePath string `json:"filepath"`
		FuncPath string `json:"funcpath"`
		Message  string `json:"message"`
	}
	if err := json.Unmarshal([]byte(got), &stack); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(stack) != 2 {
		t.Fatalf("Expected 2 errors in stack, got %d", len(stack))
	}

	if stack[0].FilePath != "inner.go:24" {
		t.Errorf("Expected filepath inner.go:24, got %s", stack[0].FilePath)
	}
	if stack[0].Message != "inner error" {
		t.Errorf("Expected message inner error, got %s", stack[0].Message)
	}

	if stack[1].FilePath != "test.go:42" {
		t.Errorf("Expected filepath test.go:42, got %s", stack[1].FilePath)
	}
	if stack[1].Message != "test error" {
		t.Errorf("Expected message test error, got %s", stack[1].Message)
	}
}

func TestError_YAML(t *testing.T) {
	e := NewError("test.go:42", fmt.Errorf("test error"), NewError("inner.go:24", fmt.Errorf("inner error"), nil))

	got := e.YAML()
	t.Logf("YAML output: %s", got)

	want := []string{
		"errors:",
		"  - filepath: inner.go:24",
		"    message: inner error",
		"  - filepath: test.go:42",
		"    message: test error",
	}

	for _, line := range want {
		if !strings.Contains(got, line) {
			t.Errorf("YAML output missing %q", line)
		}
	}
}

func TestError_Colored(t *testing.T) {
	e := NewError("test.go:42", fmt.Errorf("test error"), nil)

	got := e.Colored()
	stripped := stripColors(got)

	// Verify the format matches the standard text format
	want := "at test.go:42: test error"
	if stripped != want {
		t.Errorf("Colored output after stripping colors = %q, want %q", stripped, want)
	}

	// Verify colored content is present
	if !strings.Contains(got, "test error") {
		t.Error("Colored output missing \"test error\"")
	}
	if !strings.Contains(got, "test.go:42") {
		t.Error("Colored output missing \"test.go:42\"")
	}
}

// stripColors removes ANSI color codes from a string
func stripColors(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
