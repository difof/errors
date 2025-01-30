package errors

import (
	"testing"

	goerrors "errors"

	"github.com/stretchr/testify/assert"
)

// Custom error types for testing
type proxyCustomError struct{ msg string }

func (e *proxyCustomError) Error() string { return e.msg }

type proxyWrappedError struct {
	err error
	msg string
}

func (e *proxyWrappedError) Error() string { return e.msg }
func (e *proxyWrappedError) Unwrap() error { return e.err }

func TestAs(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		target  interface{}
		want    bool
		wantVal interface{}
	}{
		{
			name:    "nil error",
			err:     nil,
			target:  new(*proxyCustomError),
			want:    false,
			wantVal: (*proxyCustomError)(nil),
		},
		{
			name:    "direct match",
			err:     &proxyCustomError{"direct"},
			target:  new(*proxyCustomError),
			want:    true,
			wantVal: &proxyCustomError{"direct"},
		},
		{
			name:    "wrapped standard error",
			err:     &proxyWrappedError{&proxyCustomError{"wrapped"}, "outer"},
			target:  new(*proxyCustomError),
			want:    true,
			wantVal: &proxyCustomError{"wrapped"},
		},
		{
			name:    "custom Error type with inner proxyCustomError",
			err:     NewError("test.func", "test.go", 1, goerrors.New("outer"), &proxyCustomError{"inner"}),
			target:  new(*proxyCustomError),
			want:    true,
			wantVal: &proxyCustomError{"inner"},
		},
		{
			name:    "no match",
			err:     goerrors.New("test"),
			target:  new(*proxyCustomError),
			want:    false,
			wantVal: (*proxyCustomError)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := As(tt.err, tt.target)
			assert.Equal(t, tt.want, got)
			if ptr, ok := tt.target.(*(*proxyCustomError)); ok && ptr != nil {
				if *ptr == nil {
					assert.Equal(t, tt.wantVal, *ptr)
				} else {
					assert.Equal(t, tt.wantVal.(*proxyCustomError).msg, (*ptr).msg)
				}
			}
		})
	}
}

func TestIs(t *testing.T) {
	baseErr := goerrors.New("base error")
	customErr := &proxyCustomError{"custom"}
	wrappedErr := &proxyWrappedError{baseErr, "wrapped"}

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "nil error",
			err:    nil,
			target: baseErr,
			want:   false,
		},
		{
			name:   "nil target",
			err:    baseErr,
			target: nil,
			want:   false,
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
			name:   "custom Error type with matching inner",
			err:    NewError("test.func", "test.go", 1, goerrors.New("outer"), baseErr),
			target: baseErr,
			want:   true,
		},
		{
			name:   "no match",
			err:    baseErr,
			target: customErr,
			want:   false,
		},
		{
			name:   "custom Error type no match",
			err:    NewError("test.func", "test.go", 1, goerrors.New("outer"), customErr),
			target: baseErr,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnwrap(t *testing.T) {
	baseErr := goerrors.New("base error")
	wrappedErr := &proxyWrappedError{baseErr, "wrapped"}
	customErr := NewError("test.func", "test.go", 1, goerrors.New("outer"), baseErr)

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil error",
			err:  nil,
			want: nil,
		},
		{
			name: "error without Unwrap method",
			err:  baseErr,
			want: nil,
		},
		{
			name: "wrapped error",
			err:  wrappedErr,
			want: baseErr,
		},
		{
			name: "custom Error type",
			err:  customErr,
			want: baseErr,
		},
		{
			name: "multiple wraps",
			err:  &proxyWrappedError{wrappedErr, "double wrapped"},
			want: wrappedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unwrap(tt.err)
			if got != nil {
				assert.Equal(t, tt.want.Error(), got.Error())
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
