package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		inner     error
		wantNil   bool
		wantInner string
	}{
		{
			name:    "nil error",
			inner:   nil,
			wantNil: true,
		},
		{
			name:      "standard error",
			inner:     errors.New("standard error"),
			wantInner: "standard error",
		},
		{
			name:      "wrapped error",
			inner:     Wrap(errors.New("inner error")),
			wantInner: "inner error",
		},
		{
			name:      "deeply wrapped error",
			inner:     Wrap(Wrap(errors.New("deep error"))),
			wantInner: "deep error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.inner)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tt.wantInner, ErrorMessageOf(got))

			// Test that stack trace info is present
			if e, ok := got.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)
			}
		})
	}
}

func TestWrapResult(t *testing.T) {
	type testStruct struct {
		value string
	}

	tests := []struct {
		name    string
		result  testStruct
		err     error
		wantErr bool
	}{
		{
			name:    "nil error",
			result:  testStruct{value: "success"},
			err:     nil,
			wantErr: false,
		},
		{
			name:    "with error",
			result:  testStruct{value: "failure"},
			err:     errors.New("test error"),
			wantErr: true,
		},
		{
			name:    "with wrapped error",
			result:  testStruct{value: "failure"},
			err:     Wrap(errors.New("wrapped error")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WrapResult(tt.result, tt.err)

			// Check result is passed through correctly
			assert.Equal(t, tt.result, result)

			if !tt.wantErr {
				assert.Nil(t, err)
				return
			}

			assert.NotNil(t, err)
			if e, ok := err.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)
			}
		})
	}

	// Test with different types
	t.Run("different types", func(t *testing.T) {
		// Test with int
		intResult, err := WrapResult(42, errors.New("int error"))
		assert.Equal(t, 42, intResult)
		assert.NotNil(t, err)
		assert.Equal(t, "int error", ErrorMessageOf(err))

		// Test with string
		strResult, err := WrapResult("test", errors.New("string error"))
		assert.Equal(t, "test", strResult)
		assert.NotNil(t, err)
		assert.Equal(t, "string error", ErrorMessageOf(err))

		// Test with bool
		boolResult, err := WrapResult(true, errors.New("bool error"))
		assert.Equal(t, true, boolResult)
		assert.NotNil(t, err)
		assert.Equal(t, "bool error", ErrorMessageOf(err))
	})
}

func TestWrapResultf(t *testing.T) {
	type testStruct struct {
		value string
	}

	tests := []struct {
		name       string
		result     testStruct
		err        error
		format     string
		formatArgs []any
		wantErr    bool
		wantMsg    string
	}{
		{
			name:    "nil error",
			result:  testStruct{value: "success"},
			err:     nil,
			format:  "should not appear: %v",
			wantErr: false,
		},
		{
			name:       "with error and format",
			result:     testStruct{value: "failure"},
			err:        errors.New("test error"),
			format:     "operation failed: %v",
			formatArgs: []any{"custom message"},
			wantErr:    true,
			wantMsg:    "operation failed: custom message",
		},
		{
			name:       "with wrapped error",
			result:     testStruct{value: "failure"},
			err:        Wrap(errors.New("wrapped error")),
			format:     "wrapped operation failed: %v",
			formatArgs: []any{"custom message"},
			wantErr:    true,
			wantMsg:    "wrapped operation failed: custom message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := WrapResultf(tt.result, tt.err)
			result, err := formatter(tt.format, tt.formatArgs...)

			// Check result is passed through correctly
			assert.Equal(t, tt.result, result)

			if !tt.wantErr {
				assert.Nil(t, err)
				return
			}

			assert.NotNil(t, err)
			if e, ok := err.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)
				if tt.wantMsg != "" {
					assert.Equal(t, tt.wantMsg, e.Message.Error())
				}
			}
		})
	}

	// Test with different types
	t.Run("different types", func(t *testing.T) {
		// Test with int
		intFormatter := WrapResultf(42, errors.New("int error"))
		intResult, err := intFormatter("formatted: %v", "int")
		assert.Equal(t, 42, intResult)
		assert.NotNil(t, err)
		if e, ok := err.(*Error); ok {
			assert.Equal(t, "formatted: int", e.Message.Error())
		}

		// Test with string
		strFormatter := WrapResultf("test", errors.New("string error"))
		strResult, err := strFormatter("formatted: %v", "string")
		assert.Equal(t, "test", strResult)
		assert.NotNil(t, err)
		if e, ok := err.(*Error); ok {
			assert.Equal(t, "formatted: string", e.Message.Error())
		}
	})
}

