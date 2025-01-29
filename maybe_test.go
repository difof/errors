package errors

import (
	"fmt"
	"net"
	"testing"
)

// Custom error types for testing
type maybeTestError struct{ msg string }

func (e *maybeTestError) Error() string { return e.msg }

func (e *maybeTestError) Is(target error) bool {
	_, ok := target.(*maybeTestError)
	return ok
}

type maybeOtherError struct{ msg string }

func (e *maybeOtherError) Error() string { return e.msg }

func TestMaybe(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantPanic bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "matching error type",
			err:       &maybeTestError{msg: "test error"},
			wantPanic: false,
		},
		{
			name:      "non-matching error type",
			err:       &maybeOtherError{msg: "wrong type"},
			wantPanic: true,
		},
		{
			name:      "standard error",
			err:       fmt.Errorf("standard error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Maybe() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result := Maybe[*maybeTestError](tt.err)
			if tt.err != nil && !tt.wantPanic {
				if result.Error() != tt.err.Error() {
					t.Errorf("Maybe() = %v, want %v", result, tt.err)
				}
			}
		})
	}
}

func TestMaybef(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		format    string
		params    []any
		wantPanic bool
	}{
		{
			name:      "nil error",
			err:       nil,
			format:    "unexpected error: %v",
			params:    []any{"test"},
			wantPanic: false,
		},
		{
			name:      "matching error type",
			err:       &maybeTestError{msg: "test error"},
			format:    "unexpected error: %v",
			params:    []any{"test"},
			wantPanic: false,
		},
		{
			name:      "non-matching error type with format",
			err:       &maybeOtherError{msg: "wrong type"},
			format:    "unexpected type: %v",
			params:    []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Maybef() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result := Maybef[*maybeTestError](tt.err)(tt.format, tt.params...)
			if tt.err != nil && !tt.wantPanic {
				if result.Error() != tt.err.Error() {
					t.Errorf("Maybef() = %v, want %v", result, tt.err)
				}
			}
		})
	}
}

func TestMaybeResult(t *testing.T) {
	tests := []struct {
		name      string
		result    int
		err       error
		wantVal   int
		wantPanic bool
	}{
		{
			name:      "nil error",
			result:    42,
			err:       nil,
			wantVal:   42,
			wantPanic: false,
		},
		{
			name:      "matching error type",
			result:    42,
			err:       &maybeTestError{msg: "test error"},
			wantVal:   42,
			wantPanic: false,
		},
		{
			name:      "non-matching error type",
			result:    42,
			err:       &maybeOtherError{msg: "wrong type"},
			wantVal:   42,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MaybeResult() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result, err := MaybeResult[int, *maybeTestError](tt.result, tt.err)
			if !tt.wantPanic {
				if result != tt.wantVal {
					t.Errorf("MaybeResult() result = %v, want %v", result, tt.wantVal)
				}
				if tt.err != nil && err.Error() != tt.err.Error() {
					t.Errorf("MaybeResult() err = %v, want %v", err, tt.err)
				}
			}
		})
	}
}

func TestMaybeResultf(t *testing.T) {
	tests := []struct {
		name      string
		result    int
		err       error
		format    string
		params    []any
		wantVal   int
		wantPanic bool
	}{
		{
			name:      "nil error",
			result:    42,
			err:       nil,
			format:    "unexpected error: %v",
			params:    []any{"test"},
			wantVal:   42,
			wantPanic: false,
		},
		{
			name:      "matching error type",
			result:    42,
			err:       &maybeTestError{msg: "test error"},
			format:    "unexpected error: %v",
			params:    []any{"test"},
			wantVal:   42,
			wantPanic: false,
		},
		{
			name:      "non-matching error type with format",
			result:    42,
			err:       &maybeOtherError{msg: "wrong type"},
			format:    "unexpected type: %v",
			params:    []any{"test"},
			wantVal:   42,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MaybeResultf() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result, err := MaybeResultf[int, *maybeTestError](tt.result, tt.err)(tt.format, tt.params...)
			if !tt.wantPanic {
				if result != tt.wantVal {
					t.Errorf("MaybeResultf() result = %v, want %v", result, tt.wantVal)
				}
				if tt.err != nil && err.Error() != tt.err.Error() {
					t.Errorf("MaybeResultf() err = %v, want %v", err, tt.err)
				}
			}
		})
	}
}

// Example usage test with net.OpError
func TestMaybeWithNetError(t *testing.T) {
	// Create a real net.OpError
	netErr := &net.OpError{
		Op:     "read",
		Net:    "tcp",
		Source: nil,
		Addr:   nil,
		Err:    fmt.Errorf("connection refused"),
	}

	// Test successful case with net.OpError
	t.Run("successful case", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Maybe() panicked unexpectedly: %v", r)
			}
		}()

		result := Maybe[*net.OpError](netErr)
		if result.Op != "read" || result.Net != "tcp" {
			t.Errorf("Maybe() with net.OpError failed to preserve error details")
		}
	})

	// Test panic case with wrong error type
	t.Run("panic case", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Maybe() with wrong error type did not panic")
			}
		}()
		Maybe[*net.OpError](&maybeTestError{msg: "wrong type"})
	})

	// Test wrapped net.OpError
	t.Run("wrapped error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Maybe() panicked unexpectedly: %v", r)
			}
		}()

		wrapped := fmt.Errorf("wrapped: %w", netErr)
		result := Maybe[*net.OpError](wrapped)
		if result.Op != "read" || result.Net != "tcp" {
			t.Errorf("Maybe() with wrapped net.OpError failed to preserve error details")
		}
	})
}
