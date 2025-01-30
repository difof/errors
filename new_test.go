package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		wantErr string
	}{
		{
			name:    "simple error message",
			msg:     "test error",
			wantErr: "test error",
		},
		{
			name:    "empty message",
			msg:     "",
			wantErr: "",
		},
		{
			name:    "message with special characters",
			msg:     "error: %$#@!",
			wantErr: "error: %$#@!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.msg)

			// Verify it's our custom error type
			customErr, ok := err.(*Error)
			assert.True(t, ok)

			// Verify error message
			assert.Equal(t, tt.wantErr, customErr.Message.Error())

			// Verify stack information is captured
			assert.NotEmpty(t, customErr.FuncPath)
			assert.NotEmpty(t, customErr.FilePath)
			assert.Greater(t, customErr.Line, 0)

			// Verify no inner error
			assert.Nil(t, customErr.Inner)
		})
	}
}

func TestNewf(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantErr string
	}{
		{
			name:    "simple format",
			format:  "error: %s",
			args:    []interface{}{"test"},
			wantErr: "error: test",
		},
		{
			name:    "multiple arguments",
			format:  "error: %s, code: %d",
			args:    []interface{}{"test", 404},
			wantErr: "error: test, code: 404",
		},
		{
			name:    "no arguments",
			format:  "plain error",
			args:    []interface{}{},
			wantErr: "plain error",
		},
		{
			name:    "empty format",
			format:  "",
			args:    []interface{}{},
			wantErr: "",
		},
		{
			name:    "format with special characters",
			format:  "error: %s !@#$%%",
			args:    []interface{}{"test"},
			wantErr: "error: test !@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Newf(tt.format, tt.args...)

			// Verify it's our custom error type
			customErr, ok := err.(*Error)
			assert.True(t, ok)

			// Verify error message
			assert.Equal(t, tt.wantErr, customErr.Message.Error())

			// Verify stack information is captured
			assert.NotEmpty(t, customErr.FuncPath)
			assert.NotEmpty(t, customErr.FilePath)
			assert.Greater(t, customErr.Line, 0)

			// Verify no inner error
			assert.Nil(t, customErr.Inner)
		})
	}
}

func TestNewSkip(t *testing.T) {
	tests := []struct {
		name    string
		skip    int
		msg     string
		wantErr string
	}{
		{
			name:    "skip 0",
			skip:    0,
			msg:     "test error",
			wantErr: "test error",
		},
		{
			name:    "skip 1",
			skip:    1,
			msg:     "test error",
			wantErr: "test error",
		},
		{
			name:    "skip 2",
			skip:    2,
			msg:     "test error",
			wantErr: "test error",
		},
		{
			name:    "empty message",
			skip:    0,
			msg:     "",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSkip(tt.skip, tt.msg)

			// Verify it's our custom error type
			customErr, ok := err.(*Error)
			assert.True(t, ok)

			// Verify error message
			assert.Equal(t, tt.wantErr, customErr.Message.Error())

			// Verify stack information is captured
			assert.NotEmpty(t, customErr.FuncPath)
			assert.NotEmpty(t, customErr.FilePath)
			assert.Greater(t, customErr.Line, 0)

			// Verify no inner error
			assert.Nil(t, customErr.Inner)
		})
	}
}

func TestNewSkipf(t *testing.T) {
	tests := []struct {
		name    string
		skip    int
		format  string
		args    []interface{}
		wantErr string
	}{
		{
			name:    "skip 0 with format",
			skip:    0,
			format:  "error: %s",
			args:    []interface{}{"test"},
			wantErr: "error: test",
		},
		{
			name:    "skip 1 with multiple args",
			skip:    1,
			format:  "error: %s, code: %d",
			args:    []interface{}{"test", 404},
			wantErr: "error: test, code: 404",
		},
		{
			name:    "skip 2 no args",
			skip:    2,
			format:  "plain error",
			args:    []interface{}{},
			wantErr: "plain error",
		},
		{
			name:    "empty format",
			skip:    0,
			format:  "",
			args:    []interface{}{},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSkipf(tt.skip, tt.format, tt.args...)

			// Verify it's our custom error type
			customErr, ok := err.(*Error)
			assert.True(t, ok)

			// Verify error message
			assert.Equal(t, tt.wantErr, customErr.Message.Error())

			// Verify stack information is captured
			assert.NotEmpty(t, customErr.FuncPath)
			assert.NotEmpty(t, customErr.FilePath)
			assert.Greater(t, customErr.Line, 0)

			// Verify no inner error
			assert.Nil(t, customErr.Inner)
		})
	}
}

