package errors

import (
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
	stackTrace := err.(*Error).StackTrace()
	expectedLayers := []string{
		"value out of range",
		"validation layer: invalid input",
		"database layer: failed to insert record",
		"service layer: failed to process request",
	}

	if len(stackTrace) != len(expectedLayers) {
		t.Errorf("expected %d layers in stack trace, got %d", len(expectedLayers), len(stackTrace))
	}

	for i, expected := range expectedLayers {
		if !strings.Contains(stackTrace[i], expected) {
			t.Errorf("layer %d: expected to contain %q, got %q", i, expected, stackTrace[i])
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
	stackTrace := err.(*Error).StackTrace()

	// Verify stack depth
	expectedDepth := maxDepth + 1 // maxDepth + initial call
	if len(stackTrace) != expectedDepth {
		t.Errorf("expected stack depth of %d, got %d", expectedDepth, len(stackTrace))
	}

	// Verify root cause is first
	if !strings.Contains(stackTrace[0], "reached max depth") {
		t.Errorf("expected root cause 'reached max depth', got %q", stackTrace[0])
	}

	// Verify recursion messages
	for i := 1; i < len(stackTrace); i++ {
		depth := maxDepth - i
		expected := fmt.Sprintf("recursion depth %d", depth)
		if !strings.Contains(stackTrace[i], expected) {
			t.Errorf("layer %d: expected to contain %q, got %q", i, expected, stackTrace[i])
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
			message: fmt.Errorf("test message"),
			inner:   nil,
		},
		{
			name:    "with inner error",
			source:  "test.go:42",
			message: fmt.Errorf("outer message"),
			inner:   fmt.Errorf("inner message"),
		},
		{
			name:    "nil message",
			source:  "test.go:42",
			message: nil,
			inner:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.source, tt.message, tt.inner)
			if err.Source != tt.source {
				t.Errorf("NewError().Source = %v, want %v", err.Source, tt.source)
			}
			if err.Message != tt.message {
				t.Errorf("NewError().Message = %v, want %v", err.Message, tt.message)
			}
			if err.Inner != tt.inner {
				t.Errorf("NewError().Inner = %v, want %v", err.Inner, tt.inner)
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
	innermost := NewError("inner.go:1", fmt.Errorf("innermost"), nil)
	middle := NewError("middle.go:1", fmt.Errorf("middle"), innermost)
	outer := NewError("outer.go:1", fmt.Errorf("outer"), middle)

	tests := []struct {
		name     string
		err      *Error
		wantLen  int
		stopAt   int
		callback func(error) bool
	}{
		{
			name:     "traverse all",
			err:      outer,
			wantLen:  3,
			callback: func(error) bool { return true },
		},
		{
			name:     "stop after first",
			err:      outer,
			wantLen:  1,
			callback: func(error) bool { return false },
		},
		{
			name:     "nil callback",
			err:      outer,
			wantLen:  0,
			callback: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			if tt.callback != nil {
				tt.err.Each(func(err error) bool {
					count++
					return tt.callback(err)
				})
			} else {
				tt.err.Each(nil)
			}
			if count != tt.wantLen {
				t.Errorf("Each() traversed %v errors, want %v", count, tt.wantLen)
			}
		})
	}
}

func TestError_StackTrace(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantLen  int
		contains []string
	}{
		{
			name:     "single error",
			err:      NewError("test.go:1", fmt.Errorf("test error"), nil),
			wantLen:  1,
			contains: []string{"test.go:1: test error"},
		},
		{
			name: "nested errors",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				NewError("inner.go:2", fmt.Errorf("inner error"), nil)),
			wantLen: 2,
			contains: []string{
				"inner.go:2: inner error",
				"outer.go:1: outer error",
			},
		},
		{
			name: "with standard error",
			err: NewError("outer.go:1",
				fmt.Errorf("outer error"),
				fmt.Errorf("standard error")),
			wantLen: 2,
			contains: []string{
				"standard error",
				"outer.go:1: outer error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.StackTrace()
			if len(got) != tt.wantLen {
				t.Errorf("StackTrace() returned %v lines, want %v", len(got), tt.wantLen)
			}
			for i, want := range tt.contains {
				if !strings.Contains(got[i], want) {
					t.Errorf("StackTrace()[%v] = %v, want to contain %v", i, got[i], want)
				}
			}
		})
	}
}

func TestError_String(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "with message",
			err:  NewError("test.go:1", fmt.Errorf("test error"), nil),
			want: "test.go:1: test error",
		},
		{
			name: "without message",
			err:  NewError("test.go:1", nil, nil),
			want: "test.go:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
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
