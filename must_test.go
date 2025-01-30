package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustResult(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		result := MustResult(42, nil)
		assert.Equal(t, 42, result)
	})

	t.Run("panic case", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResult(42, fmt.Errorf("test error"))
		})
	})
}

func TestMustResultf(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		result := MustResultf(42, nil)("format %s", "test")
		assert.Equal(t, 42, result)
	})

	t.Run("panic case with format", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResultf(42, fmt.Errorf("test error"))("custom: %v", fmt.Errorf("test error"))
		})
	})

	t.Run("panic case without format", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResultf(42, fmt.Errorf("test error"))("")
		})
	})
}

func TestMustResult2(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		a, b := MustResult2("hello", 42, nil)
		assert.Equal(t, "hello", a)
		assert.Equal(t, 42, b)
	})

	t.Run("panic case", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResult2("hello", 42, fmt.Errorf("test error"))
		})
	})
}

func TestMustResult2f(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		a, b := MustResult2f("hello", 42, nil)("format %s", "test")
		assert.Equal(t, "hello", a)
		assert.Equal(t, 42, b)
	})

	t.Run("panic case with format", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResult2f("hello", 42, fmt.Errorf("test error"))("custom: %v", fmt.Errorf("test error"))
		})
	})

	t.Run("panic case without format", func(t *testing.T) {
		assert.Panics(t, func() {
			MustResult2f("hello", 42, fmt.Errorf("test error"))("")
		})
	})
}

func TestMust(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Must(nil)
		})
	})

	t.Run("panic case", func(t *testing.T) {
		assert.Panics(t, func() {
			Must(fmt.Errorf("test error"))
		})
	})
}

func TestMustf(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Mustf(nil)("format %s", "test")
		})
	})

	t.Run("panic case with format", func(t *testing.T) {
		assert.Panics(t, func() {
			Mustf(fmt.Errorf("test error"))("custom: %v", fmt.Errorf("test error"))
		})
	})

	t.Run("panic case without format", func(t *testing.T) {
		assert.Panics(t, func() {
			Mustf(fmt.Errorf("test error"))("")
		})
	})
}

func TestIgnore(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		result := Ignore(42, nil)
		assert.Equal(t, 42, result)
	})

	t.Run("error case", func(t *testing.T) {
		result := Ignore(42, fmt.Errorf("test error"))
		assert.Equal(t, 42, result)
	})

	t.Run("zero value case", func(t *testing.T) {
		var zero int
		result := Ignore(zero, fmt.Errorf("test error"))
		assert.Equal(t, zero, result)
	})
}

func TestAssert(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Assert(true, "should not panic")
		})
	})

	t.Run("panic case", func(t *testing.T) {
		assert.PanicsWithValue(t, "should panic", func() {
			Assert(false, "should panic")
		})
	})
}

func TestAssertf(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Assertf(true, "should not %s", "panic")
		})
	})

	t.Run("panic case", func(t *testing.T) {
		assert.PanicsWithValue(t, "should panic: test", func() {
			Assertf(false, "should panic: %s", "test")
		})
	})
}

// Helper function to verify panic message
func verifyPanicMessage(t *testing.T, expectedMsg string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(*Error); ok {
				if err.Message != nil {
					assert.Equal(t, expectedMsg, err.Message.Error())
				} else {
					assert.Equal(t, expectedMsg, err.Inner.Error())
				}
			} else {
				t.Errorf("Expected *Error type, got %T", r)
			}
		}
	}()
	fn()
	t.Error("Expected panic did not occur")
}