func TestWrape(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		inner     error
		wantNil   bool
		wantMsg   string
		wantInner string
	}{
		{
			name:    "both nil",
			err:     nil,
			inner:   nil,
			wantNil: true,
		},
		{
			name:      "err only",
			err:       errors.New("outer error"),
			inner:     nil,
			wantMsg:   "outer error",
			wantInner: "",
		},
		{
			name:      "inner only",
			err:       nil,
			inner:     errors.New("inner error"),
			wantMsg:   "",
			wantInner: "inner error",
		},
		{
			name:      "both errors",
			err:       errors.New("outer error"),
			inner:     errors.New("inner error"),
			wantMsg:   "outer error",
			wantInner: "inner error",
		},
		{
			name:      "wrapped inner",
			err:       errors.New("outer error"),
			inner:     Wrap(errors.New("wrapped inner")),
			wantMsg:   "outer error",
			wantInner: "wrapped inner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrape(tt.err, tt.inner)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if e, ok := got.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)

				if tt.wantMsg != "" {
					assert.Equal(t, tt.wantMsg, e.Message.Error())
				}
				if tt.wantInner != "" && e.Inner != nil {
					assert.Equal(t, tt.wantInner, ErrorMessageOf(e.Inner))
				}
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		name      string
		inner     error
		format    string
		args      []any
		wantNil   bool
		wantMsg   string
		wantInner string
	}{
		{
			name:    "nil error and empty format",
			inner:   nil,
			format:  "",
			wantNil: true,
		},
		{
			name:      "standard error with format",
			inner:     errors.New("standard error"),
			format:    "wrapped: %v",
			args:      []any{"test message"},
			wantMsg:   "wrapped: test message",
			wantInner: "standard error",
		},
		{
			name:      "wrapped error with format",
			inner:     Wrap(errors.New("inner error")),
			format:    "outer: %v %v",
			args:      []any{"message", 42},
			wantMsg:   "outer: message 42",
			wantInner: "inner error",
		},
		{
			name:      "empty format",
			inner:     errors.New("inner error"),
			format:    "",
			wantMsg:   "",
			wantInner: "inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrapf(tt.inner, tt.format, tt.args...)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if e, ok := got.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)

				if tt.wantMsg != "" {
					assert.Equal(t, tt.wantMsg, e.Message.Error())
				}
				if tt.wantInner != "" {
					assert.Equal(t, tt.wantInner, ErrorMessageOf(e.Inner))
				}
			}
		})
	}
}

func TestWrapSkip(t *testing.T) {
	tests := []struct {
		name      string
		skip      int
		inner     error
		wantNil   bool
		wantInner string
	}{
		{
			name:    "nil error",
			skip:    1,
			inner:   nil,
			wantNil: true,
		},
		{
			name:      "standard error with skip 0",
			skip:      0,
			inner:     errors.New("test error"),
			wantInner: "test error",
		},
		{
			name:      "standard error with skip 1",
			skip:      1,
			inner:     errors.New("test error"),
			wantInner: "test error",
		},
		{
			name:      "wrapped error with skip 2",
			skip:      2,
			inner:     Wrap(errors.New("inner error")),
			wantInner: "inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapSkip(tt.skip, tt.inner)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			assert.Equal(t, tt.wantInner, ErrorMessageOf(got))

			if e, ok := got.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)
			}
		})
	}
}

func TestWrapSkipf(t *testing.T) {
	tests := []struct {
		name      string
		skip      int
		inner     error
		format    string
		args      []any
		wantNil   bool
		wantMsg   string
		wantInner string
	}{
		{
			name:    "nil error and empty format",
			skip:    1,
			inner:   nil,
			format:  "",
			wantNil: true,
		},
		{
			name:      "standard error with format",
			skip:      1,
			inner:     errors.New("test error"),
			format:    "wrapped: %v",
			args:      []any{"test message"},
			wantMsg:   "wrapped: test message",
			wantInner: "test error",
		},
		{
			name:      "wrapped error with format and skip 2",
			skip:      2,
			inner:     Wrap(errors.New("inner error")),
			format:    "outer: %v %v",
			args:      []any{"message", 42},
			wantMsg:   "outer: message 42",
			wantInner: "inner error",
		},
		{
			name:      "nil error with format",
			skip:      1,
			inner:     nil,
			format:    "should appear: %v",
			args:      []any{"test"},
			wantMsg:   "should appear: test",
			wantInner: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapSkipf(tt.skip, tt.inner, tt.format, tt.args...)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			assert.NotNil(t, got)
			if e, ok := got.(*Error); ok {
				assert.NotEmpty(t, e.FuncPath)
				assert.NotEmpty(t, e.FilePath)
				assert.Greater(t, e.Line, 0)

				if tt.wantMsg != "" {
					assert.Equal(t, tt.wantMsg, e.Message.Error())
				}
				if tt.wantInner != "" {
					assert.Equal(t, tt.wantInner, ErrorMessageOf(e.Inner))
				}
			}
		})
	}

	// Test error formatting
	t.Run("format error", func(t *testing.T) {
		err := fmt.Errorf("format error")
		got := WrapSkipf(1, err, "wrapped: %v", "test")
		assert.NotNil(t, got)
		if e, ok := got.(*Error); ok {
			assert.Equal(t, "wrapped: test", e.Message.Error())
			assert.Equal(t, "format error", ErrorMessageOf(e.Inner))
		}
	})
}
