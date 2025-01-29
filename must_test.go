package errors

import (
	"fmt"
	"strconv"
	"testing"
)

func TestMustResult(t *testing.T) {
	tests := []struct {
		name      string
		result    int
		err       error
		wantVal   int
		wantPanic bool
	}{
		{
			name:      "success case",
			result:    42,
			err:       nil,
			wantVal:   42,
			wantPanic: false,
		},
		{
			name:      "error case",
			result:    0,
			err:       fmt.Errorf("test error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MustResult() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result := MustResult(tt.result, tt.err)
			if !tt.wantPanic && result != tt.wantVal {
				t.Errorf("MustResult() = %v, want %v", result, tt.wantVal)
			}
		})
	}
}

func TestMustResultf(t *testing.T) {
	tests := []struct {
		name      string
		result    string
		err       error
		format    string
		args      []any
		wantVal   string
		wantPanic bool
	}{
		{
			name:      "success case",
			result:    "success",
			err:       nil,
			format:    "error: %s",
			args:      []any{"test"},
			wantVal:   "success",
			wantPanic: false,
		},
		{
			name:      "error case with format",
			result:    "",
			err:       fmt.Errorf("base error"),
			format:    "custom error: %s",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MustResultf() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			result := MustResultf(tt.result, tt.err)(tt.format, tt.args...)
			if !tt.wantPanic && result != tt.wantVal {
				t.Errorf("MustResultf() = %v, want %v", result, tt.wantVal)
			}
		})
	}
}

func TestMustResult2(t *testing.T) {
	tests := []struct {
		name      string
		a         int
		b         string
		err       error
		wantA     int
		wantB     string
		wantPanic bool
	}{
		{
			name:      "success case",
			a:         42,
			b:         "test",
			err:       nil,
			wantA:     42,
			wantB:     "test",
			wantPanic: false,
		},
		{
			name:      "error case",
			a:         0,
			b:         "",
			err:       fmt.Errorf("test error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MustResult2() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			a, b := MustResult2(tt.a, tt.b, tt.err)
			if !tt.wantPanic {
				if a != tt.wantA || b != tt.wantB {
					t.Errorf("MustResult2() = (%v, %v), want (%v, %v)", a, b, tt.wantA, tt.wantB)
				}
			}
		})
	}
}

func TestMustResult2f(t *testing.T) {
	tests := []struct {
		name      string
		a         int
		b         string
		err       error
		format    string
		args      []any
		wantA     int
		wantB     string
		wantPanic bool
	}{
		{
			name:      "success case",
			a:         42,
			b:         "test",
			err:       nil,
			format:    "error: %s",
			args:      []any{"test"},
			wantA:     42,
			wantB:     "test",
			wantPanic: false,
		},
		{
			name:      "error case with format",
			a:         0,
			b:         "",
			err:       fmt.Errorf("base error"),
			format:    "custom error: %s",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("MustResult2f() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			a, b := MustResult2f(tt.a, tt.b, tt.err)(tt.format, tt.args...)
			if !tt.wantPanic {
				if a != tt.wantA || b != tt.wantB {
					t.Errorf("MustResult2f() = (%v, %v), want (%v, %v)", a, b, tt.wantA, tt.wantB)
				}
			}
		})
	}
}

func TestMust(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantPanic bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantPanic: false,
		},
		{
			name:      "with error",
			err:       fmt.Errorf("test error"),
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Must() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			Must(tt.err)
		})
	}
}

func TestMustf(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		format    string
		args      []any
		wantPanic bool
	}{
		{
			name:      "nil error",
			err:       nil,
			format:    "error: %s",
			args:      []any{"test"},
			wantPanic: false,
		},
		{
			name:      "with error and format",
			err:       fmt.Errorf("base error"),
			format:    "custom error: %s",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Mustf() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			Mustf(tt.err)(tt.format, tt.args...)
		})
	}
}

func TestIgnore(t *testing.T) {
	tests := []struct {
		name    string
		result  int
		err     error
		wantVal int
	}{
		{
			name:    "success case",
			result:  42,
			err:     nil,
			wantVal: 42,
		},
		{
			name:    "error case",
			result:  42,
			err:     fmt.Errorf("ignored error"),
			wantVal: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Ignore(tt.result, tt.err)
			if result != tt.wantVal {
				t.Errorf("Ignore() = %v, want %v", result, tt.wantVal)
			}
		})
	}
}

func TestAssert(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		message   string
		wantPanic bool
	}{
		{
			name:      "true condition",
			condition: true,
			message:   "should not panic",
			wantPanic: false,
		},
		{
			name:      "false condition",
			condition: false,
			message:   "should panic",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Assert() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
				if tt.wantPanic && r != tt.message {
					t.Errorf("Assert() panic message = %v, want %v", r, tt.message)
				}
			}()

			Assert(tt.condition, tt.message)
		})
	}
}

func TestAssertf(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		format    string
		args      []any
		wantPanic bool
	}{
		{
			name:      "true condition",
			condition: true,
			format:    "should not panic: %s",
			args:      []any{"test"},
			wantPanic: false,
		},
		{
			name:      "false condition",
			condition: false,
			format:    "should panic: %s",
			args:      []any{"test"},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Assertf() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
				if tt.wantPanic {
					expected := fmt.Sprintf(tt.format, tt.args...)
					if r != expected {
						t.Errorf("Assertf() panic message = %v, want %v", r, expected)
					}
				}
			}()

			Assertf(tt.condition, tt.format, tt.args...)
		})
	}
}

// Example test to demonstrate real-world usage
func TestMustExamples(t *testing.T) {
	t.Run("MustResult with strconv", func(t *testing.T) {
		result := MustResult(strconv.Atoi("42"))
		if result != 42 {
			t.Errorf("Expected 42, got %v", result)
		}
	})

	t.Run("Ignore with strconv", func(t *testing.T) {
		// Even with invalid input, Ignore should return the first value
		result := Ignore(strconv.Atoi("invalid"))
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}
	})
}

// TestIgnoreAtYourOwnRisk demonstrates proper and improper uses of the Ignore function.
// It shows cases where ignoring errors can lead to serious issues, and contrasts them
// with legitimate use cases where Ignore is appropriate.
func TestIgnoreAtYourOwnRisk(t *testing.T) {
	t.Run("critical errors that should not be ignored", func(t *testing.T) {
		// Anti-pattern: Ignoring database errors
		dbResult := Ignore(simulateDBOperation())
		t.Log("Warning: Ignoring DB errors can lead to data corruption, got:", dbResult)

		// Anti-pattern: Ignoring file system errors
		fileContent := Ignore(simulateFileRead())
		t.Log("Warning: Ignoring file errors can lead to security issues, got:", fileContent)
	})

	t.Run("acceptable use cases for Ignore", func(t *testing.T) {
		// Valid use: Converting compile-time constant strings
		port := Ignore(strconv.Atoi("8080"))
		if port != 8080 {
			t.Error("Failed to convert known-valid port number")
		}

		// Valid use: Operations with deterministic results
		str := "Hello, 世界"
		runeCount := Ignore(strconv.Atoi(fmt.Sprint(len([]rune(str)))))
		if runeCount != 9 {
			t.Errorf("Expected 9 runes in %q, got %d", str, runeCount)
		}
	})
}

// Helper functions that simulate operations where error handling is critical
func simulateDBOperation() (int, error) {
	return 0, fmt.Errorf("lost connection to database")
}

func simulateFileRead() (string, error) {
	return "", fmt.Errorf("file not found: sensitive_data.txt")
}
