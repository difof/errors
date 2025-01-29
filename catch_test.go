package errors

import (
	"fmt"
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
				t.Errorf("Catch() error = %v, wantErr %v", err, tt.wantErr)
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
			wantErrMsg: "test message: param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Catchf(tt.err, tt.msg, tt.params...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Catchf() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !Is(err, New(tt.wantErrMsg)) {
				t.Errorf("Catchf() error message = %v, want to contain %v", err, tt.wantErrMsg)
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
			if tt.wantErr && err != nil && !Is(err, New(tt.wantErrMsg)) {
				t.Errorf("CatchResultf() error message = %v, want to contain %v", err, tt.wantErrMsg)
			}
		})
	}
}
