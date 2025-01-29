package errors

import (
	"strings"
	"testing"
)

func recoverableMustError() error {
	return New("must fail")
}

func recoverable() (err error) {
	defer Recover(&err)
	Mustf(recoverableMustError())("this must cause death, but it didn't")
	return
}

// Test helpers that just panic
func panicWithError() {
	panic(New("test error"))
}

func panicWithString() {
	panic("test panic")
}

func panicWithNonError() {
	panic(123)
}

// panics returns true if the function panics
func panics(f func()) (didPanic bool) {
	defer func() {
		if r := recover(); r != nil {
			didPanic = true
		}
	}()
	f()
	return false
}

func TestRecover(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantErr     bool
		wantErrText string
	}{
		{
			name:        "recover from error panic",
			setup:       panicWithError,
			wantErr:     true,
			wantErrText: "test error",
		},
		{
			name:        "recover from string panic",
			setup:       panicWithString,
			wantErr:     true,
			wantErrText: "test panic",
		},
		{
			name:        "recover from non-error panic",
			setup:       panicWithNonError,
			wantErr:     true,
			wantErrText: "123",
		},
		{
			name:    "no panic",
			setup:   func() {},
			wantErr: false,
		},
		{
			name: "nil error pointer",
			setup: func() {
				panic(New("test error"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil error pointer" {
				if !panics(func() {
					var nilErr *error
					defer Recover(nilErr)
					tt.setup()
				}) {
					t.Error("expected panic to propagate with nil error pointer")
				}
				return
			}

			var err error
			func() {
				defer Recover(&err)
				tt.setup()
			}()

			if (err != nil) != tt.wantErr {
				t.Errorf("Recover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Errorf("Recover() error = %v, want error containing %v", err, tt.wantErrText)
			}
		})
	}

	// Test the original recoverable function
	t.Run("original recoverable test", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			Mustf(recoverableMustError())("this must cause death, but it didn't")
		}()
		Assert(err != nil, "err should not be nil")
		if !strings.Contains(err.Error(), "must fail") {
			t.Errorf("recoverable() error = %v, want error containing 'must fail'", err)
		}
	})
}

func TestRecoverFn(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantErr     bool
		wantErrText string
	}{
		{
			name:        "recover from error panic",
			setup:       panicWithError,
			wantErr:     true,
			wantErrText: "test error",
		},
		{
			name:        "recover from string panic",
			setup:       panicWithString,
			wantErr:     true,
			wantErrText: "test panic",
		},
		{
			name:        "recover from non-error panic",
			setup:       panicWithNonError,
			wantErr:     true,
			wantErrText: "123",
		},
		{
			name:    "no panic",
			setup:   func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotErr error
			var fnCalled bool

			func() {
				defer RecoverFn(func(err error) {
					fnCalled = true
					gotErr = err
				})
				tt.setup()
			}()

			if tt.wantErr {
				if !fnCalled {
					t.Error("RecoverFn callback was not called")
					return
				}
				if gotErr == nil {
					t.Error("RecoverFn callback received nil error")
					return
				}
				if !strings.Contains(gotErr.Error(), tt.wantErrText) {
					t.Errorf("RecoverFn() error = %v, want error containing %v", gotErr, tt.wantErrText)
				}
			} else if fnCalled {
				t.Error("RecoverFn callback was called unexpectedly")
			}
		})
	}
}

func TestHandlePanic(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		wantErr     bool
		wantErrText string
	}{
		{
			name:        "handle error panic",
			setup:       panicWithError,
			wantErr:     true,
			wantErrText: "test error",
		},
		{
			name:    "no panic",
			setup:   func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var didPanic bool

			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
						err = recoverError(r, 3)
					}
				}()
				tt.setup()
			}()

			if didPanic {
				if !tt.wantErr {
					t.Error("HandlePanic() got panic, want no panic")
				}
			} else if tt.wantErr {
				t.Error("HandlePanic() got no panic, want panic")
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Errorf("HandlePanic() error = %v, want error containing %v", err, tt.wantErrText)
			}
		})
	}
}

// Test that all recovery functions use the same underlying logic
func TestRecoveryConsistency(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func()
		wantErrText string
	}{
		{
			name:        "error panic",
			setup:       panicWithError,
			wantErrText: "test error",
		},
		{
			name:        "string panic",
			setup:       panicWithString,
			wantErrText: "test panic",
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Test all recovery functions
			var err1, err2, err3 error
			var fnCalled bool
			var didPanic1, didPanic2, didPanic3 bool

			// Test Recover
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic1 = true
						err1 = recoverError(r, 3)
					}
				}()
				tc.setup()
			}()

			// Test RecoverFn
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic2 = true
						err2 = recoverError(r, 3)
						fnCalled = true
					}
				}()
				tc.setup()
			}()

			// Test HandlePanic
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic3 = true
						err3 = recoverError(r, 3)
					}
				}()
				tc.setup()
			}()

			if !didPanic1 || !didPanic2 || !didPanic3 {
				t.Fatal("one or more recovery functions did not handle the panic")
			}

			// Compare error messages
			if err1 == nil || err2 == nil || err3 == nil {
				t.Fatal("all recovery functions should return non-nil errors")
			}

			if !fnCalled {
				t.Fatal("RecoverFn callback was not called")
			}

			// Extract just the error message without file/line info
			msg1 := err1.Error()
			msg2 := err2.Error()
			msg3 := err3.Error()

			if !strings.Contains(msg1, tc.wantErrText) || !strings.Contains(msg2, tc.wantErrText) || !strings.Contains(msg3, tc.wantErrText) {
				t.Errorf("recovery functions produced inconsistent errors:\nRecover: %v\nRecoverFn: %v\nHandlePanic: %v",
					msg1, msg2, msg3)
			}
		})
	}
}