// Helper function to verify error creation through different layers
func createNestedError(depth int) error {
	if depth == 0 {
		return New("base error")
	}
	innerErr := createNestedError(depth - 1)
	funcpath, filepath, line := getCallerPath(1)
	return NewError(funcpath, filepath, line, errors.New(fmt.Sprintf("level %d", depth)), innerErr)
}

func TestErrorNesting(t *testing.T) {
	err := createNestedError(3)
	customErr := err.(*Error)

	// Test error message extraction (gets the innermost error message)
	assert.Equal(t, "base error", ErrorMessageOf(err))

	// Verify each level has proper stack information and collect raw messages
	var messages []string
	customErr.Each(func(e error) bool {
		if ce, ok := e.(*Error); ok {
			messages = append(messages, ce.Message.Error())
			assert.NotEmpty(t, ce.FuncPath)
			assert.NotEmpty(t, ce.FilePath)
			assert.Greater(t, ce.Line, 0)
		}
		return true
	})

	// Verify we have all 4 levels (3 nested + base)
	assert.Equal(t, 4, len(messages))

	// Verify the message content
	assert.Equal(t, "level 3", messages[0])
	assert.Equal(t, "level 2", messages[1])
	assert.Equal(t, "level 1", messages[2])
	assert.Equal(t, "base error", messages[3])

	// Test error unwrapping
	current := err
	depth := 3
	for current != nil {
		if ce, ok := current.(*Error); ok {
			assert.NotEmpty(t, ce.FuncPath)
			assert.NotEmpty(t, ce.FilePath)
			assert.Greater(t, ce.Line, 0)
			if depth > 0 {
				assert.Equal(t, fmt.Sprintf("level %d", depth), ce.Message.Error())
			} else {
				assert.Equal(t, "base error", ce.Message.Error())
			}
			depth--
			current = ce.Unwrap()
		} else {
			current = nil
		}
	}
	assert.Equal(t, -1, depth, "Should have unwrapped through all levels")
}

// Test error creation in goroutines
func TestErrorInGoroutine(t *testing.T) {
	done := make(chan error)

	go func() {
		err := New("error from goroutine")
		done <- err
	}()

	err := <-done
	customErr, ok := err.(*Error)
	assert.True(t, ok)

	// Verify error details
	assert.Equal(t, "error from goroutine", customErr.Message.Error())
	assert.NotEmpty(t, customErr.FuncPath)
	assert.NotEmpty(t, customErr.FilePath)
	assert.Greater(t, customErr.Line, 0)
}

// Test error creation with various types of format arguments
func TestNewfVariousTypes(t *testing.T) {
	type customType struct {
		value string
	}

	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantErr string
	}{
		{
			name:    "with struct",
			format:  "error with struct: %v",
			args:    []interface{}{struct{ name string }{"test"}},
			wantErr: "error with struct: {test}",
		},
		{
			name:    "with custom type",
			format:  "error with custom type: %v",
			args:    []interface{}{&customType{"test"}},
			wantErr: fmt.Sprintf("error with custom type: &{test}"),
		},
		{
			name:    "with multiple types",
			format:  "string: %s, int: %d, float: %.2f, bool: %t",
			args:    []interface{}{"test", 42, 3.14, true},
			wantErr: "string: test, int: 42, float: 3.14, bool: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Newf(tt.format, tt.args...)
			assert.Equal(t, tt.wantErr, err.(*Error).Message.Error())
		})
	}
}
