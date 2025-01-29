package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	const msg = "simple error message"
	err := New(msg)

	if err == nil {
		t.Fatal("New() returned nil")
	}

	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("New() returned %T, want *Error", err)
	}

	if e.Message == nil || e.Message.Error() != msg {
		t.Errorf("New() message = %v, want %v", e.Message, msg)
	}

	if e.Inner != nil {
		t.Errorf("New() inner = %v, want nil", e.Inner)
	}

	if e.Source == "" {
		t.Error("New() source is empty")
	}

	// Source should contain this file name
	if !strings.Contains(e.Source, "new_test.go") {
		t.Errorf("New() source = %v, want to contain 'new_test.go'", e.Source)
	}
}

func TestNewf(t *testing.T) {
	const format = "formatted error: %d %s"
	const num = 42
	const str = "test"
	expected := fmt.Sprintf(format, num, str)

	err := Newf(format, num, str)

	if err == nil {
		t.Fatal("Newf() returned nil")
	}

	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("Newf() returned %T, want *Error", err)
	}

	if e.Message == nil || e.Message.Error() != expected {
		t.Errorf("Newf() message = %v, want %v", e.Message, expected)
	}

	if e.Inner != nil {
		t.Errorf("Newf() inner = %v, want nil", e.Inner)
	}

	if !strings.Contains(e.Source, "new_test.go") {
		t.Errorf("Newf() source = %v, want to contain 'new_test.go'", e.Source)
	}
}

func TestNewSkip(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		msg      string
		wantFile string // expected file name in source
	}{
		{
			name:     "no skip",
			skip:     0,
			msg:      "direct error",
			wantFile: "new_test.go",
		},
		{
			name:     "skip one",
			skip:     1,
			msg:      "skipped error",
			wantFile: "runtime.go", // or some other file in the call stack
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSkip(tt.skip, tt.msg)

			if err == nil {
				t.Fatal("NewSkip() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatalf("NewSkip() returned %T, want *Error", err)
			}

			if e.Message == nil || e.Message.Error() != tt.msg {
				t.Errorf("NewSkip() message = %v, want %v", e.Message, tt.msg)
			}

			if e.Inner != nil {
				t.Errorf("NewSkip() inner = %v, want nil", e.Inner)
			}

			if tt.skip == 0 && !strings.Contains(e.Source, tt.wantFile) {
				t.Errorf("NewSkip() source = %v, want to contain %v", e.Source, tt.wantFile)
			}
		})
	}
}

func TestNewSkipf(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		format   string
		args     []any
		wantMsg  string
		wantFile string
	}{
		{
			name:     "no skip with format",
			skip:     0,
			format:   "error %d: %s",
			args:     []any{1, "test"},
			wantMsg:  "error 1: test",
			wantFile: "new_test.go",
		},
		{
			name:     "skip one with format",
			skip:     1,
			format:   "skipped error: %v",
			args:     []any{"message"},
			wantMsg:  "skipped error: message",
			wantFile: "runtime.go", // or some other file in the call stack
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSkipf(tt.skip, tt.format, tt.args...)

			if err == nil {
				t.Fatal("NewSkipf() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatalf("NewSkipf() returned %T, want *Error", err)
			}

			if e.Message == nil || e.Message.Error() != tt.wantMsg {
				t.Errorf("NewSkipf() message = %v, want %v", e.Message, tt.wantMsg)
			}

			if e.Inner != nil {
				t.Errorf("NewSkipf() inner = %v, want nil", e.Inner)
			}

			if tt.skip == 0 && !strings.Contains(e.Source, tt.wantFile) {
				t.Errorf("NewSkipf() source = %v, want to contain %v", e.Source, tt.wantFile)
			}
		})
	}
}

// Helper function to test error creation through multiple stack frames
func createErrorThroughFrames(depth int, useFormat bool) error {
	if depth == 0 {
		if useFormat {
			return Newf("depth %d error", depth)
		}
		return New("depth 0 error")
	}
	return createErrorThroughFrames(depth-1, useFormat)
}

func TestErrorStackFrames(t *testing.T) {
	tests := []struct {
		name      string
		depth     int
		useFormat bool
	}{
		{
			name:      "simple error",
			depth:     0,
			useFormat: false,
		},
		{
			name:      "nested error",
			depth:     3,
			useFormat: false,
		},
		{
			name:      "formatted nested error",
			depth:     3,
			useFormat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createErrorThroughFrames(tt.depth, tt.useFormat)

			if err == nil {
				t.Fatal("createErrorThroughFrames() returned nil")
			}

			e, ok := err.(*Error)
			if !ok {
				t.Fatalf("createErrorThroughFrames() returned %T, want *Error", err)
			}

			// Verify source contains correct file
			if !strings.Contains(e.Source, "new_test.go") {
				t.Errorf("source = %v, want to contain 'new_test.go'", e.Source)
			}

			// Verify message format
			if tt.useFormat {
				expected := fmt.Sprintf("depth %d error", 0)
				if e.Message.Error() != expected {
					t.Errorf("message = %v, want %v", e.Message, expected)
				}
			} else {
				if e.Message.Error() != "depth 0 error" {
					t.Errorf("message = %v, want 'depth 0 error'", e.Message)
				}
			}
		})
	}
}
