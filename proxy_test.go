package errors

import (
	"fmt"
	"testing"
)

// proxyTestError is a custom error type that implements error interface
type proxyTestError struct {
	msg string
}

func (e *proxyTestError) Error() string { return e.msg }

// proxyWrappedError is a custom error type that implements error interface and wraps another error
type proxyWrappedError struct {
	msg   string
	inner error
}

func (e *proxyWrappedError) Error() string { return e.msg }
func (e *proxyWrappedError) Unwrap() error { return e.inner }

// proxyTargetError is a custom error type used for testing error type assertions
type proxyTargetError struct {
	msg string
}

func (e *proxyTargetError) Error() string { return e.msg }

func TestIs(t *testing.T) {
	baseErr := &proxyTestError{msg: "base error"}
	wrappedErr := &proxyWrappedError{msg: "wrapped error", inner: baseErr}
	doubleWrappedErr := fmt.Errorf("outer: %w", wrappedErr)
	unrelatedErr := &proxyTestError{msg: "unrelated error"}

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "nil errors",
			err:    nil,
			target: nil,
			want:   true,
		},
		{
			name:   "direct match",
			err:    baseErr,
			target: baseErr,
			want:   true,
		},
		{
			name:   "wrapped error match",
			err:    wrappedErr,
			target: baseErr,
			want:   true,
		},
		{
			name:   "double wrapped error match",
			err:    doubleWrappedErr,
			target: baseErr,
			want:   true,
		},
		{
			name:   "no match",
			err:    baseErr,
			target: unrelatedErr,
			want:   false,
		},
		{
			name:   "nil error no match",
			err:    nil,
			target: baseErr,
			want:   false,
		},
		{
			name:   "error with nil target",
			err:    baseErr,
			target: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAs(t *testing.T) {
	baseErr := &proxyTargetError{msg: "target error"}
	wrappedErr := &proxyWrappedError{msg: "wrapped error", inner: baseErr}
	doubleWrappedErr := fmt.Errorf("outer: %w", wrappedErr)
	unrelatedErr := &proxyTestError{msg: "unrelated error"}

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "direct match",
			err:     baseErr,
			wantErr: true,
		},
		{
			name:    "wrapped error match",
			err:     wrappedErr,
			wantErr: true,
		},
		{
			name:    "double wrapped error match",
			err:     doubleWrappedErr,
			wantErr: true,
		},
		{
			name:    "no match",
			err:     unrelatedErr,
			wantErr: false,
		},
		{
			name:    "nil error",
			err:     nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *proxyTargetError
			got := As(tt.err, &target)
			if got != tt.wantErr {
				t.Errorf("As() = %v, want %v", got, tt.wantErr)
			}
			if got && target == nil {
				t.Error("As() succeeded but target is nil")
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	innerErr := fmt.Errorf("inner error")
	wrappedErr := &proxyWrappedError{msg: "wrapped error", inner: innerErr}
	doubleWrappedErr := fmt.Errorf("outer: %w", wrappedErr)
	nonWrappingErr := &proxyTestError{msg: "non-wrapping error"}

	tests := []struct {
		name    string
		err     error
		wantErr error
		wantNil bool
	}{
		{
			name:    "unwrap wrapped error",
			err:     wrappedErr,
			wantErr: innerErr,
		},
		{
			name:    "unwrap double wrapped error",
			err:     doubleWrappedErr,
			wantErr: wrappedErr,
		},
		{
			name:    "unwrap non-wrapping error",
			err:     nonWrappingErr,
			wantNil: true,
		},
		{
			name:    "unwrap nil error",
			err:     nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unwrap(tt.err)
			if tt.wantNil {
				if got != nil {
					t.Errorf("Unwrap() = %v, want nil", got)
				}
				return
			}
			if got != tt.wantErr {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}
