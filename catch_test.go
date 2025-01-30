package errors

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
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
			name:    "standard error",
			err:     errors.New("test error"),
			wantErr: true,
		},
		{
			name:    "custom error",
			err:     NewError("pkg.func", "file.go", 42, errors.New("custom error"), nil),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Catch(tt.err)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrorMessageOf(tt.err), ErrorMessageOf(err))
			} else {
				assert.NoError(t, err)
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
		wantPrefix string
	}{
		{
			name:    "nil error",
			err:     nil,
			msg:     "test message: %v",
			params:  []any{"param"},
			wantErr: false,
		},
		{
			name:       "standard error with format",
			err:        errors.New("base error"),
			msg:        "wrapped message: %v",
			params:     []any{"param"},
			wantErr:    true,
			wantPrefix: "wrapped message: param",
		},
		{
			name:       "custom error with empty format",
			err:        NewError("pkg.func", "file.go", 42, errors.New("custom error"), nil),
			msg:        "",
			params:     nil,
			wantErr:    true,
			wantPrefix: "",
		},
		{
			name:       "error with multiple params",
			err:        errors.New("base error"),
			msg:        "%v: %v - %v",
			params:     []any{"first", "second", "third"},
			wantErr:    true,
			wantPrefix: "first: second - third",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Catchf(tt.err, tt.msg, tt.params...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantPrefix != "" {
					assert.Contains(t, err.Error(), tt.wantPrefix)
				}
				// Check that the original error message is somewhere in the chain
				var found bool
				for e := err; e != nil; e = errors.Unwrap(e) {
					if e.Error() == tt.err.Error() {
						found = true
						break
					}
				}
				assert.True(t, found, "Original error message not found in chain")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIgnoreResult(t *testing.T) {
	tests := []struct {
		name string
		val  any
	}{
		{
			name: "ignore string",
			val:  "test string",
		},
		{
			name: "ignore int",
			val:  42,
		},
		{
			name: "ignore struct",
			val:  struct{ field string }{"test"},
		},
		{
			name: "ignore nil",
			val:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callback := IgnoreResult[any]()
			err := callback(tt.val)
			assert.NoError(t, err)
		})
	}
}

func TestCatchResult(t *testing.T) {
	type testStruct struct {
		field string
	}

	tests := []struct {
		name       string
		result     any
		err        error
		callback   func(any) error
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "nil error with successful callback",
			result: "test result",
			err:    nil,
			callback: func(r any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "nil error with failing callback",
			result: "test result",
			err:    nil,
			callback: func(r any) error {
				return errors.New("callback error")
			},
			wantErr:    true,
			wantErrMsg: "callback error",
		},
		{
			name:   "error with unused callback",
			result: "test result",
			err:    errors.New("initial error"),
			callback: func(r any) error {
				return nil
			},
			wantErr:    true,
			wantErrMsg: "initial error",
		},
		{
			name:   "nil error with nil callback result",
			result: (*testStruct)(nil),
			err:    nil,
			callback: func(r any) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CatchResult(tt.result, tt.err)(tt.callback)
			if tt.wantErr {
				assert.Error(t, err)
				// Check that the expected error message is somewhere in the chain
				var found bool
				for e := err; e != nil; e = errors.Unwrap(e) {
					if e.Error() == tt.wantErrMsg {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message not found in chain")
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test with type parameter
	t.Run("typed result", func(t *testing.T) {
		result := &sql.Row{}
		err := CatchResult[*sql.Row](result, nil)(func(r *sql.Row) error {
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestCatchResultf(t *testing.T) {
	type testStruct struct {
		field string
	}

	tests := []struct {
		name       string
		result     any
		err        error
		callback   func(any) error
		format     string
		params     []any
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "nil error with successful callback",
			result: "test result",
			err:    nil,
			callback: func(r any) error {
				return nil
			},
			format:  "error: %v",
			params:  []any{"test"},
			wantErr: false,
		},
		{
			name:   "nil error with failing callback",
			result: "test result",
			err:    nil,
			callback: func(r any) error {
				return errors.New("callback error")
			},
			format:     "wrapped error: %v",
			params:     []any{"test"},
			wantErr:    true,
			wantErrMsg: "callback error",
		},
		{
			name:   "error with unused callback",
			result: "test result",
			err:    errors.New("initial error"),
			callback: func(r any) error {
				return nil
			},
			format:     "wrapped error: %v",
			params:     []any{"test"},
			wantErr:    true,
			wantErrMsg: "initial error",
		},
		{
			name:   "nil error with nil callback result",
			result: (*testStruct)(nil),
			err:    nil,
			callback: func(r any) error {
				return nil
			},
			format:  "error: %v",
			params:  []any{"test"},
			wantErr: false,
		},
		{
			name:   "error with empty format",
			result: "test result",
			err:    errors.New("initial error"),
			callback: func(r any) error {
				return nil
			},
			format:     "",
			params:     nil,
			wantErr:    true,
			wantErrMsg: "initial error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CatchResultf(tt.result, tt.err)(tt.callback, tt.format, tt.params...)
			if tt.wantErr {
				assert.Error(t, err)
				// Check that the expected error message is somewhere in the chain
				var found bool
				for e := err; e != nil; e = errors.Unwrap(e) {
					if e.Error() == tt.wantErrMsg {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message not found in chain")
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test with type parameter
	t.Run("typed result", func(t *testing.T) {
		result := &sql.Row{}
		err := CatchResultf[*sql.Row](result, nil)(func(r *sql.Row) error {
			return nil
		}, "error: %v", "test")
		assert.NoError(t, err)
	})
}
