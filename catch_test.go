package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestCatch(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "nil error",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "non-nil error",
			err:     fmt.Errorf("test error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Catch(tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Catch() error = '%v', wantErr '%v'", err, tt.wantErr)
			}
		})
	}
}

func TestCatchf(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		msg        string
		params     []any
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "nil error",
			err:     nil,
			msg:     "test message: %v",
			params:  []any{"param"},
			wantErr: false,
		},
		{
			name:       "non-nil error with params",
			err:        fmt.Errorf("base error"),
			msg:        "test message: %v",
			params:     []any{"param"},
			wantErr:    true,
			wantErrMsg: "base error\ncatch_test.go:64: test message: param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Catchf(tt.err, tt.msg, tt.params...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Catchf() error = '%v', wantErr '%v'", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !strings.Contains(err.Error(), "base error") && !strings.Contains(err.Error(), "test message: param") {
				t.Errorf("Catchf() error message = '%v', want to contain both 'base error' and 'test message: param'", err)
			}
		})
	}
}

func TestIgnoreResult(t *testing.T) {
	callback := IgnoreResult[string]()
	if err := callback("test"); err != nil {
		t.Errorf("IgnoreResult callback returned error: %v", err)
	}
}

func TestCatchResult(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		err        error
		callback   func(string) error
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "success case",
			result: "success",
			err:    nil,
			callback: func(s string) error {
				if s != "success" {
					return fmt.Errorf("unexpected result: %s", s)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "input error case",
			result:  "",
			err:     fmt.Errorf("input error"),
			wantErr: true,
		},
		{
			name:   "callback error case",
			result: "fail",
			err:    nil,
			callback: func(s string) error {
				return fmt.Errorf("callback error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.callback == nil {
				tt.callback = func(s string) error { return nil }
			}
			err := CatchResult(tt.result, tt.err)(tt.callback)
			if (err != nil) != tt.wantErr {
				t.Errorf("CatchResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCatchResultf(t *testing.T) {
	tests := []struct {
		name       string
		result     string
		err        error
		callback   func(string) error
		format     string
		params     []any
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "success case",
			result: "success",
			err:    nil,
			callback: func(s string) error {
				if s != "success" {
					return fmt.Errorf("unexpected result: %s", s)
				}
				return nil
			},
			format:  "test format: %v",
			params:  []any{"param"},
			wantErr: false,
		},
		{
			name:       "input error case",
			result:     "",
			err:        fmt.Errorf("input error"),
			format:     "error occurred: %v",
			params:     []any{"test param"},
			wantErr:    true,
			wantErrMsg: "error occurred: test param",
		},
		{
			name:   "callback error case",
			result: "fail",
			err:    nil,
			callback: func(s string) error {
				return fmt.Errorf("callback error")
			},
			format:     "processing failed: %v",
			params:     []any{"test"},
			wantErr:    true,
			wantErrMsg: "processing failed: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.callback == nil {
				tt.callback = func(s string) error { return nil }
			}
			err := CatchResultf(tt.result, tt.err)(tt.callback, tt.format, tt.params...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CatchResultf() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				errStr := err.Error()
				switch tt.name {
				case "input error case":
					if !strings.Contains(errStr, "input error") || !strings.Contains(errStr, "error occurred: test param") {
						t.Errorf("CatchResultf() error message = '%v', want to contain both 'input error' and 'error occurred: test param'", err)
					}
				case "callback error case":
					if !strings.Contains(errStr, "callback error") || !strings.Contains(errStr, "processing failed: test") {
						t.Errorf("CatchResultf() error message = '%v', want to contain both 'callback error' and 'processing failed: test'", err)
					}
				}
			}
		})
	}
}

func TestCatchResultRecursive(t *testing.T) {
	tests := []struct {
		name           string
		outerResult    string
		outerErr       error
		innerResult    int
		innerErr       error
		wantErr        bool
		wantSideEffect int
	}{
		{
			name:           "success case - both callbacks execute",
			outerResult:    "outer",
			outerErr:       nil,
			innerResult:    42,
			innerErr:       nil,
			wantErr:        false,
			wantSideEffect: 2, // both callbacks increment
		},
		{
			name:           "outer error - no callbacks execute",
			outerResult:    "outer",
			outerErr:       fmt.Errorf("outer error"),
			innerResult:    42,
			innerErr:       nil,
			wantErr:        true,
			wantSideEffect: 0,
		},
		{
			name:           "inner error - only outer callback executes",
			outerResult:    "outer",
			outerErr:       nil,
			innerResult:    42,
			innerErr:       fmt.Errorf("inner error"),
			wantErr:        true,
			wantSideEffect: 1, // only outer callback increments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sideEffect := 0

			err := CatchResult(tt.outerResult, tt.outerErr)(func(outer string) error {
				sideEffect++ // outer callback side effect

				return CatchResult(tt.innerResult, tt.innerErr)(func(inner int) error {
					sideEffect++ // inner callback side effect
					return nil
				})
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("CatchResult() recursive error = %v, wantErr %v", err, tt.wantErr)
			}

			if sideEffect != tt.wantSideEffect {
				t.Errorf("CatchResult() side effect = %v, want %v", sideEffect, tt.wantSideEffect)
			}
		})
	}
}
