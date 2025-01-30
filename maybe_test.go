package errors

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Custom error types for testing
type maybeCustomError struct{ msg string }

func (e *maybeCustomError) Error() string { return e.msg }

type maybeWrappedError struct {
	err error
}

func (e *maybeWrappedError) Error() string { return e.err.Error() }
func (e *maybeWrappedError) Unwrap() error { return e.err }

func TestMaybe(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantPanic bool
		setup     func() *maybeCustomError // For cases where we need to prepare the target
	}{
		{
			name:      "nil error",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "direct type match",
			err:       &maybeCustomError{msg: "direct"},
			wantPanic: false,
		},
		{
			name:      "wrapped error match",
			err:       &maybeWrappedError{err: &maybeCustomError{msg: "wrapped"}},
			wantPanic: false,
		},
		{
			name:      "panic on non-matching error",
			err:       fmt.Errorf("standard error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					Maybe[*maybeCustomError](tt.err)
				})
				return
			}

			var result *maybeCustomError
			assert.NotPanics(t, func() {
				result = Maybe[*maybeCustomError](tt.err)
			})

			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMaybef(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		format    string
		args      []any
		wantPanic bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "direct type match",
			err:       &maybeCustomError{msg: "direct"},
			wantPanic: false,
		},
		{
			name:      "wrapped error match",
			err:       &maybeWrappedError{err: &maybeCustomError{msg: "wrapped"}},
			wantPanic: false,
		},
		{
			name:      "panic with custom message",
			err:       fmt.Errorf("standard error"),
			format:    "custom panic: %v",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					Maybef[*maybeCustomError](tt.err)(tt.format, tt.args...)
				})
				return
			}

			var result *maybeCustomError
			assert.NotPanics(t, func() {
				result = Maybef[*maybeCustomError](tt.err)("unused format")
			})

			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMaybeResult(t *testing.T) {
	tests := []struct {
		name      string
		result    string
		err       error
		wantPanic bool
	}{
		{
			name:      "nil error",
			result:    "success",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "direct type match",
			result:    "with error",
			err:       &maybeCustomError{msg: "direct"},
			wantPanic: false,
		},
		{
			name:      "wrapped error match",
			result:    "wrapped error",
			err:       &maybeWrappedError{err: &maybeCustomError{msg: "wrapped"}},
			wantPanic: false,
		},
		{
			name:      "panic on non-matching error",
			result:    "panic",
			err:       fmt.Errorf("standard error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					MaybeResult[string, *maybeCustomError](tt.result, tt.err)
				})
				return
			}

			var result string
			var errResult *maybeCustomError
			assert.NotPanics(t, func() {
				result, errResult = MaybeResult[string, *maybeCustomError](tt.result, tt.err)
			})

			assert.Equal(t, tt.result, result)
			if tt.err == nil {
				assert.Nil(t, errResult)
			} else {
				assert.NotNil(t, errResult)
			}
		})
	}
}

func TestMaybeResultf(t *testing.T) {
	tests := []struct {
		name      string
		result    string
		err       error
		format    string
		args      []any
		wantPanic bool
	}{
		{
			name:      "nil error",
			result:    "success",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "direct type match",
			result:    "with error",
			err:       &maybeCustomError{msg: "direct"},
			wantPanic: false,
		},
		{
			name:      "wrapped error match",
			result:    "wrapped error",
			err:       &maybeWrappedError{err: &maybeCustomError{msg: "wrapped"}},
			wantPanic: false,
		},
		{
			name:      "panic with custom message",
			result:    "panic",
			err:       fmt.Errorf("standard error"),
			format:    "custom panic: %v",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					MaybeResultf[string, *maybeCustomError](tt.result, tt.err)(tt.format, tt.args...)
				})
				return
			}

			var result string
			var errResult *maybeCustomError
			assert.NotPanics(t, func() {
				result, errResult = MaybeResultf[string, *maybeCustomError](tt.result, tt.err)("unused format")
			})

			assert.Equal(t, tt.result, result)
			if tt.err == nil {
				assert.Nil(t, errResult)
			} else {
				assert.NotNil(t, errResult)
			}
		})
	}
}

// Integration test with real error types
func TestMaybeWithNetError(t *testing.T) {
	// Test with actual net.OpError
	netErr := &net.OpError{
		Op:     "read",
		Net:    "tcp",
		Source: nil,
		Addr:   nil,
		Err:    fmt.Errorf("connection reset"),
	}

	// Direct match
	result := Maybe[*net.OpError](netErr)
	assert.Equal(t, netErr, result)

	// Wrapped match
	wrapped := &maybeWrappedError{err: netErr}
	result = Maybe[*net.OpError](wrapped)
	assert.Equal(t, netErr, result)

	// Nil case
	var nilResult *net.OpError
	nilResult = Maybe[*net.OpError](nil)
	assert.Nil(t, nilResult)

	// Panic case
	assert.Panics(t, func() {
		Maybe[*net.OpError](fmt.Errorf("not a net error"))
	})
}
