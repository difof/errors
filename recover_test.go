package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverError(t *testing.T) {
	tests := []struct {
		name string
		r    any
	}{
		{
			name: "nil recovery",
			r:    nil,
		},
		{
			name: "error recovery",
			r:    fmt.Errorf("test error"),
		},
		{
			name: "string recovery",
			r:    "panic message",
		},
		{
			name: "integer recovery",
			r:    42,
		},
		{
			name: "struct recovery",
			r:    struct{ msg string }{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := recoverError(tt.r, 1)
			if tt.r == nil {
				assert.Nil(t, err)
			} else if e, ok := tt.r.(error); ok {
				assert.Equal(t, e.Error(), err.Error())
			} else {
				assert.Contains(t, err.Error(), fmt.Sprintf("%v", tt.r))
			}
		})
	}
}

func TestRecover(t *testing.T) {
	t.Run("nil error pointer", func(t *testing.T) {
		// When errp is nil, panic should propagate
		assert.Panics(t, func() {
			defer Recover(nil)
			panic("test panic")
		})
	})

	t.Run("non-nil error pointer with no panic", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			// No panic
		}()
		assert.Nil(t, err)
	})

	t.Run("error panic", func(t *testing.T) {
		var err error
		expectedErr := fmt.Errorf("test error")
		func() {
			defer Recover(&err)
			panic(expectedErr)
		}()
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("string panic", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			panic("test panic")
		}()
		assert.Contains(t, err.Error(), "test panic")
	})

	t.Run("integration with Must", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			Must(fmt.Errorf("must error"))
		}()
		assert.NotNil(t, err)
	})

	t.Run("integration with MustResult", func(t *testing.T) {
		var err error
		var result string
		func() {
			defer Recover(&err)
			result = MustResult("", fmt.Errorf("must result error"))
		}()
		assert.NotNil(t, err)
		assert.Empty(t, result)
	})
}

func TestRecoverFn(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		called := false
		func() {
			defer RecoverFn(func(err error) {
				called = true
			})
			// No panic
		}()
		assert.False(t, called)
	})

	t.Run("error panic", func(t *testing.T) {
		var captured error
		expectedErr := fmt.Errorf("test error")
		func() {
			defer RecoverFn(func(err error) {
				captured = err
			})
			panic(expectedErr)
		}()
		assert.Equal(t, expectedErr.Error(), captured.Error())
	})

	t.Run("string panic", func(t *testing.T) {
		var captured error
		func() {
			defer RecoverFn(func(err error) {
				captured = err
			})
			panic("test panic")
		}()
		assert.Contains(t, captured.Error(), "test panic")
	})

	t.Run("nil callback", func(t *testing.T) {
		assert.NotPanics(t, func() {
			defer RecoverFn(nil)
			// Even with nil callback, should not panic
		})
	})

	t.Run("nil callback with panic", func(t *testing.T) {
		assert.Panics(t, func() {
			defer RecoverFn(nil)
			panic("test panic")
		})
	})

	t.Run("integration with Must chain", func(t *testing.T) {
		var captured error
		func() {
			defer RecoverFn(func(err error) {
				captured = err
			})
			MustResultf(0, fmt.Errorf("first error"))("with format %s", "test")
		}()
		assert.NotNil(t, captured)
	})
}

// Test real-world scenario with nested functions and error chains
func TestRecoverIntegration(t *testing.T) {
	// Helper function that may panic
	riskyFunc := func(shouldPanic bool) (err error) {
		defer Recover(&err)
		if shouldPanic {
			Must(fmt.Errorf("risky operation failed"))
		}
		return nil
	}

	// Test successful case
	err := riskyFunc(false)
	assert.Nil(t, err)

	// Test panic case
	err = riskyFunc(true)
	assert.NotNil(t, err)

	// Test nested recovery
	var outerErr error
	func() {
		defer Recover(&outerErr)
		err := riskyFunc(true)
		if err != nil {
			panic(fmt.Errorf("outer error: %w", err))
		}
	}()
	assert.NotNil(t, outerErr)
	assert.Contains(t, outerErr.Error(), "outer error")
}
