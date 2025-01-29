package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantNil  bool
		contains string
	}{
		{
			name:    "nil error",
			err:     nil,
			wantNil: true,
		},
		{
			name:     "simple error",
			err:      fmt.Errorf("test error"),
			contains: "test error",
		},
		{
			name:     "already wrapped error",
			err:      Wrap(fmt.Errorf("inner error")),
			contains: "inner error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrap() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Error("Wrap() = nil, want wrapped error")
				return
			}
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("Wrap() = %v, want to contain %v", got, tt.contains)
			}
		})
	}
}

func TestWrapResult(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		err        error
		wantResult string
		wantErr    bool
		contains   string
	}{
		{
			name:       "nil error",
			result:     "success",
			err:        nil,
			wantResult: "success",
			wantErr:    false,
		},
		{
			name:       "with error",
			result:     "failed",
			err:        fmt.Errorf("test error"),
			wantResult: "failed",
			wantErr:    true,
			contains:   "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WrapResult(tt.result, tt.err)
			if result != tt.wantResult {
				t.Errorf("WrapResult() result = %v, want %v", result, tt.wantResult)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("WrapResult() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("WrapResult() error = %v, want to contain %v", err, tt.contains)
			}
		})
	}
}

func TestWrapResultf(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		err        error
		format     string
		args       []any
		wantResult string
		wantErr    bool
		contains   string
	}{
		{
			name:       "nil error",
			result:     "success",
			err:        nil,
			format:     "operation failed: %v",
			args:       []any{"test"},
			wantResult: "success",
			wantErr:    false,
		},
		{
			name:       "with error and format",
			result:     "failed",
			err:        fmt.Errorf("base error"),
			format:     "operation failed: %v",
			args:       []any{"test"},
			wantResult: "failed",
			wantErr:    true,
			contains:   "base error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WrapResultf(tt.result, tt.err)(tt.format, tt.args...)
			if result != tt.wantResult {
				t.Errorf("WrapResultf() result = %v, want %v", result, tt.wantResult)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("WrapResultf() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("WrapResultf() error = %v, want to contain %v", err, tt.contains)
			}
		})
	}
}

func TestWrape(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		inner    error
		wantNil  bool
		contains []string
	}{
		{
			name:    "both nil",
			err:     nil,
			inner:   nil,
			wantNil: true,
		},
		{
			name:     "nil inner",
			err:      fmt.Errorf("test error"),
			inner:    nil,
			contains: []string{"test error"},
		},
		{
			name:     "nil err",
			err:      nil,
			inner:    fmt.Errorf("inner error"),
			contains: []string{"inner error"},
		},
		{
			name:     "both errors",
			err:      fmt.Errorf("test error"),
			inner:    fmt.Errorf("inner error"),
			contains: []string{"test error", "inner error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrape(tt.err, tt.inner)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrape() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Error("Wrape() = nil, want wrapped error")
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got.Error(), want) {
					t.Errorf("Wrape() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	tests := []struct {
		name     string
		inner    error
		format   string
		args     []any
		wantNil  bool
		contains []string
	}{
		{
			name:    "nil inner with empty format",
			inner:   nil,
			format:  "",
			wantNil: true,
		},
		{
			name:     "nil inner with format",
			inner:    nil,
			format:   "test error: %v",
			args:     []any{"value"},
			contains: []string{"test error: value"},
		},
		{
			name:     "with inner and format",
			inner:    fmt.Errorf("base error"),
			format:   "test error: %v",
			args:     []any{"value"},
			contains: []string{"base error", "test error: value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrapf(tt.inner, tt.format, tt.args...)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Wrapf() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Error("Wrapf() = nil, want wrapped error")
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got.Error(), want) {
					t.Errorf("Wrapf() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestWrapSkip(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		err      error
		wantNil  bool
		contains string
	}{
		{
			name:    "nil error",
			skip:    0,
			err:     nil,
			wantNil: true,
		},
		{
			name:     "skip 0",
			skip:     0,
			err:      fmt.Errorf("test error"),
			contains: "test error",
		},
		{
			name:     "skip 1",
			skip:     1,
			err:      fmt.Errorf("test error"),
			contains: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapSkip(tt.skip, tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapSkip() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Error("WrapSkip() = nil, want wrapped error")
				return
			}
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("WrapSkip() = %v, want to contain %v", got, tt.contains)
			}
		})
	}
}

func TestWrapSkipf(t *testing.T) {
	tests := []struct {
		name     string
		skip     int
		err      error
		format   string
		args     []any
		wantNil  bool
		contains []string
	}{
		{
			name:    "nil error with empty format",
			skip:    0,
			err:     nil,
			format:  "",
			wantNil: true,
		},
		{
			name:     "nil error with format",
			skip:     0,
			err:      nil,
			format:   "test format: %v",
			args:     []any{"value"},
			contains: []string{"test format: value"},
		},
		{
			name:     "skip 0 with format",
			skip:     0,
			err:      fmt.Errorf("base error"),
			format:   "test format: %v",
			args:     []any{"value"},
			contains: []string{"base error", "test format: value"},
		},
		{
			name:     "skip 1 with format",
			skip:     1,
			err:      fmt.Errorf("base error"),
			format:   "test format: %v",
			args:     []any{"value"},
			contains: []string{"base error", "test format: value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapSkipf(tt.skip, tt.err, tt.format, tt.args...)
			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapSkipf() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Error("WrapSkipf() = nil, want wrapped error")
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got.Error(), want) {
					t.Errorf("WrapSkipf() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}
